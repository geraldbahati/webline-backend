package model

import (
	"time"

	"github.com/google/uuid"
)

type RegisterUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserSchema struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	Name          string     `json:"name"`
	Image         string     `json:"image"`
	Phone         string     `json:"phone"`
	DateOfBirth   string     `json:"dateOfBirth"`
	IsActive      bool       `json:"isActive"`
	Provider      *string    `json:"provider"`
	ProviderID    *string    `json:"providerID"`
	EmailVerified *time.Time `json:"emailVerified"`
	Roles         []string   `json:"roles"`
}

type UserProfile struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	ProfileImageUrl    string     `json:"profileImageUrl"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	PhoneNumber        string     `json:"phoneNumber"`
	DateOfBirth        *time.Time `json:"dateOfBirth"`
	RequestAdmin       bool       `json:"requestAdmin"`
	AdminRequestReason string     `json:"adminRequestReason"`
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserProfileParams struct {
	Id              string `json:"id"`
	ProfileImageUrl string `json:"profile_image_url"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	PhoneNumber     string `json:"phone_number"`
	DateOfBirth     string `json:"date_of_birth"`
}

type GoogleUser struct {
	ID            string  `json:"id"`            // Unique identifier
	Email         string  `json:"email"`         // User's email address
	VerifiedEmail *bool   `json:"verifiedEmail"` // Whether the email is verified (optional)
	Name          *string `json:"name"`          // Full name (optional)
	GivenName     *string `json:"givenName"`     // First name (optional)
	FamilyName    *string `json:"familyName"`    // Last name (optional)
	Picture       *string `json:"image"`         // URL to the user's profile picture (optional)
	Locale        *string `json:"locale"`        // User's locale (optional)
	HD            *string `json:"hd"`            // Hosted domain, if any (optional)
}

type UpdateUserInfoParams struct {
	Email              string `json:"email"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	PhoneNumber        string `json:"phoneNumber"`
	DateOfBirth        string `json:"dateOfBirth"`
	AdminRequestReason string `json:"adminRequestReason"`
	RequestAdmin       bool   `json:"requestAdmin"`
}
