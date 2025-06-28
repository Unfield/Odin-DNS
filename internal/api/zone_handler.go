package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// CreateZoneRequest represents the request payload for creating a DNS zone.
type CreateZoneRequest struct {
	ZoneName string `json:"zone_name" binding:"required" example:"example.com" minLength:"3" maxLength:"255"`
	Owner    string `json:"owner" binding:"required" example:"user_id_123"` // This typically would be derived from session, but for demo, let's keep it direct.
}

// CreateZoneResponse represents the response payload after creating a DNS zone.
type CreateZoneResponse struct {
	ZoneID string `json:"zone_id" example:"V1StGXR8_Z5jdHi6B-zone"`
}

// CreateZoneHandler creates a new DNS zone.
// @Summary Create DNS Zone
// @Description Creates a new DNS zone with a specified name and owner.
// @Tags zones
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param request body CreateZoneRequest true "Zone creation details"
// @Success 201 {object} CreateZoneResponse "Zone created successfully"
// @Failure 400 {object} GenericErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/zone [post]
func (h *Handler) CreateZoneHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use util.RespondWithJSON for consistent error responses as defined in models.go
		// Assuming util.RespondWithJSON marshals a struct like GenericErrorResponse
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if req.ZoneName == "" || req.Owner == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Zone name and owner are required"})
		return
	}

	zoneId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to generate zone ID"})
		return
	}
	zone := &types.DBZone{
		ID:        zoneId,
		Name:      req.ZoneName,
		Owner:     req.Owner,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.store.CreateZone(zone)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create zone"})
		return
	}

	response := CreateZoneResponse{
		ZoneID: zoneId,
	}

	h.logger.Info("Zone created", "zone_name", req.ZoneName, "zone_id", zoneId, "owner", req.Owner)

	util.RespondWithJSON(w, http.StatusCreated, response) // Use util.RespondWithJSON for consistency
}

// DeleteZoneRequest represents the request payload for deleting a DNS zone.
type DeleteZoneRequest struct {
	ZoneID string `json:"zone_id" binding:"required" example:"V1StGXR8_Z5jdHi6B-zone"`
}

// DeleteZoneResponse represents the response payload after deleting a DNS zone.
type DeleteZoneResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Zone deleted successfully"`
}

// DeleteZoneHandler deletes a DNS zone.
// @Summary Delete DNS Zone
// @Description Deletes a DNS zone by its ID, marking it as deleted.
// @Tags zones
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param request body DeleteZoneRequest true "Zone ID to delete"
// @Success 200 {object} DeleteZoneResponse "Zone deleted successfully"
// @Failure 400 {object} GenericErrorResponse "Invalid request body or missing Zone ID"
// @Failure 404 {object} GenericErrorResponse "Zone not found (or already deleted)"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/zone [delete]
func (h *Handler) DeleteZoneHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if req.ZoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Zone ID is required"})
		return
	}

	// Assuming UpdateZone handles finding the zone and setting DeletedAt
	// You might want to explicitly check if the zone exists before attempting to update/delete
	err := h.store.UpdateZone(&types.DBZone{ID: req.ZoneID, UpdatedAt: time.Now(), DeletedAt: sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}})
	if err != nil {
		// Differentiate between "not found" and "internal error" if possible from store.UpdateZone
		if err == sql.ErrNoRows { // Example check for "not found"
			util.RespondWithJSON(w, http.StatusNotFound, &GenericErrorResponse{Error: true, ErrorMessage: "Zone not found or already deleted"})
			return
		}
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to delete zone"})
		return
	}

	response := DeleteZoneResponse{
		Success: true,
		Message: "Zone deleted successfully",
	}

	util.RespondWithJSON(w, http.StatusOK, response)
	h.logger.Info("Zone deleted", "zone_id", req.ZoneID)
}

// CreateRecordRequest represents the request payload for creating a DNS record.
type CreateRecordRequest struct {
	ZoneID string `json:"zone_id" binding:"required" example:"V1StGXR8_Z5jdHi6B-zone"`
	Name   string `json:"name" binding:"required" example:"www"`
	Type   string `json:"type" binding:"required" example:"A" enums:"A,AAAA,CNAME,MX,TXT,NS,PTR,SRV"`
	Class  string `json:"class" binding:"required" example:"IN" enums:"IN,CH,HS"`
	TTL    uint32 `json:"ttl" example:"300"`
	RData  string `json:"rdata" binding:"required" example:"192.168.1.1"`
}

// CreateRecordResponse represents the response payload after creating a DNS record.
type CreateRecordResponse struct {
	RecordID string `json:"record_id" example:"V1StGXR8_Z5jdHi6B-record"`
}

