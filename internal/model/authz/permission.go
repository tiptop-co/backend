package authz

// permission style: '<resource>:<action>'
//
// Example: 'order:update', 'table:close', etc.
type Permission string

const (
	// User
	PermUpdatePassword Permission = "password:update"

	// Table
	PermCreateTable  Permission = "table:create"
	PermDeleteTable  Permission = "table:delete"
	PermFreeTable    Permission = "table:free"
	PermGetTableByID Permission = "table:get"

	// Waiter
	PermGetWaiterTables Permission = "waiter:get_tables"

	// Venue
	PermGetVenueTables Permission = "venue:get_tables"
)

type permissionSet map[Permission]struct{}

func (s permissionSet) has(p Permission) bool {
	_, ok := s[p]
	return ok
}
