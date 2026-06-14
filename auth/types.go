package auth

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string       `json:"token"`
	Author AuthorPublic `json:"author"`
}

type AuthorPublic struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
