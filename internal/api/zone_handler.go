package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Unfield/Odin-DNS/internal/models"
	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/go-sql-driver/mysql"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// GetZonesHandler retrieves all zones for the authenticated user
// @Summary Get User Zones
// @Description Returns a list of all DNS zones owned by the authenticated user
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.GetZonesResponse "Zones retrieved successfully"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to get zones"
// @Router /api/v1/zones [get]
func (h *Handler) GetZonesHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	dbZones, err := h.store.GetZones(userSession.UserID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to get zones"})
		return
	}

	var zones []models.ZoneResponse

	for _, current := range dbZones {
		var deletedAt *time.Time = nil
		if current.DeletedAt.Valid {
			deletedAt = &current.DeletedAt.Time
		}
		zones = append(zones, models.ZoneResponse{
			ID:        current.ID,
			Name:      current.Name,
			CreatedAt: current.CreatedAt,
			DeletedAt: deletedAt,
		})
	}

	util.RespondWithJSON(w, http.StatusOK, &models.GetZonesResponse{Count: len(dbZones), Zones: zones})
}

// CreateZoneHandler creates a new DNS zone
// @Summary Create DNS Zone
// @Description Creates a new DNS zone for the authenticated user
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param createZoneRequest body models.CreateZoneRequest true "Zone creation details"
// @Success 200 {object} models.CreateZoneResponse "Zone created successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body, zone already exists, or invalid owner ID"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to create zone"
// @Router /api/v1/zones [post]
func (h *Handler) CreateZoneHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var createZoneRequest models.CreateZoneRequest

	err := json.NewDecoder(r.Body).Decode(&createZoneRequest)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	zoneId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create zone id"})
		return
	}
	zone := types.DBZone{
		ID:        zoneId,
		Owner:     userSession.UserID,
		Name:      createZoneRequest.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{},
	}

	err = h.store.CreateZone(&zone)
	if err != nil {
		// yes this is bad because it directly casts the error to an mysql error but due to the very basic nature of this project we will leave it at this.
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Zone with this name already exists",
				})
				return
			case 1452:
				util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Invalid owner ID or related data not found",
				})
				return
			default:
				util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Failed to create zone due to a database issue",
				})
				return
			}
		}
	}

	util.RespondWithJSON(w, http.StatusOK, &models.CreateZoneResponse{Id: zone.ID})
}

// DeleteZoneHandler deletes an existing DNS zone
// @Summary Delete DNS Record
// @Description Delete an existing DNS zone in the specified zone
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Success 200 {object} models.DeleteZoneResponse "Zone deleted successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to delete zone"
// @Router /api/v1/zone/{zone_id} [delete]
func (h *Handler) DeleteZoneHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var zoneID = r.PathValue("zone_id")
	if zoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone_id missing"})
		return
	}

	err := h.store.DeleteZone(zoneID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "failed to delete record"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &models.DeleteZoneResponse{Id: zoneID})
}

// GetZoneRecordsHandler retrieves all records for a specific zone
// @Summary Get Zone Records
// @Description Returns all DNS records for a specific zone
// @Tags records
// @Security BearerAuth
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Success 200 {object} models.GetZoneRecordsResponse "Zone records retrieved successfully"
// @Failure 400 {object} models.GenericErrorResponse "Missing zone_id parameter"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to get zone records or parse MX value"
// @Router /api/v1/zone/{zone_id}/entries [get]
func (h *Handler) GetZoneRecordsHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var zoneID = r.PathValue("zone_id")
	if zoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone_id missing"})
		return
	}

	dbZoneRecords, err := h.store.GetZoneEntries(zoneID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to get zones"})
		return
	}

	var records []models.ZoneRecordResponse

	for _, current := range dbZoneRecords {
		var priority *uint16 = nil
		var value string = ""
		if current.Type == "MX" {
			prio, val, err := util.ConvertMXRData(current.RData)
			if err != nil {
				util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to parse MX value"})
				return
			}
			priority = &prio
			value = val
		} else {
			value = current.RData
		}
		records = append(records, models.ZoneRecordResponse{
			ID:       current.ID,
			Name:     current.Name,
			Type:     current.Type,
			Class:    current.Class,
			TTl:      current.TTL,
			Priority: priority,
			Value:    value,
		})
	}

	util.RespondWithJSON(w, http.StatusOK, &models.GetZoneRecordsResponse{Count: len(records), Records: records})
}

