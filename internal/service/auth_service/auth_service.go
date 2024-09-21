package auth_service

import "github.com/casbin/casbin/v2"

type AuthService struct {
	enforcer *casbin.Enforcer
}
type Action string

const (
	Read  Action = "r"
	Write Action = "w"
)

const HttpPrefix = "api:"

func NewAuthService(enforcer *casbin.Enforcer) *AuthService {
	return &AuthService{enforcer: enforcer}
}

func (a *AuthService) CheckHttpPermission(username string, path string, act Action) bool {
	if username == "" {
		username = "*"
	}
	path = HttpPrefix + path
	allowed, err := a.enforcer.Enforce(username, path, string(act))
	if err != nil {
		return false
	}
	return allowed
}
