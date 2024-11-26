package services

import (
    "context"
    "errors"
    "time"
    "math"
    "mime/multipart"
    "fmt"
    "onboarding-app/models"
    "onboarding-app/config"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
    ErrDuplicateSubmission = errors.New("duplicate submission")
    ErrSubmissionNotFound = errors.New("submission not found")
)

// SubmissionFilter defines the filtering options for submissions
type SubmissionFilter struct {
    Page      int
    Limit     int
    SortField string
    SortOrder int
    Status    string
    Search    string
}

// SubmissionResult holds paginated submission results
type SubmissionResult struct {
    Submissions []models.FormSubmission `json:"submissions"`
    Total      int64                   `json:"total"`
    Page       int                     `json:"page"`
    TotalPages int                     `json:"total_pages"`
}

// SubmissionStats represents submission statistics
type SubmissionStats struct {
    Total         int64             `json:"total"`
    StatusCounts  map[string]int64  `json:"status_counts"`
    RecentCount   int64             `json:"recent_count"` // submissions in last 24h
    PendingCount  int64             `json:"pending_count"`
}

// SubmissionService handles business logic for form submissions
type SubmissionService struct {
    collection  *mongo.Collection
    fileService *S3FileService
}

// NewSubmissionService creates a new submission service instance
func NewSubmissionService(s3Config S3Config) (*SubmissionService, error) {
    fileService, err := NewS3FileService(s3Config)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize file service: %v", err)
    }

    return &SubmissionService{
        collection:  config.Client.Database("onboarding").Collection("submissions"),
        fileService: fileService,
    }, nil
}

func (s *SubmissionService) GetStats(ctx context.Context) (*SubmissionStats, error) {
    // Get total count
    total, err := s.collection.CountDocuments(ctx, bson.M{})
    if err != nil {
        return nil, err
    }

    // Get counts by status
    pipeline := []bson.M{
        {
            "$group": bson.M{
                "_id": "$status",
                "count": bson.M{"$sum": 1},
            },
        },
    }

    cursor, err := s.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }

    // Get recent submissions (last 24h)
    yesterday := time.Now().Add(-24 * time.Hour)
    recentCount, err := s.collection.CountDocuments(ctx, bson.M{
        "submitted_at": bson.M{"$gte": yesterday},
    })
    if err != nil {
        return nil, err
    }

    // Get pending submissions
    pendingCount, err := s.collection.CountDocuments(ctx, bson.M{
        "status": "pending",
    })
    if err != nil {
        return nil, err
    }

    // Format results
    stats := &SubmissionStats{
        Total:        total,
        StatusCounts: make(map[string]int64),
        RecentCount:  recentCount,
        PendingCount: pendingCount,
    }

    for _, result := range results {
        status := result["_id"].(string)
        count := result["count"].(int32)
        stats.StatusCounts[status] = int64(count)
    }

    return stats, nil
}

func (s *SubmissionService) GetSubmissions(ctx context.Context, filter SubmissionFilter) (*SubmissionResult, error) {
    // Calculate skip value for pagination
    skip := (filter.Page - 1) * filter.Limit

    // Build query filter
    queryFilter := bson.M{}
    if filter.Status != "" {
        queryFilter["status"] = filter.Status
    }

    // Add search functionality
    if filter.Search != "" {
        queryFilter["$or"] = []bson.M{
            {"issuer_info.company_name": bson.M{"$regex": filter.Search, "$options": "i"}},
            {"issuer_info.trade_name": bson.M{"$regex": filter.Search, "$options": "i"}},
            {"issuer_info.incorporation_details.lei": bson.M{"$regex": filter.Search, "$options": "i"}},
        }
    }

    // Set up options
    findOptions := options.Find().
        SetSort(bson.D{{Key: filter.SortField, Value: filter.SortOrder}}).
        SetSkip(int64(skip)).
        SetLimit(int64(filter.Limit))

    // Get total count
    total, err := s.collection.CountDocuments(ctx, queryFilter)
    if err != nil {
        return nil, err
    }

    // Calculate total pages
    totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

    // Execute query
    cursor, err := s.collection.Find(ctx, queryFilter, findOptions)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    // Parse results
    var submissions []models.FormSubmission
    if err := cursor.All(ctx, &submissions); err != nil {
        return nil, err
    }

    return &SubmissionResult{
        Submissions: submissions,
        Total:      total,
        Page:       filter.Page,
        TotalPages: totalPages,
    }, nil
}

