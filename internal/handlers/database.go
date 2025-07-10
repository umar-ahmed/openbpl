package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"openbpl/pkg/models"
)

// DatabaseHandlers holds repository dependencies
type UserRepositoryInterface interface {
	GetAll() ([]models.User, error)
	GetByID(id int) (*models.User, error)
}

type ThreatRepositoryInterface interface {
	GetAll() ([]models.Threat, error)
	GetByID(id int) (*models.Threat, error)
}

// DatabaseHandlers holds repository dependencies
type DatabaseHandlers struct {
	UserRepo   UserRepositoryInterface   // instead of *models.UserRepository
	ThreatRepo ThreatRepositoryInterface // instead of *models.ThreatRepository
}

// NewDatabaseHandlers creates new database handlers
func NewDatabaseHandlers(userRepo UserRepositoryInterface, threatRepo ThreatRepositoryInterface) *DatabaseHandlers {
	return &DatabaseHandlers{
		UserRepo:   userRepo,
		ThreatRepo: threatRepo,
	}
}

// ListUsers handles GET /api/v1/users
func (h *DatabaseHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserRepo.GetAll()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	response := Response{
		Status:  "ok",
		Message: "Users retrieved successfully",
		Data: map[string]interface{}{
			"users": users,
			"count": len(users),
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GetUser handles GET /api/v1/users/{id}
func (h *DatabaseHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		writeErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	idStr := pathParts[4] // /api/v1/users/{id}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	if user == nil {
		writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	response := Response{
		Status:  "ok",
		Message: "User retrieved successfully",
		Data:    user,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// ListThreats handles GET /api/v1/threats
func (h *DatabaseHandlers) ListThreats(w http.ResponseWriter, r *http.Request) {
	threats, err := h.ThreatRepo.GetAll()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve threats")
		return
	}

	response := Response{
		Status:  "ok",
		Message: "Threats retrieved successfully",
		Data: map[string]interface{}{
			"threats": threats,
			"count":   len(threats),
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GetThreat handles GET /api/v1/threats/{id}
func (h *DatabaseHandlers) GetThreat(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		writeErrorResponse(w, http.StatusBadRequest, "Threat ID is required")
		return
	}

	idStr := pathParts[4] // /api/v1/threats/{id}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid threat ID")
		return
	}

	threat, err := h.ThreatRepo.GetByID(id)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve threat")
		return
	}

	if threat == nil {
		writeErrorResponse(w, http.StatusNotFound, "Threat not found")
		return
	}

	response := Response{
		Status:  "ok",
		Message: "Threat retrieved successfully",
		Data:    threat,
	}

	writeJSONResponse(w, http.StatusOK, response)
}
