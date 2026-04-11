package authz

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

var supportedRoles = []string{
	RoleAdmin,
	RoleOperator,
	RoleViewer,
}

var supportedRoleSet = map[string]struct{}{
	RoleAdmin:    {},
	RoleOperator: {},
	RoleViewer:   {},
}

func SupportedRoles() []string {
	roles := make([]string, len(supportedRoles))
	copy(roles, supportedRoles)
	return roles
}

func IsSupportedRole(role string) bool {
	_, ok := supportedRoleSet[role]
	return ok
}

func DefaultRole(role string) string {
	if role == "" {
		return RoleViewer
	}
	return role
}
