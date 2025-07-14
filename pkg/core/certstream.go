package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// CertstreamEntry represents a certificate transparency log entry
type CertstreamEntry struct {
	MessageType string `json:"message_type"`
	Data        struct {
		UpdateType string `json:"update_type"`
		LeafCert   struct {
			Subject struct {
				CN string `json:"CN"`
			} `json:"subject"`
			Extensions struct {
				SubjectAltName string `json:"subjectAltName"`
			} `json:"extensions"`
		} `json:"leaf_cert"`
	} `json:"data"`
}

// Start begins monitoring certstream and sends events to the channel
func (s *CertstreamSource) Start(ctx context.Context, events chan<- Event) error {
	log.Printf("🔌 Connecting to certstream: %s", s.URL)

	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 Certstream source stopped")
			return nil
		default:
			if err := s.connect(ctx, events); err != nil {
				log.Printf("❌ Certstream connection failed: %v", err)
				log.Printf("🔄 Reconnecting in 5 seconds...")

				// Wait before reconnecting
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(5 * time.Second):
					continue
				}
			}
		}
	}
}

func (s *CertstreamSource) connect(ctx context.Context, events chan<- Event) error {
	// Connect to certstream WebSocket
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(s.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to certstream: %w", err)
	}
	defer conn.Close()

	log.Printf("✅ Connected to certstream")
	log.Printf("🔍 Monitoring keywords: %v", s.Keywords)

	// Set read deadline for periodic checks
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Read message from certstream
			_, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("failed to read from certstream: %w", err)
			}

			// Reset read deadline
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			// Process the certificate entry
			if err := s.processCertEntry(message, events); err != nil {
				log.Printf("⚠️ Failed to process cert entry: %v", err)
				// Continue processing other entries
			}
		}
	}
}

func (s *CertstreamSource) processCertEntry(message []byte, events chan<- Event) error {
	var entry CertstreamEntry
	if err := json.Unmarshal(message, &entry); err != nil {
		return err
	}

	// Only process certificate updates
	if entry.MessageType != "certificate_update" {
		return nil
	}

	// Extract domains from the certificate
	domains := s.extractDomains(&entry)

	// Check each domain against our keywords
	for _, domain := range domains {
		if s.shouldProcess(domain) {
			// Create event for this domain
			event := Event{
				ID:        fmt.Sprintf("cert_%d", time.Now().UnixNano()),
				Source:    s.Name(),
				Type:      "certificate_update",
				Domain:    domain,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"cn":          entry.Data.LeafCert.Subject.CN,
					"sans":        entry.Data.LeafCert.Extensions.SubjectAltName,
					"update_type": entry.Data.UpdateType,
				},
				Metadata: map[string]interface{}{
					"matched_keywords": s.getMatchedKeywords(domain),
				},
			}

			// Send event to processing pipeline
			select {
			case events <- event:
				log.Printf("🆕 New certificate: %s (matched: %v)", domain, event.Metadata["matched_keywords"])
			case <-time.After(1 * time.Second):
				log.Printf("⚠️ Event channel full, dropping certificate: %s", domain)
			}
		}
	}

	return nil
}

func (s *CertstreamSource) extractDomains(entry *CertstreamEntry) []string {
	var domains []string

	// Add CN if present
	if cn := entry.Data.LeafCert.Subject.CN; cn != "" {
		domains = append(domains, strings.ToLower(cn))
	}

	// Add SANs if present
	if sans := entry.Data.LeafCert.Extensions.SubjectAltName; sans != "" {
		// Parse SAN field (format: "DNS:example.com, DNS:www.example.com")
		parts := strings.Split(sans, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "DNS:") {
				domain := strings.TrimPrefix(part, "DNS:")
				domains = append(domains, strings.ToLower(domain))
			}
		}
	}

	return domains
}

func (s *CertstreamSource) shouldProcess(domain string) bool {
	// Skip wildcard domains
	if strings.HasPrefix(domain, "*.") {
		return false
	}

	// Skip very short domains (likely not interesting)
	if len(domain) < 4 {
		return false
	}

	// Check against keywords
	for _, keyword := range s.Keywords {
		if strings.Contains(domain, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func (s *CertstreamSource) getMatchedKeywords(domain string) []string {
	var matched []string
	domain = strings.ToLower(domain)

	for _, keyword := range s.Keywords {
		if strings.Contains(domain, strings.ToLower(keyword)) {
			matched = append(matched, keyword)
		}
	}

	return matched
}
