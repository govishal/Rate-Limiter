package models

// PostRequest is the JSON body for POST /request.
type PostRequest struct {
	UserID  string                 `json:"user_id"`
	Payload map[string]interface{} `json:"payload"`
}