// CreateRecordHandler creates a new DNS record.
// @Summary Create DNS Record
// @Description Creates a new DNS record within a specified zone.
// @Tags records
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param request body CreateRecordRequest true "DNS record creation details"
// @Success 201 {object} CreateRecordResponse "Record created successfully"
// @Failure 400 {object} GenericErrorResponse "Invalid request body or missing required fields"
// @Failure 404 {object} GenericErrorResponse "Zone not found" // If your store checks for zone existence
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/record [post]
func (h *Handler) CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("error decoding request body", "error", err) // More descriptive log
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if req.ZoneID == "" || req.Name == "" || req.Type == "" || req.Class == "" || req.RData == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "All fields are required (ZoneID, Name, Type, Class, RData)"})
		return
	}

	recordId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to generate record ID"})
		return
	}

	record := &types.DBRecord{ // Ensure types.DBRecord has the necessary fields for MarshalJSON/UnmarshalJSON for Swagger example
		ID:     recordId,
		ZoneID: req.ZoneID,
		Name:   req.Name,
		Type:   req.Type,
		Class:  req.Class,
		TTL:    req.TTL,
		RData:  req.RData,
		// Add CreatedAt/UpdatedAt if your DBRecord has them
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.store.CreateRecord(record)
	if err != nil {
		// You might want to check for specific errors, e.g., if ZoneID doesn't exist.
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create record"})
		return
	}

	response := CreateRecordResponse{
		RecordID: recordId,
	}

	util.RespondWithJSON(w, http.StatusCreated, response)
	h.logger.Info("Record created", "record_id", recordId, "zone_id", req.ZoneID, "name", req.Name)
}

// GetZoneRecordsRequest represents the request payload for getting zone records.
// Note: In your api.go, this handler uses a path parameter {session_id}.
// This request struct likely isn't used for decoding body, but for a query param if changed.
// For consistency with your api.go routing, I'll adjust the handler doc below.
// If it's truly a body request, please confirm.
/*
type GetZoneRecordsRequest struct {
	ZoneName string `json:"zone_name"` // If this is used, it means you send ZoneName in body.
}
*/

// GetZoneRecordsResponse represents the response payload containing records for a zone.
type GetZoneRecordsResponse struct {
	Records []types.DBRecord `json:"records"`
	Count   int              `json:"count" example:"2"` // Add count for better API response
}

// GetZoneRecordsHandler retrieves records for a given zone.
// @Summary Get Zone Records
// @Description Retrieves all DNS records associated with a specific zone based on user session.
// @Tags zones
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param session_id path string true "User Session ID (associated with zone ownership)" example:"V1StGXR8_Z5jdHi6B-myT"
// @Success 200 {object} GetZoneRecordsResponse "Zone records retrieved successfully"
// @Failure 400 {object} GenericErrorResponse "Invalid session ID or missing Zone ID"
// @Failure 404 {object} GenericErrorResponse "Zone or records not found"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/zone/records/{session_id} [get]
func (h *Handler) GetZoneRecordsHandler(w http.ResponseWriter, r *http.Request) {
	// Your api.go defines this as GET /api/v1/zone/records/{session_id}
	// So, ZoneName would likely come from looking up the user's zones
	// or potentially be a query parameter.
	// Your current handler expects a JSON body, which contradicts the GET method
	// and the {session_id} path param. Let's assume you meant to get session_id from path
	// and then fetch user's zones/records based on that.

	// Based on your api.go routing:
	// mux.Handle("GET /api/v1/zone/records/{session_id}", ...)
	sessionId := r.PathValue("session_id")
	if sessionId == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Session ID is required as path parameter"})
		return
	}

	// You need to fetch the zone based on the session ID.
	// This part of your logic needs adjustment if `GetFullZone` expects `ZoneName` from request body
	// but the route implies fetching zones for a user by session ID.
	// Assuming GetFullZone can implicitly get user's zone(s) from session_id
	// or you pass a known zone name based on user's default/selected zone.

	// Placeholder: You'd typically find the user from the session and then their associated zones.
	// For demo purposes, I'll keep your original GetFullZone call but note the potential logic gap.
	// If ZoneName is meant to be derived from the session/user, remove the GetZoneRecordsRequest struct.

	// If GetZoneRecordsRequest is not used, remove its decoding:
	// var req GetZoneRecordsRequest
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil { ... }

	// For example, if you want to allow querying by zone name as well, it should be a query parameter.
	zoneName := r.URL.Query().Get("zone_name") // Example if zone name is a query param
	if zoneName == "" {
		// If zoneName is not provided as query param, then the handler must use session to find default zone(s)
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Zone name query parameter is required (or must be derived from session)"})
		return
	}

	_, records, err := h.store.GetFullZone(zoneName) // Assuming your GetFullZone takes a zone name
	if err != nil {
		if err == sql.ErrNoRows { // Check if zone not found
			util.RespondWithJSON(w, http.StatusNotFound, &GenericErrorResponse{Error: true, ErrorMessage: "Zone not found or no records"})
			return
		}
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to get zone records"})
		return
	}
	if records == nil || len(records) == 0 { // Check if zone found but no records
		util.RespondWithJSON(w, http.StatusNotFound, &GenericErrorResponse{Error: true, ErrorMessage: "Zone found but has no records"})
		return
	}
	response := GetZoneRecordsResponse{
		Records: records,
		Count:   len(records), // Provide a count
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to encode response"})
		return
	}
	h.logger.Info("Zone records retrieved", "zone_name", zoneName, "records_count", len(records))
}
