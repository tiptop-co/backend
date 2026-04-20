package auth

type UserCredentials struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}
