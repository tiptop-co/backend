package authz

var rolePermissions = map[UserRole]permissionSet{
	RoleAdmin: {
		PermUpdatePassword: {},
	},
	RoleManager: {
		PermUpdatePassword: {},
		PermCreateTable:    {},
		PermDeleteTable:    {},
		PermGetVenueTables: {},
		PermGetTableByID:   {},
	},
	RoleWaiter: {
		PermUpdatePassword:  {},
		PermGetWaiterTables: {},
		PermFreeTable:       {},
		PermGetTableByID:    {},
	},
}

func HasPermission(role UserRole, p Permission) bool {
	return rolePermissions[role].has(p)
}
