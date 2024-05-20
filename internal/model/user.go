package model

type RegisterUserParams struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
