package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	Admin   UserRole = "admin"
	AppUser UserRole = "appuser"
)

// swagger:model domain.User
type User struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=50"`
	DateOfBirth string    `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Role        UserRole  `json:"role"`
}
