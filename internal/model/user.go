package model

import "github.com/google/uuid"

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
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
	Roles []string  `json:"roles"`
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
