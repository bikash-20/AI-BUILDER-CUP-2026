package domain

import (
	"encoding/json"
	"time"
)

// DocumentType enumerates the kinds of documents FinGuard accepts.
type DocumentType string

const (
	DocumentKYC      DocumentType = "KYC"
	DocumentInvoice  DocumentType = "INVOICE"
	DocumentClaim    DocumentType = "CLAIM"
	DocumentLoanApp  DocumentType = "LOAN_APPLICATION"
	DocumentID       DocumentType = "IDENTITY"
	DocumentBankStmt DocumentType = "BANK_STATEMENT"
)

// DocumentStatus is the lifecycle of a document.
type DocumentStatus string

const (
	DocStatusProcessing  DocumentStatus = "PROCESSING"
	DocStatusExtracted   DocumentStatus = "EXTRACTED"
	DocStatusValidated   DocumentStatus = "VALIDATED"
	DocStatusNeedsReview DocumentStatus = "NEEDS_REVIEW"
	DocStatusRejected    DocumentStatus = "REJECTED"
	DocStatusApproved    DocumentStatus = "APPROVED"
)

// ExtractedEntity is a single piece of structured information extracted
// from a document by Document AI.
type ExtractedEntity struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Page       int     `json:"page,omitempty"`
}

// Document is a document uploaded to FinGuard.
type Document struct {
	DocID         string           `json:"doc_id"`
	CustomerID    string           `json:"customer_id"`
	DocType       DocumentType     `json:"doc_type"`
	GCSURI        string           `json:"gcs_uri"`
	ExtractedJSON json.RawMessage  `json:"extracted_json,omitempty"`
	Entities      []ExtractedEntity `json:"entities,omitempty"`
	Confidence    float64          `json:"confidence"`
	Status        DocumentStatus   `json:"status"`
	UploadedAt    time.Time        `json:"uploaded_at"`
	ProcessedAt   *time.Time       `json:"processed_at,omitempty"`
}

// UploadDocumentRequest is the request body for POST /v1/documents/upload.
type UploadDocumentRequest struct {
	DocType  DocumentType `json:"doc_type" validate:"required"`
	FileURL  string       `json:"file_url" validate:"required,startswith=gs://"`
	CustomerID string     `json:"customer_id" validate:"required"`
}

// UploadDocumentResponse is returned (202) after acceptance.
type UploadDocumentResponse struct {
	DocID  string         `json:"doc_id"`
	Status DocumentStatus `json:"status"`
}
