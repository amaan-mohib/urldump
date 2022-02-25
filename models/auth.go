package models

type Token struct {
	AccessToken  string
	IdToken      string
	ExpiresIn    float64
	ExpiresAt    float64
	RefreshToken string
	TokenType    string
}

type User struct {
	Name  string
	Email string
	Sub   string
}
