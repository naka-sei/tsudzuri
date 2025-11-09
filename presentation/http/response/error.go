package response

// ErrResponse represents an error response with a message.
type ErrResponse struct {
	Message string `json:"message"`
}
