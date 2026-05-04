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
	PermVenueRead      Permission = "venue:read"
	PermVenueUpdate    Permission = "venue:update"

	// Profile
	PermProfileUpdate Permission = "profile:update"

	// Table
	PermCloseTable Permission = "table:close"

	// Menu / Dish
	PermMenuRead   Permission = "menu:read"
	PermDishCreate Permission = "dish:create"
	PermDishDelete Permission = "dish:delete"

	// Waiter accounts (manager scope)
	PermWaiterList   Permission = "waiter:list"
	PermWaiterCreate Permission = "waiter:create"
	PermWaiterDelete Permission = "waiter:delete"

	// Manager accounts (admin scope)
	PermManagerList   Permission = "manager:list"
	PermManagerCreate Permission = "manager:create"
	PermManagerDelete Permission = "manager:delete"

	// Waiter requests
	PermRequestListWaiter Permission = "request:list_waiter"
	PermRequestAccept     Permission = "request:accept"

	// Tips
	PermTipsRead Permission = "tips:read"

	// Stats
	PermStatsReadVenue  Permission = "stats:read_venue"
	PermStatsReadGlobal Permission = "stats:read_global"
)

type permissionSet map[Permission]struct{}

func (s permissionSet) has(p Permission) bool {
	_, ok := s[p]
	return ok
}
