package authz

var rolePermissions = map[UserRole]permissionSet{
	RoleAdmin: {
		PermUpdatePassword: {},
	},
	RoleManager: {
		PermUpdatePassword: {},
	},
	RoleWaiter: {
		PermUpdatePassword: {},
	},
}

func HasPermission(role UserRole, p Permission) bool {
	return rolePermissions[role].has(p)
}
