package authz

var rolePermissions = map[UserRole]permissionSet{
	RoleAdmin: {
		PermUpdatePassword:  {},
		PermProfileUpdate:   {},
		PermManagerList:     {},
		PermManagerCreate:   {},
		PermManagerDelete:   {},
		PermStatsReadGlobal: {},
	},
	RoleManager: {
		PermUpdatePassword: {},
		PermProfileUpdate:  {},

		PermCreateTable:    {},
		PermDeleteTable:    {},
		PermGetVenueTables: {},
		PermGetTableByID:   {},

		PermVenueRead:   {},
		PermVenueUpdate: {},

		PermMenuRead:   {},
		PermDishCreate: {},
		PermDishDelete: {},

		PermWaiterList:   {},
		PermWaiterCreate: {},
		PermWaiterDelete: {},

		PermStatsReadVenue: {},
	},
	RoleWaiter: {
		PermUpdatePassword: {},
		PermProfileUpdate:  {},

		PermGetWaiterTables: {},
		PermFreeTable:       {},
		PermGetTableByID:    {},
		PermCloseTable:      {},

		PermRequestListWaiter: {},
		PermRequestAccept:     {},

		PermTipsRead: {},
	},
}

func HasPermission(role UserRole, p Permission) bool {
	return rolePermissions[role].has(p)
}
