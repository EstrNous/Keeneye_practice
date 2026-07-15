package dto

type BatchUploadResponse struct {
	BatchID string          `json:"batch_id"`
	Total   int             `json:"total"`
	Created int             `json:"created"`
	Failed  int             `json:"failed"`
	Partial bool            `json:"partial"`
	Errors  []BatchRowError `json:"errors,omitempty"`
}

type BatchRowError struct {
	Row     int    `json:"row"`
	Email   string `json:"email,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchStatusResponse struct {
	BatchID      string                    `json:"batch_id"`
	TotalRows    int32                     `json:"total_rows"`
	SuccessCount int32                     `json:"success_count"`
	ErrorCount   int32                     `json:"error_count"`
	Requests     []RegistrationRequestItem `json:"requests"`
}

type RegistrationRequestItem struct {
	ID        int32  `json:"id"`
	Email     string `json:"email"`
	Fio       string `json:"fio"`
	Role      string `json:"role"`
	GroupName string `json:"group_name,omitempty"`
	Status    string `json:"status"`
}

type CompleteRegistrationPreview struct {
	Email     string `json:"email"`
	Fio       string `json:"fio"`
	Role      string `json:"role"`
	GroupName string `json:"group_name,omitempty"`
}

type CompleteRegistrationRequest struct {
	Token       string `json:"token" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number" binding:"required,customphone"`
}

type CompleteRegistrationResponse struct {
	UserID int32 `json:"user_id"`
}
