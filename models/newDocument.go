package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UploadResult represents the result of a file upload
type UploadResult struct {
    Location  string    `json:"location" bson:"location"`
    Filename  string    `json:"filename" bson:"filename"`
    Key       string    `json:"key" bson:"key"`
    Size      int64     `json:"size" bson:"size"`
    MimeType  string    `json:"mime_type" bson:"mime_type"`
}

// Document represents an uploaded file
type Document struct {
    ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Type            string    `json:"type" bson:"type"`
    Name            string    `json:"name" bson:"name"`
    Description     string    `json:"description" bson:"description"`
    
    // S3 specific fields
    Key             string    `json:"key" bson:"key"`
    Location        string    `json:"location" bson:"location"`
    Path            string    `json:"path" bson:"path"`
    Size           int64     `json:"size" bson:"size"`
    ContentType    string    `json:"content_type" bson:"content_type"`
    
    // Metadata
    UploadedAt      time.Time `json:"uploaded_at" bson:"uploaded_at"`
    UploadedBy      string    `json:"uploaded_by" bson:"uploaded_by"`
    Status          string    `json:"status" bson:"status"`
    Version         int       `json:"version" bson:"version"`
}

// NewDocument creates a new document from upload result
func NewDocument(uploadResult *UploadResult, docType string) Document {
    return Document{
        ID:          primitive.NewObjectID(),
        Type:        docType,
        Name:        uploadResult.Filename,
        Key:         uploadResult.Key,
        Location:    uploadResult.Location,
        Path:        uploadResult.Key,
        Size:        uploadResult.Size,
        ContentType: uploadResult.MimeType,
        UploadedAt:  time.Now(),
        Status:      "pending",
        Version:     1,
    }
}