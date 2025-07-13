// pkg/core/interfaces.go
package core

import (
	"context"
	"time"

	"github.com/openBPL/internal/config"
)

// Source represents a data source for monitoring
type Source interface {
	// Name returns the source name
	Name() string

	// Start begins monitoring and sends events to the channel
	Start(ctx context.Context, events chan<- Event) error

	// Stop gracefully stops the source
	Stop() error
}

// Enricher adds additional data to events
type Enricher interface {
	// Name returns the enricher name
	Name() string

	// Enrich adds data to the event
	Enrich(ctx context.Context, event *Event) error
}

// Detector analyzes events for threats
type Detector interface {
	// Name returns the detector name
	Name() string

	// Detect analyzes an event and returns detection results
	Detect(ctx context.Context, event *Event) ([]DetectionResult, error)
}

// Enforcer takes action on detected threats
type Enforcer interface {
	// Name returns the enforcer name
	Name() string

	// Enforce takes action on a detection result
	Enforce(ctx context.Context, result DetectionResult, dryRun bool) error
}

// Storage handles persistent data
type Storage interface {
	// SaveEvent stores an event
	SaveEvent(event Event) error

	// SaveDetection stores a detection result
	SaveDetection(result DetectionResult) error

	// GetEvents retrieves events with optional filters
	GetEvents(filters map[string]interface{}) ([]Event, error)

	// GetDetections retrieves detection results with optional filters
	GetDetections(filters map[string]interface{}) ([]DetectionResult, error)

	// Close closes the storage connection
	Close() error
}

// Event represents a monitoring event (e.g., new certificate)
type Event struct {
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Domain    string                 `json:"domain"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DetectionResult represents the result of threat detection
type DetectionResult struct {
	ID         string                 `json:"id"`
	EventID    string                 `json:"event_id"`
	Domain     string                 `json:"domain"`
	IsThreat   bool                   `json:"is_threat"`
	Confidence float64                `json:"confidence"`
	Brand      string                 `json:"brand"`
	Rule       string                 `json:"rule"`
	DetectedAt time.Time              `json:"detected_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Component stubs - these will be implemented in separate files

// CertstreamSource monitors Certificate Transparency logs
type CertstreamSource struct {
	URL      string
	Keywords []string
}

func (s *CertstreamSource) Name() string {
	return "certstream"
}

func (s *CertstreamSource) Stop() error {
	// TODO: Implement cleanup
	return nil
}

// HTMLEnricher fetches HTML content for domains
type HTMLEnricher struct {
	Timeout   string
	UserAgent string
}

func (e *HTMLEnricher) Name() string {
	return "html_content"
}

func (e *HTMLEnricher) Enrich(ctx context.Context, event *Event) error {
	// TODO: Implement HTML fetching
	return nil
}

// FaviconEnricher extracts favicon from domains
type FaviconEnricher struct {
	Timeout string
}

func (e *FaviconEnricher) Name() string {
	return "favicon"
}

func (e *FaviconEnricher) Enrich(ctx context.Context, event *Event) error {
	// TODO: Implement favicon extraction
	return nil
}

// FaviconSimilarityDetector compares favicons using pHash
type FaviconSimilarityDetector struct {
	Threshold         float64
	ReferenceFavicons map[string]string
}

func (d *FaviconSimilarityDetector) Name() string {
	return "favicon_similarity"
}

func (d *FaviconSimilarityDetector) Detect(ctx context.Context, event *Event) ([]DetectionResult, error) {
	// TODO: Implement favicon similarity detection
	return nil, nil
}

// LoggerEnforcer logs detection results
type LoggerEnforcer struct{}

func (e *LoggerEnforcer) Name() string {
	return "logger"
}

func (e *LoggerEnforcer) Enforce(ctx context.Context, result DetectionResult, dryRun bool) error {
	// TODO: Implement logging
	return nil
}

// EmailEnforcer sends abuse emails
type EmailEnforcer struct {
	SMTP config.SMTPConfig
	From string
}

func (e *EmailEnforcer) Name() string {
	return "email_abuse"
}

func (e *EmailEnforcer) Enforce(ctx context.Context, result DetectionResult, dryRun bool) error {
	// TODO: Implement email sending
	return nil
}
