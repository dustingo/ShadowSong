package authz

type Capability string

const (
	CapabilityViewAlerts    Capability = "view_alerts"
	CapabilityProcessAlerts Capability = "process_alerts"
	CapabilityViewConfig    Capability = "view_config"
	CapabilityManageConfig  Capability = "manage_config"
	CapabilityManageUsers   Capability = "manage_users"
)

var roleCapabilities = map[string]map[Capability]struct{}{
	RoleAdmin: {
		CapabilityViewAlerts:    {},
		CapabilityProcessAlerts: {},
		CapabilityViewConfig:    {},
		CapabilityManageConfig:  {},
		CapabilityManageUsers:   {},
	},
	RoleOperator: {
		CapabilityViewAlerts:    {},
		CapabilityProcessAlerts: {},
		CapabilityViewConfig:    {},
	},
	RoleViewer: {
		CapabilityViewAlerts: {},
		CapabilityViewConfig: {},
	},
}

func Can(role string, capability Capability) bool {
	capabilities, ok := roleCapabilities[role]
	if !ok {
		return false
	}

	_, ok = capabilities[capability]
	return ok
}

func CanViewAlerts(role string) bool {
	return Can(role, CapabilityViewAlerts)
}

func CanProcessAlerts(role string) bool {
	return Can(role, CapabilityProcessAlerts)
}

func CanViewConfig(role string) bool {
	return Can(role, CapabilityViewConfig)
}

func CanManageConfig(role string) bool {
	return Can(role, CapabilityManageConfig)
}

func CanManageUsers(role string) bool {
	return Can(role, CapabilityManageUsers)
}
