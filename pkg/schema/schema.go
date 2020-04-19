package schema

import (
	"github.com/dovbysh/go-skeleton/pkg/models"
	"time"
)

type HealthResponse struct {
	R    string    `json:"r"`
	Time time.Time `json:"time"`
}

type RegisterRequest struct {
	Email         string `json:"email"`
	PasswordPlain string `json:"password_plain"`
	Name          string `json:"name"`
}

type RegisterResponse struct {
	User *models.User `json:"user,omitempty"`
}
type LoginRequest struct {
	Email         string `json:"email"`
	PasswordPlain string `json:"password_plain"`
}
type LoginResponse struct {
	Bearer string `json:"bearer"`
}

type HelloResponse struct {
	Now  time.Time   `json:"now"`
	User models.User `json:"user"`
}
