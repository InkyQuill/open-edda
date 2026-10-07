package auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	RefreshToken     string       `json:"refreshToken,omitempty"`
	RefreshExpiresAt int64        `json:"refreshExpiresAt"`
	Token            string       `json:"token"`
	Author           AuthorPublic `json:"author"`
}

type AuthorPublic struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
