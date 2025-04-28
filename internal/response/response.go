package response

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJI..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOi..."`
}

type ProfileResponse struct {
	Nickname   string `json:"nickname,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"` // Ссылка на фото профиля
}

type UploadAvatarResponse struct {
	Message    string `json:"message"`
	ProfilePic string `json:"profile_pic"`
}
