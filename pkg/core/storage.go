package core

import (
	"fmt"
	"sync"
	"time"
)

func NewStorage(storageType string) (Storage, error) {
	switch storageType {
	case "memory":
		return NewMemoryStorage(), nil
	case "sqlite":
		// TODO: Implement SQLite storage
		return nil, fmt.Errorf("SQLite storage not implemented yet")
	case "postgres":
		// TODO: Implement PostgreSQL storage
		return nil, fmt.Errorf("PostgreSQL storage not implemented yet")
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

type MemoryStorage struct {
	mu             sync.RWMutex
	events         []Event
	detections     []DetectionResult
	eventIndex     map[string]int
	detectionIndex map[string]int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events:         make([]Event, 0),
		detections:     make([]DetectionResult, 0),
		eventIndex:     make(map[string]int),
		detectionIndex: make(map[string]int),
	}
}

func (m *MemoryStorage) SaveEvent(event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if event.ID == "" {
		event.ID = fmt.Sprintf("event_%d_%d", len(m.events), time.Now().UnixNano())
	}

	// Add to slice and index
	index := len(m.events)
	m.events = append(m.events, event)
	m.eventIndex[event.ID] = index

	return nil
}

func (m *MemoryStorage) SaveDetection(result DetectionResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if result.ID == "" {
		result.ID = fmt.Sprintf("detection_%d_%d", len(m.detections), time.Now().UnixNano())
	}

	// Add to slice and index
	index := len(m.detections)
	m.detections = append(m.detections, result)
	m.detectionIndex[result.ID] = index

	return nil
}

// GetEvents retrieves events with optional filters
func (m *MemoryStorage) GetEvents(filters map[string]interface{}) ([]Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If no filters, return all events
	if len(filters) == 0 {
		result := make([]Event, len(m.events))
		copy(result, m.events)
		return result, nil
	}

	var filtered []Event
	for _, event := range m.events {
		if matchesFilters(event, filters) {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

func (m *MemoryStorage) GetDetections(filters map[string]interface{}) ([]DetectionResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If no filters, return all detections
	if len(filters) == 0 {
		result := make([]DetectionResult, len(m.detections))
		copy(result, m.detections)
		return result, nil
	}

	// Apply filters
	var filtered []DetectionResult
	for _, detection := range m.detections {
		if matchesDetectionFilters(detection, filters) {
			filtered = append(filtered, detection)
		}
	}

	return filtered, nil
}

// Close closes the storage (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear data
	m.events = nil
	m.detections = nil
	m.eventIndex = nil
	m.detectionIndex = nil

	return nil
}

// Helper function to match event filters
func matchesFilters(event Event, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "source":
			if event.Source != value {
				return false
			}
		case "type":
			if event.Type != value {
				return false
			}
		case "domain":
			if event.Domain != value {
				return false
			}
		}
	}
	return true
}

func matchesDetectionFilters(detection DetectionResult, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "domain":
			if detection.Domain != value {
				return false
			}
		case "brand":
			if detection.Brand != value {
				return false
			}
		case "rule":
			if detection.Rule != value {
				return false
			}
		case "is_threat":
			if detection.IsThreat != value {
				return false
			}
		}
	}
	return true
}
