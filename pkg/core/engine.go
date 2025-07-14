// pkg/core/engine.go
package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/openBPL/internal/config"
)

// Engine is the main monitoring engine
type Engine struct {
	cfg       *config.Config
	sources   []Source
	enrichers []Enricher
	detectors []Detector
	enforcers []Enforcer
	storage   Storage
	stats     *Statistics
}

// Statistics tracks monitoring statistics
type Statistics struct {
	mu             sync.RWMutex
	CertsProcessed int64     `json:"certs_processed"`
	ThreatsFound   int64     `json:"threats_found"`
	ActionsLive    int64     `json:"actions_live"`
	ActionsDryRun  int64     `json:"actions_dry_run"`
	StartTime      time.Time `json:"start_time"`
}

// NewEngine creates a new monitoring engine
func NewEngine(cfg *config.Config) (*Engine, error) {
	log.Printf("🔧 Initializing OpenBPL engine...")

	// Initialize storage
	storage, err := NewStorage(cfg.Storage.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	log.Printf("💾 Storage initialized: %s", cfg.Storage.Type)

	// Initialize sources
	sources, err := initializeSources(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sources: %w", err)
	}
	log.Printf("🔌 Sources initialized: %d", len(sources))

	// Initialize enrichers
	enrichers, err := initializeEnrichers(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize enrichers: %w", err)
	}
	log.Printf("🔍 Enrichers initialized: %d", len(enrichers))

	// Initialize detectors
	detectors, err := initializeDetectors(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detectors: %w", err)
	}
	log.Printf("🚨 Detectors initialized: %d", len(detectors))

	// Initialize enforcers
	enforcers, err := initializeEnforcers(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize enforcers: %w", err)
	}
	log.Printf("⚡ Enforcers initialized: %d", len(enforcers))

	return &Engine{
		cfg:       cfg,
		sources:   sources,
		enrichers: enrichers,
		detectors: detectors,
		enforcers: enforcers,
		storage:   storage,
		stats:     &Statistics{StartTime: time.Now()},
	}, nil
}

// Run starts the monitoring engine
func (e *Engine) Run(ctx context.Context) error {
	log.Printf("🚀 Starting OpenBPL monitoring engine...")
	log.Printf("📊 Mode: %s", func() string {
		if e.cfg.DryRun {
			return "DRY-RUN"
		}
		return "LIVE"
	}())

	// Create event channel for sources to send events
	events := make(chan Event, 100)

	// Start statistics reporter
	go e.reportStats(ctx)

	// Start all sources
	var wg sync.WaitGroup
	for _, source := range e.sources {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			log.Printf("🎯 Starting source: %s", s.Name())
			if err := s.Start(ctx, events); err != nil {
				log.Printf("❌ Source %s failed: %v", s.Name(), err)
			}
		}(source)
	}

	// Start event processor
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.processEvents(ctx, events)
	}()

	// Wait for context cancellation
	<-ctx.Done()
	log.Printf("🛑 Monitoring engine stopping...")

	// Close event channel
	close(events)

	// Wait for all goroutines to finish
	wg.Wait()

	log.Printf("🛑 Monitoring engine stopped")
	return nil
}

// processEvents handles incoming events from sources
func (e *Engine) processEvents(ctx context.Context, events <-chan Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return // Channel closed
			}

			// Process the event
			if err := e.processEvent(ctx, event); err != nil {
				log.Printf("❌ Failed to process event %s: %v", event.ID, err)
			}
		}
	}
}

