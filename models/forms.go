package models

type ContactForm struct {
	FirstName      string `validate:"required,max=50"`
	LastName       string `validate:"required,max=50"`
	Email          string `validate:"required,email"`
	Subject        string `validate:"required,max=200"`
	Message        string `validate:"required"`
	RecaptchaToken string `validate:"required"`
}
