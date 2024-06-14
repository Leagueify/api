package model

type EmailConfig struct {
	ID        string
	Email     string `json:"email" validate:"required,email"`
	SMTPHost  string `json:"smtpHost" validate:"required"`
	SMTPPort  int    `json:"smtpPort" validate:"required"`
	SMTPUser  string `json:"smtpUser" validate:"required"`
	SMTPPass  string `json:"smtpPass" validate:"required"`
	IsEnabled bool
	HasError  bool
}