// CreateZoneEntryHandler creates a new DNS record in a zone
// @Summary Create DNS Record
// @Description Creates a new DNS record in the specified zone
// @Tags records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Param createZoneEntryRequest body models.CreateZoneEntryRequest true "DNS record details"
// @Success 200 {object} models.CreateZoneEntryResponse "Zone record created successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body, missing zone_id, missing priority for MX record, or entry already exists"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to create zone record"
// @Router /api/v1/zone/{zone_id}/entries [post]
func (h *Handler) CreateZoneEntryHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var zoneID = r.PathValue("zone_id")
	if zoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone_id missing"})
		return
	}

	zone, err := h.store.GetZone(zoneID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone not found"})
		return
	}

	var createZoneEntryRequest models.CreateZoneEntryRequest

	err = json.NewDecoder(r.Body).Decode(&createZoneEntryRequest)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	zoneEntryId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create entry id"})
		return
	}

	var rdata string

	createZoneEntryRequest.Name = strings.TrimSuffix(createZoneEntryRequest.Name, ".")

	if !strings.HasSuffix(createZoneEntryRequest.Name, zone.Name) {
		createZoneEntryRequest.Name = fmt.Sprintf("%s.%s", createZoneEntryRequest.Name, zone.Name)
	}

	createZoneEntryRequest.Name = strings.TrimPrefix(createZoneEntryRequest.Name, "@")
	createZoneEntryRequest.Name = strings.TrimPrefix(createZoneEntryRequest.Name, ".")

	if createZoneEntryRequest.Type == "MX" {
		if createZoneEntryRequest.Priority == nil {
			util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "priority missing"})
			return
		}
		prioString := strconv.FormatUint(uint64(*createZoneEntryRequest.Priority), 10)
		rdata = strings.Join([]string{prioString, createZoneEntryRequest.Value}, " ")
	} else {
		rdata = createZoneEntryRequest.Value
	}

	entry := types.DBRecord{
		ID:     zoneEntryId,
		ZoneID: zoneID,
		Name:   createZoneEntryRequest.Name,
		Type:   createZoneEntryRequest.Type,
		Class:  createZoneEntryRequest.Class,
		TTL:    createZoneEntryRequest.TTl,
		RData:  rdata,
	}

	err = h.store.CreateRecord(&entry)
	if err != nil {
		// same as with create zone...
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Entry already exists",
				})
				return
			default:
				util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Failed to create entry due to a database issue",
				})
				return
			}
		}
	}

	util.RespondWithJSON(w, http.StatusOK, &models.CreateZoneEntryResponse{Id: entry.ID})
}

// UpdateZoneEntryHandler updates an existing DNS record
// @Summary Update DNS Record
// @Description Updates an existing DNS record in the specified zone
// @Tags records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Param entry_id path string true "Entry ID"
// @Param updateZoneEntryRequest body models.UpdateZoneEntryRequest true "Updated DNS record details"
// @Success 200 {object} models.UpdateZoneEntryResponse "Zone record updated successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body, missing parameters, missing priority for MX record, or entry already exists"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to update zone record"
// @Router /api/v1/zone/{zone_id}/entry/{entry_id} [put]
func (h *Handler) UpdateZoneEntryHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var zoneID = r.PathValue("zone_id")
	if zoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone_id missing"})
		return
	}

	var entryID = r.PathValue("entry_id")
	if entryID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "entry_id missing"})
		return
	}

	var updateZoneEntryRequest models.UpdateZoneEntryRequest

	err := json.NewDecoder(r.Body).Decode(&updateZoneEntryRequest)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	var rdata string

	if updateZoneEntryRequest.Type == "MX" {
		if updateZoneEntryRequest.Priority == nil {
			util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "priority missing"})
			return
		}
		prioString := strconv.FormatUint(uint64(*updateZoneEntryRequest.Priority), 10)
		rdata = strings.Join([]string{prioString, updateZoneEntryRequest.Value}, " ")
	} else {
		rdata = updateZoneEntryRequest.Value
	}

	entry := types.DBRecord{
		ID:     entryID,
		ZoneID: zoneID,
		Name:   updateZoneEntryRequest.Name,
		Type:   updateZoneEntryRequest.Type,
		Class:  updateZoneEntryRequest.Class,
		TTL:    updateZoneEntryRequest.TTl,
		RData:  rdata,
	}

	err = h.store.UpdateRecord(&entry)
	if err != nil {
		// same as with create zone...
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Entry already exists",
				})
				return
			default:
				util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{
					Error:        true,
					ErrorMessage: "Failed to create entry due to a database issue",
				})
				return
			}
		}
	}

	util.RespondWithJSON(w, http.StatusOK, &models.UpdateZoneEntryResponse{Id: entry.ID})
}

// DeleteZoneEntryHandler deletes an existing DNS record
// @Summary Delete DNS Record
// @Description Delete an existing DNS record in the specified zone
// @Tags records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Param entry_id path string true "Entry ID"
// @Success 200 {object} models.DeleteZoneEntryResponse "Zone record deleted successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Failed to delete zone record"
// @Router /api/v1/zone/{zone_id}/entry/{entry_id} [delete]
func (h *Handler) DeleteZoneEntryHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)
	if !sessionValid || userSession.Token == "" || userSession.UserID == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	var zoneID = r.PathValue("zone_id")
	if zoneID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "zone_id missing"})
		return
	}

	var entryID = r.PathValue("entry_id")
	if entryID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "entry_id missing"})
		return
	}

	entry, err := h.store.GetRecord(entryID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "record not found missing"})
		return
	}

	if entry.ZoneID != zoneID {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "record not part of that zone"})
		return
	}

	// we would ususally check if the user has access to delete this entry but we are gonna skip it for this simple demo

	err = h.store.DeleteRecord(entryID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "failed to delete record"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &models.DeleteZoneEntryResponse{Id: entry.ID})
}
