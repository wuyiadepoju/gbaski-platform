package common

// APIResponse represents a standard API response
type Response struct {
	Status      string      `json:"status" example:"success"`
	Message     string      `json:"message"`
	Data        interface{} `json:"data,omitempty"`
	Error       interface{} `json:"error,omitempty"`
	OtpRequired bool        `json:"otp_required,omitempty"`
}
