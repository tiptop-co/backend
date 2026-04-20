package auth

import "strconv"

type UserRole int

const (
	RoleUnspecified = iota
	RoleGuest
	RoleWaiter
	RoleManager
	RoleAdmin
)

func (u UserRole) String() string {
	return strconv.Itoa(int(u))
}

func (u UserRole) Name() string {
	return []string{"Unspecified", "Guest", "Waiter", "Manager", "Admin"}[u]
}

func (u UserRole) IsValid() bool {
	return u >= RoleUnspecified && u <= RoleAdmin
}
