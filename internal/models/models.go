package models

import (
	"time"
)

type GenericErrorResponse struct {
	Error        bool   `json:"error" example:"true" description:"Indicates if an error occurred"`
	ErrorMessage string `json:"error_message" example:"Invalid request body" description:"Human-readable error message"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe" description:"User's username"`
	Password string `json:"password" binding:"required" example:"password123" description:"User's password"`
}

type LoginResponse struct {
	SessionID string `json:"session_id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique session identifier"`
	Token     string `json:"token" example:"kWnEeaODiH5Sb1H1REbfLA3VTl7jbvlpAn4vKNDXEcgOcgdmRhRjRb" description:"Bearer token for authentication"`
	Username  string `json:"username" example:"john_doe" description:"Authenticated user's username"`
}

type RegisterRequest struct {
	Username        string `json:"username" binding:"required" example:"john_doe" minLength:"3" maxLength:"50" description:"Desired username (3-50 characters)"`
	Email           string `json:"email" binding:"required" example:"john@example.com" format:"email" description:"User's email address"`
	Password        string `json:"password" binding:"required" example:"password123" minLength:"8" description:"Password (minimum 8 characters)"`
	PasswordConfirm string `json:"password_confirm" binding:"required" example:"password123" minLength:"8" description:"Password confirmation (must match password)"`
}

type RegisterResponse struct {
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique user identifier"`
	Username string `json:"username" example:"john_doe" description:"Registered username"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"Successfully logged out" description:"Confirmation message"`
}

type GetUserResponse struct {
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique user identifier"`
	Username string `json:"username" example:"john_doe" description:"User's username"`
	Email    string `json:"email" example:"john@example.com" description:"User's email address"`
}

type TimeSeriesData struct {
	Time       time.Time `json:"time" example:"2025-01-01T00:00:00Z"`
	Requests   uint64    `json:"requests" example:"1234"`
	Errors     int64     `json:"errors" example:"5"`
	Percentage float64   `json:"percentage,omitempty" example:"95.5"`
}

type GlobalAvgMetrics struct {
	AvgResponseTimeMs        float64 `json:"avgResponseTimeMs" example:"25.34"`
	AvgSuccessResponseTimeMs float64 `json:"avgSuccessResponseTimeMs" example:"20.15"`
	AvgErrorResponseTimeMs   float64 `json:"avgErrorResponseTimeMs" example:"150.88"`
	CacheHitPercentage       float64 `json:"cacheHitPercentage" example:"85.23"`
	TotalRequests            uint64  `json:"totalRequests" example:"100000"`
	TotalErrors              uint64  `json:"totalErrors" example:"500"`
}

type TopNData struct {
	Name  string `json:"name" example:"example.com"`
	Count uint64 `json:"count" example:"5000"`
}

type RcodeData struct {
	Rcode     uint8  `json:"rcode" example:"3"`
	Count     uint64 `json:"count" example:"150"`
	RcodeName string `json:"rcodeName" example:"NXDOMAIN"`
}

type ZoneResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type GetZonesResponse struct {
	Count int            `json:"count"`
	Zones []ZoneResponse `json:"zones"`
}

type CreateZoneRequest struct {
	Name string `json:"name" binding:"required" example:"example.com" description:"Domain name for the zone"`
}

type CreateZoneResponse struct {
	Id string `json:"id"`
}

type ZoneRecordResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Class    string  `json:"class"`
	TTl      uint32  `json:"ttl"`
	Priority *uint16 `json:"priority,omitempty"`
	Value    string  `json:"value"`
}

type GetZoneRecordsResponse struct {
	Count   int                  `json:"count"`
	Records []ZoneRecordResponse `json:"records"`
}

type CreateZoneEntryRequest struct {
	Name     string  `json:"name" example:"www" description:"Record name (subdomain)"`
	Type     string  `json:"type" example:"A" description:"DNS record type (A, AAAA, CNAME, MX, TXT, etc.)"`
	Class    string  `json:"class" example:"IN" description:"DNS record class (typically 'IN')"`
	TTl      uint32  `json:"ttl" example:"300" description:"Time to live in seconds"`
	Priority *uint16 `json:"priority,omitempty" example:"10" description:"Priority for MX records (required for MX type)"`
	Value    string  `json:"value" example:"192.168.1.1" description:"Record value (IP address, hostname, etc.)"`
}

type CreateZoneEntryResponse struct {
	Id string `json:"id"`
}

type UpdateZoneEntryRequest struct {
	Name     string  `json:"name" example:"www" description:"Record name (subdomain)"`
	Type     string  `json:"type" example:"A" description:"DNS record type (A, AAAA, CNAME, MX, TXT, etc.)"`
	Class    string  `json:"class" example:"IN" description:"DNS record class (typically 'IN')"`
	TTl      uint32  `json:"ttl" example:"300" description:"Time to live in seconds"`
	Priority *uint16 `json:"priority,omitempty" example:"10" description:"Priority for MX records (required for MX type)"`
	Value    string  `json:"value" example:"192.168.1.1" description:"Record value (IP address, hostname, etc.)"`
}

type GetZoneResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type UpdateZoneEntryResponse struct {
	Id string `json:"id"`
}

type DeleteZoneEntryResponse struct {
	Id string `json:"id"`
}

type DeleteZoneResponse struct {
	Id string `json:"id"`
}
