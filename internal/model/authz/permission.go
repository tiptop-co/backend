package authz

// permission style: '<resource>:<action>'
//
// Example: 'order:update', 'table:close', etc.
type Permission string

const (
	PermUpdatePassword Permission = "password:update"
)

type permissionSet map[Permission]struct{}

func (s permissionSet) has(p Permission) bool {
	_, ok := s[p]
	return ok
}
