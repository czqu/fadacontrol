package schema

type TokenResponse struct {
	Token string `json:"token"`
}
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
