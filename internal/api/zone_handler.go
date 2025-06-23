package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Unfield/Odin-DNS/internal/types"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type CreateZoneRequest struct {
	ZoneName string `json:"zone_name"`
	Owner    string `json:"owner"`
}

type CreateZoneResponse struct {
	ZoneID string `json:"zone_id"`
}

func (h *Handler) CreateZoneHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ZoneName == "" || req.Owner == "" {
		http.Error(w, "Zone name and owner are required", http.StatusBadRequest)
		return
	}

	zoneId, err := gonanoid.New()
	if err != nil {
		http.Error(w, "Failed to generate zone ID", http.StatusInternalServerError)
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
		http.Error(w, "Failed to create zone", http.StatusInternalServerError)
		return
	}

	response := CreateZoneResponse{
		ZoneID: zoneId,
	}

	h.logger.Info("Zone created", "zone_name", req.ZoneName, "zone_id", zoneId, "owner", req.Owner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Zone created", "zone_name", req.ZoneName, "zone_id", zoneId, "owner", req.Owner)
}

type DeleteZoneRequest struct {
	ZoneID string `json:"zone_id"`
}

type DeleteZoneResponse struct {
	Success bool `json:"success"`
}

func (h *Handler) DeleteZoneHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ZoneID == "" {
		http.Error(w, "Zone ID is required", http.StatusBadRequest)
		return
	}

	err := h.store.UpdateZone(&types.DBZone{ID: req.ZoneID, UpdatedAt: time.Now(), DeletedAt: sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}})
	if err != nil {
		http.Error(w, "Failed to delete zone", http.StatusInternalServerError)
		return
	}

	response := DeleteZoneResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Zone deleted", "zone_id", req.ZoneID)
}

type CreateRecordRequest struct {
	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Class  string `json:"class"`
	TTL    uint32 `json:"ttl"`
	RData  string `json:"rdata"`
}

type CreateRecordResponse struct {
	RecordID string `json:"record_id"`
}

func (h *Handler) CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("error", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ZoneID == "" || req.Name == "" || req.Type == "" || req.Class == "" || req.RData == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	recordId, err := gonanoid.New()
	if err != nil {
		http.Error(w, "Failed to generate record ID", http.StatusInternalServerError)
		return
	}

	record := &types.DBRecord{
		ID:     recordId,
		ZoneID: req.ZoneID,
		Name:   req.Name,
		Type:   req.Type,
		Class:  req.Class,
		TTL:    req.TTL,
		RData:  req.RData,
	}

	err = h.store.CreateRecord(record)
	if err != nil {
		http.Error(w, "Failed to create record", http.StatusInternalServerError)
		return
	}

	response := CreateRecordResponse{
		RecordID: recordId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Record created", "record_id", recordId, "zone_id", req.ZoneID, "name", req.Name)
}

type GetZoneRecordsRequest struct {
	ZoneName string `json:"zone_name"`
}

type GetZoneRecordsResponse struct {
	Records []types.DBRecord `json:"records"`
}

func (h *Handler) GetZoneRecordsHandler(w http.ResponseWriter, r *http.Request) {
	var req GetZoneRecordsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ZoneName == "" {
		http.Error(w, "Zone ID is required", http.StatusBadRequest)
		return
	}
	_, records, err := h.store.GetFullZone(req.ZoneName)
	if err != nil {
		http.Error(w, "Failed to get zone records", http.StatusInternalServerError)
		return
	}
	if records == nil {
		http.Error(w, "Zone has no records", http.StatusNotFound)
		return
	}
	response := GetZoneRecordsResponse{
		Records: records,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Zone records retrieved", "zone_name", req.ZoneName, "records_count", len(records))
}