func (s *SubmissionService) GetSubmissionByID(ctx context.Context, id primitive.ObjectID) (*models.FormSubmission, error) {
    var submission models.FormSubmission
    err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&submission)
    if err == mongo.ErrNoDocuments {
        return nil, ErrSubmissionNotFound
    }
    if err != nil {
        return nil, err
    }
    return &submission, nil
}

func (s *SubmissionService) Create(submission *models.FormSubmission) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Check for duplicates based on company name
    if submission.IssuerInfo.CompanyName != "" {
        count, err := s.collection.CountDocuments(ctx, bson.M{
            "issuer_info.company_name": submission.IssuerInfo.CompanyName,
        })
        if err != nil {
            return err
        }
        if count > 0 {
            return ErrDuplicateSubmission
        }
    }
    
    // Set timestamps
    submission.SubmittedAt = time.Now()
    submission.UpdatedAt = time.Now()
    
    // Insert submission
    _, err := s.collection.InsertOne(ctx, submission)
    return err
}

// Update updates a submission with the provided fields
func (s *SubmissionService) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
    result, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
    if err != nil {
        return err
    }

    if result.MatchedCount == 0 {
        return ErrSubmissionNotFound
    }

    return nil
}

func (s *SubmissionService) AddDocument(ctx context.Context, submissionID primitive.ObjectID, file *multipart.FileHeader, docType, uploadedBy string) (*models.Document, error) {
    // Upload file to S3
    uploadResult, err := s.fileService.UploadFile(ctx, file)
    if err != nil {
        return nil, err
    }

    // Create document record
    doc := models.NewDocument(uploadResult, docType)
    doc.UploadedBy = uploadedBy

    // Update submission with new document
    update := bson.M{
        "$push": bson.M{"documents": doc},
        "$set":  bson.M{"updated_at": time.Now()},
    }

    err = s.Update(ctx, submissionID, update)
    if err != nil {
        // If mongodb update fails, try to delete the uploaded file
        deleteCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        _ = s.fileService.DeleteFile(deleteCtx, uploadResult.Key)
        return nil, err
    }

    return &doc, nil
}

// UpdateDocumentStatus updates the status of a document
func (s *SubmissionService) UpdateDocumentStatus(ctx context.Context, submissionID primitive.ObjectID, documentID primitive.ObjectID, status string) error {
    update := bson.M{
        "$set": bson.M{
            "documents.$[doc].status": status,
            "updated_at": time.Now(),
        },
    }

    arrayFilters := options.ArrayFilters{
        Filters: []interface{}{
            bson.M{"doc._id": documentID},
        },
    }

    opts := options.Update().SetArrayFilters(arrayFilters)

    result, err := s.collection.UpdateOne(ctx, bson.M{"_id": submissionID}, update, opts)
    if err != nil {
        return err
    }

    if result.MatchedCount == 0 {
        return ErrSubmissionNotFound
    }

    return nil
}

// DeleteDocument removes a document from a submission and S3
func (s *SubmissionService) DeleteDocument(ctx context.Context, submissionID primitive.ObjectID, documentID primitive.ObjectID) error {
    // First get the document to know the S3 key
    submission, err := s.GetSubmissionByID(ctx, submissionID)
    if err != nil {
        return err
    }

    var targetDoc *models.Document
    for _, doc := range submission.Documents {
        if doc.ID == documentID {
            targetDoc = &doc
            break
        }
    }

    if targetDoc == nil {
        return fmt.Errorf("document not found")
    }

    // Delete from S3
    err = s.fileService.DeleteFile(ctx, targetDoc.Key)
    if err != nil {
        return err
    }

    // Remove from MongoDB
    update := bson.M{
        "$pull": bson.M{
            "documents": bson.M{"_id": documentID},
        },
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    result, err := s.collection.UpdateOne(ctx, bson.M{"_id": submissionID}, update)
    if err != nil {
        return err
    }

    if result.MatchedCount == 0 {
        return ErrSubmissionNotFound
    }

    return nil
}