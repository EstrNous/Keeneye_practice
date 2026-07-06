package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,customemail"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,customemail"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"required,oneof=student teacher admin"`
	PhoneNumber string `json:"phone_number" binding:"required,customphone"`
	ProfileFIO  string `json:"profile_fio" binding:"required"`
	GroupID     *int32 `json:"group_id"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateStudentRequest struct {
	Email       string `json:"email" binding:"required,customemail"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number" binding:"required,customphone"`
	GroupID     int32  `json:"group_id" binding:"required"`
	Fio         string `json:"fio" binding:"required"`
}

type UpdateStudentRequest struct {
	GroupID int32  `json:"group_id"`
	Fio     string `json:"fio" binding:"required"`
}

type UpdateTeacherRequest struct {
	Fio string `json:"fio" binding:"required"`
}

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
