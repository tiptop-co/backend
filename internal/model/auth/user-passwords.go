package auth

type UserPasswords struct {
	Old string `json:"old"`
	New string `json:"new"`
}
