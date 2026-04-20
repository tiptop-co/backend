package auth

type AuthInfo struct {
	HashedPassword string `db:"password"`
	*Claims
}
