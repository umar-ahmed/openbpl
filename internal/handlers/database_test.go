package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"openbpl/pkg/models"
)

// Mock repositories
type mockUserRepository struct {
	users        []models.User
	getByIDError error
	getAllError  error
}

func (m *mockUserRepository) GetAll() ([]models.User, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	return m.users, nil
}

func (m *mockUserRepository) GetByID(id int) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	for _, user := range m.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, nil
}

type mockThreatRepository struct {
	threats      []models.Threat
	getByIDError error
	getAllError  error
}

func (m *mockThreatRepository) GetAll() ([]models.Threat, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	return m.threats, nil
}

func (m *mockThreatRepository) GetByID(id int) (*models.Threat, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	for _, threat := range m.threats {
		if threat.ID == id {
			return &threat, nil
		}
	}
	return nil, nil
}

func setupTestHandlers() *DatabaseHandlers {
	return &DatabaseHandlers{
		UserRepo: &mockUserRepository{
			users: []models.User{
				{ID: 1, Email: "user1@test.com", Name: "User One", CreatedAt: time.Now()},
				{ID: 2, Email: "user2@test.com", Name: "User Two", CreatedAt: time.Now()},
			},
		},
		ThreatRepo: &mockThreatRepository{
			threats: []models.Threat{
				{ID: 1, URL: "https://fake1.com", Status: "new", CreatedAt: time.Now()},
				{ID: 2, URL: "https://fake2.com", Status: "investigating", CreatedAt: time.Now()},
			},
		},
	}
}

func TestListUsers(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	handlers.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "ok" {
		t.Errorf("Expected status ok, got %s", resp.Status)
	}
}

func TestListUsers_Error(t *testing.T) {
	handlers := setupTestHandlers()
	handlers.UserRepo = &mockUserRepository{getAllError: errors.New("db error")}

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	handlers.ListUsers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestGetUser(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()

	handlers.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "ok" {
		t.Errorf("Expected status ok, got %s", resp.Status)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/users/999", nil)
	w := httptest.NewRecorder()

	handlers.GetUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/users/abc", nil)
	w := httptest.NewRecorder()

	handlers.GetUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestGetUser_Error(t *testing.T) {
	handlers := setupTestHandlers()
	handlers.UserRepo = &mockUserRepository{getByIDError: errors.New("db error")}

	req := httptest.NewRequest("GET", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()

	handlers.GetUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestListThreats(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/threats", nil)
	w := httptest.NewRecorder()

	handlers.ListThreats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestGetThreat(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/threats/1", nil)
	w := httptest.NewRecorder()

	handlers.GetThreat(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestGetThreat_NotFound(t *testing.T) {
	handlers := setupTestHandlers()
	req := httptest.NewRequest("GET", "/api/v1/threats/999", nil)
	w := httptest.NewRecorder()

	handlers.GetThreat(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}
