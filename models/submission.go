// models/submission.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// FormSubmission represents the main application form
type FormSubmission struct {
    ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Status          string            `json:"status" bson:"status"`
    SubmittedAt     time.Time         `json:"submitted_at" bson:"submitted_at"`
    UpdatedAt       time.Time         `json:"updated_at" bson:"updated_at"`
    Notes           []Note            `json:"notes" bson:"notes"`
    
    // Part A: Issuer Information
    IssuerInfo      IssuerInformation `json:"issuer_info" bson:"issuer_info"`
    
    // Part B: Securities Information
    SecuritiesInfo  SecuritiesInfo    `json:"securities_info" bson:"securities_info"`
    
    // Part C: Declaration
    Declaration     Declaration       `json:"declaration" bson:"declaration"`
    
    // Documents - using Document type from newDocument.go
    Documents       []Document        `json:"documents" bson:"documents"`
}

type IssuerInformation struct {
    CompanyName        string    `json:"company_name" bson:"company_name"`
    TradeName         string    `json:"trade_name" bson:"trade_name"`
    Constitution      string    `json:"constitution" bson:"constitution"`
    IncorporationDetails struct {
        Country       string    `json:"country" bson:"country"`
        Date         string    `json:"date" bson:"date"`
        RegNumber    string    `json:"reg_number" bson:"reg_number"`
        LEI          string    `json:"lei" bson:"lei"`
    } `json:"incorporation_details" bson:"incorporation_details"`
    Address struct {
        Registered   string    `json:"registered" bson:"registered"`
        Corporate    string    `json:"corporate" bson:"corporate"`
        Website      string    `json:"website" bson:"website"`
    } `json:"address" bson:"address"`
    Operations       string    `json:"operations" bson:"operations"`
    GroupDescription string    `json:"group_description" bson:"group_description"`
    IndustrySector   string    `json:"industry_sector" bson:"industry_sector"`
    ContactPerson struct {
        Name         string    `json:"name" bson:"name"`
        Position     string    `json:"position" bson:"position"`
        Phone        string    `json:"phone" bson:"phone"`
        Email        string    `json:"email" bson:"email"`
    } `json:"contact_person" bson:"contact_person"`
}

type SecuritiesInfo struct {
    ShareCapital struct {
        Authorized struct {
            Amount      float64 `json:"amount" bson:"amount"`
            NumShares   int     `json:"num_shares" bson:"num_shares"`
        } `json:"authorized" bson:"authorized"`
        PaidUp struct {
            Amount      float64 `json:"amount" bson:"amount"`
            NumShares   int     `json:"num_shares" bson:"num_shares"`
        } `json:"paid_up" bson:"paid_up"`
    } `json:"share_capital" bson:"share_capital"`
    Exchanges struct {
        Primary       string    `json:"primary" bson:"primary"`
        Secondary     string    `json:"secondary" bson:"secondary"`
    } `json:"exchanges" bson:"exchanges"`
    Security struct {
        Type         string    `json:"type" bson:"type"`
        ISIN         string    `json:"isin" bson:"isin"`
        ExpectedSize float64   `json:"expected_size" bson:"expected_size"`
        MarketCap    float64   `json:"market_cap" bson:"market_cap"`
        NumSecurities int      `json:"num_securities" bson:"num_securities"`
    } `json:"security" bson:"security"`
    Trading struct {
        ExpectedDate  time.Time `json:"expected_date" bson:"expected_date"`
        Symbol        string    `json:"symbol" bson:"symbol"`
    } `json:"trading" bson:"trading"`
    PricingMethod    string    `json:"pricing_method" bson:"pricing_method"`
}

type Declaration struct {
    IssuerName      string    `json:"issuer_name" bson:"issuer_name"`
    Signatory struct {
        Name        string    `json:"name" bson:"name"`
        Role        string    `json:"role" bson:"role"`
        Date        time.Time `json:"date" bson:"date"`
    } `json:"signatory" bson:"signatory"`
}

type Note struct {
    Content         string    `json:"content" bson:"content"`
    CreatedAt       time.Time `json:"created_at" bson:"created_at"`
    CreatedBy       string    `json:"created_by" bson:"created_by"`
}

func (f *FormSubmission) Validate() []string {
    var errors []string
    
    if f.IssuerInfo.CompanyName == "" {
        errors = append(errors, "Company name is required")
    }
    
    if f.IssuerInfo.Constitution == "" {
        errors = append(errors, "Constitution is required")
    }
    
    // Add more validation as needed
    
    return errors
}