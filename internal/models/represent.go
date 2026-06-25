package models

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"Invalid request body"`
}

// BadRequestResponse represents an error response
type BadRequestResponse struct {
	Message string `json:"message" example:"Not found"`
}

// ServerErrorResponse represents an error response
type ServerErrorResponse struct {
	Message string `json:"message" example:"Internal server error"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// TaskListResponse represents response with tasks list
type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
}

// SetStatusRequest represents status update request
type SetStatusRequest struct {
	Status string `json:"status" example:"done"`
}

// OverrideMessageResponse представляет ответ при удалении переопределения
type OverrideMessageResponse struct {
	Message string `json:"message" example:"Override for task 1 on 2024-01-15 deleted successfully"`
}
