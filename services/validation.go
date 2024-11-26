package services

import (
    "regexp"
    "strings"
    "time"
    "onboarding-app/models"
)

type ValidationService struct{}

func NewValidationService() *ValidationService {
    return &ValidationService{}
}

func (v *ValidationService) ValidateSubmission(submission *models.FormSubmission) map[string]string {
    errors := make(map[string]string)

    // Company Name validation
    if strings.TrimSpace(submission.IssuerInfo.CompanyName) == "" {
        errors["company_name"] = "Company name is required"
    } else if len(submission.IssuerInfo.CompanyName) > 100 {
        errors["company_name"] = "Company name must not exceed 100 characters"
    }

    // Constitution validation
    validConstitutions := []string{"LLC", "Corporation", "Partnership", "Sole Proprietorship"}
    if !contains(validConstitutions, submission.IssuerInfo.Constitution) {
        errors["constitution"] = "Invalid constitution type"
    }

    // Incorporation Details validation
    if submission.IssuerInfo.IncorporationDetails.Country == "" {
        errors["incorporation_country"] = "Country of incorporation is required"
    }

    if submission.IssuerInfo.IncorporationDetails.Date != "" {
        if _, err := time.Parse("2006-01-02", submission.IssuerInfo.IncorporationDetails.Date); err != nil {
            errors["incorporation_date"] = "Invalid date format. Use YYYY-MM-DD"
        }
    }

    if submission.IssuerInfo.IncorporationDetails.RegNumber != "" {
        if len(submission.IssuerInfo.IncorporationDetails.RegNumber) < 5 {
            errors["reg_number"] = "Registration number must be at least 5 characters"
        }
    }

    // LEI validation (if provided)
    if submission.IssuerInfo.IncorporationDetails.LEI != "" {
        leiRegex := regexp.MustCompile(`^[0-9A-Z]{20}$`)
        if !leiRegex.MatchString(submission.IssuerInfo.IncorporationDetails.LEI) {
            errors["lei"] = "Invalid LEI format"
        }
    }

    // Address validation
    if strings.TrimSpace(submission.IssuerInfo.Address.Registered) == "" {
        errors["registered_address"] = "Registered address is required"
    }

    if submission.IssuerInfo.Address.Website != "" {
        websiteRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
        if !websiteRegex.MatchString(submission.IssuerInfo.Address.Website) {
            errors["website"] = "Invalid website URL format"
        }
    }

    return errors
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// Optional: Add more specific validation functions
func (v *ValidationService) ValidateNote(note *models.Note) map[string]string {
    errors := make(map[string]string)

    if strings.TrimSpace(note.Content) == "" {
        errors["content"] = "Note content is required"
    }

    if note.CreatedBy == "" {
        errors["created_by"] = "Note creator is required"
    }

    return errors
}