// processEvent processes a single event through the pipeline
func (e *Engine) processEvent(ctx context.Context, event Event) error {
	// Update statistics
	e.stats.mu.Lock()
	e.stats.CertsProcessed++
	e.stats.mu.Unlock()

	// Save event to storage
	if err := e.storage.SaveEvent(event); err != nil {
		log.Printf("⚠️ Failed to save event: %v", err)
	}

	// Run enrichment pipeline
	for _, enricher := range e.enrichers {
		if err := enricher.Enrich(ctx, &event); err != nil {
			log.Printf("⚠️ Enricher %s failed for %s: %v", enricher.Name(), event.Domain, err)
			// Continue with other enrichers
		}
	}

	// Run detection pipeline
	var allResults []DetectionResult
	for _, detector := range e.detectors {
		results, err := detector.Detect(ctx, &event)
		if err != nil {
			log.Printf("⚠️ Detector %s failed for %s: %v", detector.Name(), event.Domain, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	// Process detection results
	for _, result := range allResults {
		if err := e.processDetectionResult(ctx, result); err != nil {
			log.Printf("❌ Failed to process detection result: %v", err)
		}
	}

	return nil
}

// processDetectionResult handles a detection result
func (e *Engine) processDetectionResult(ctx context.Context, result DetectionResult) error {
	// Save detection result
	if err := e.storage.SaveDetection(result); err != nil {
		log.Printf("⚠️ Failed to save detection result: %v", err)
	}

	// Update statistics
	e.stats.mu.Lock()
	if result.IsThreat {
		e.stats.ThreatsFound++
	}
	e.stats.mu.Unlock()

	// Only proceed with enforcement if it's a threat
	if !result.IsThreat {
		return nil
	}

	log.Printf("🚨 THREAT DETECTED: %s (confidence: %.2f, rule: %s)",
		result.Domain, result.Confidence, result.Rule)

	// Run enforcement pipeline
	for _, enforcer := range e.enforcers {
		if err := enforcer.Enforce(ctx, result, e.cfg.DryRun); err != nil {
			log.Printf("⚠️ Enforcer %s failed for %s: %v", enforcer.Name(), result.Domain, err)
			continue
		}

		// Update enforcement statistics
		e.stats.mu.Lock()
		if e.cfg.DryRun {
			e.stats.ActionsDryRun++
		} else {
			e.stats.ActionsLive++
		}
		e.stats.mu.Unlock()
	}

	return nil
}

// reportStats periodically reports engine statistics
func (e *Engine) reportStats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.stats.mu.RLock()
			uptime := time.Since(e.stats.StartTime)
			log.Printf("📊 Stats - Uptime: %v, Certs: %d, Threats: %d, Actions: %d (live) + %d (dry-run)",
				uptime.Round(time.Second),
				e.stats.CertsProcessed,
				e.stats.ThreatsFound,
				e.stats.ActionsLive,
				e.stats.ActionsDryRun,
			)
			e.stats.mu.RUnlock()
		}
	}
}

// Helper functions to initialize components
func initializeSources(cfg *config.Config) ([]Source, error) {
	var sources []Source

	// Initialize certstream source if enabled
	if cfg.Monitoring.Sources.Certstream.Enabled {
		source := &CertstreamSource{
			URL:      cfg.Monitoring.Sources.Certstream.URL,
			Keywords: cfg.Monitoring.Sources.Certstream.Keywords,
		}
		sources = append(sources, source)
	}

	return sources, nil
}

func initializeEnrichers(cfg *config.Config) ([]Enricher, error) {
	var enrichers []Enricher

	// Initialize HTML content enricher if enabled
	if cfg.Enrichment.HTMLContent.Enabled {
		enricher := &HTMLEnricher{
			Timeout:   cfg.Enrichment.HTMLContent.Timeout,
			UserAgent: cfg.Enrichment.HTMLContent.UserAgent,
		}
		enrichers = append(enrichers, enricher)
	}

	// Initialize favicon enricher if enabled
	if cfg.Enrichment.Favicon.Enabled {
		enricher := &FaviconEnricher{
			Timeout: cfg.Enrichment.Favicon.Timeout,
		}
		enrichers = append(enrichers, enricher)
	}

	return enrichers, nil
}

func initializeDetectors(cfg *config.Config) ([]Detector, error) {
	var detectors []Detector

	// Initialize favicon similarity detector if enabled
	if cfg.Rules.FaviconSimilarity.Enabled {
		detector := &FaviconSimilarityDetector{
			Threshold:         cfg.Rules.FaviconSimilarity.Threshold,
			ReferenceFavicons: cfg.Rules.FaviconSimilarity.ReferenceFavicons,
		}
		detectors = append(detectors, detector)
	}

	return detectors, nil
}

func initializeEnforcers(cfg *config.Config) ([]Enforcer, error) {
	var enforcers []Enforcer

	// Always add logger enforcer if enabled
	if cfg.Enforcement.Logger.Enabled {
		enforcer := &LoggerEnforcer{}
		enforcers = append(enforcers, enforcer)
	}

	// Add email enforcer if enabled
	if cfg.Enforcement.EmailAbuse.Enabled {
		enforcer := &EmailEnforcer{
			SMTP: cfg.Enforcement.EmailAbuse.SMTP,
			From: cfg.Enforcement.EmailAbuse.From,
		}
		enforcers = append(enforcers, enforcer)
	}

	return enforcers, nil
}
