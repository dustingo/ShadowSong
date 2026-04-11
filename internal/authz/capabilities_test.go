package authz

import "testing"

func TestCapabilityMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		capability Capability
		want       bool
	}{
		{name: "admin can view alerts", role: RoleAdmin, capability: CapabilityViewAlerts, want: true},
		{name: "admin can process alerts", role: RoleAdmin, capability: CapabilityProcessAlerts, want: true},
		{name: "admin can view config", role: RoleAdmin, capability: CapabilityViewConfig, want: true},
		{name: "admin can manage config", role: RoleAdmin, capability: CapabilityManageConfig, want: true},
		{name: "admin can manage users", role: RoleAdmin, capability: CapabilityManageUsers, want: true},
		{name: "operator can view alerts", role: RoleOperator, capability: CapabilityViewAlerts, want: true},
		{name: "operator can process alerts", role: RoleOperator, capability: CapabilityProcessAlerts, want: true},
		{name: "operator can view config", role: RoleOperator, capability: CapabilityViewConfig, want: true},
		{name: "operator cannot manage config", role: RoleOperator, capability: CapabilityManageConfig, want: false},
		{name: "operator cannot manage users", role: RoleOperator, capability: CapabilityManageUsers, want: false},
		{name: "viewer can view alerts", role: RoleViewer, capability: CapabilityViewAlerts, want: true},
		{name: "viewer can view config", role: RoleViewer, capability: CapabilityViewConfig, want: true},
		{name: "viewer cannot process alerts", role: RoleViewer, capability: CapabilityProcessAlerts, want: false},
		{name: "viewer cannot manage config", role: RoleViewer, capability: CapabilityManageConfig, want: false},
		{name: "viewer cannot manage users", role: RoleViewer, capability: CapabilityManageUsers, want: false},
		{name: "unsupported role denied", role: "guest", capability: CapabilityViewAlerts, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Can(tt.role, tt.capability); got != tt.want {
				t.Fatalf("Can(%q, %q) = %v, want %v", tt.role, tt.capability, got, tt.want)
			}
		})
	}
}
