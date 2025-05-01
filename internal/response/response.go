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

type SummarizeResponse struct {
	Summary string `json:"summary"`
}

type NoteResponse struct {
	ID          uint              `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Summary     string            `json:"summary,omitempty"`
	TopicID     uint              `json:"topic_id,omitempty"`
	Attachments []AttachmentShort `json:"attachments,omitempty"`
	IsArchived  bool              `json:"is_archived"`
	Tags        []TagShort        `json:"tags,omitempty"`
	RelatedIDs  []int64           `json:"related_ids,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

type AttachmentShort struct {
	ID       uint   `json:"id"`
	FileURL  string `json:"file_url"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
}

type TagShort struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type NotesListResponse struct {
	Notes []NoteResponse `json:"notes"`
	Total int            `json:"total"`
}
