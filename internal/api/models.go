package api

type GenericErrorResponse struct {
	Error        bool   `json:"error" example:"true"`
	ErrorMessage string `json:"error_message" example:"Invalid request body"`
}
