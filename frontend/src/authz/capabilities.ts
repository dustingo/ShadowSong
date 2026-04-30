import type { User } from '../types'

export type Role = User['role']

export const capabilityViewAlerts = 'view_alerts'
export const capabilityProcessAlerts = 'process_alerts'
export const capabilityViewConfig = 'view_config'
export const capabilityManageConfig = 'manage_config'
export const capabilityManageUsers = 'manage_users'

export type Capability =
  | typeof capabilityViewAlerts
  | typeof capabilityProcessAlerts
  | typeof capabilityViewConfig
  | typeof capabilityManageConfig
  | typeof capabilityManageUsers

const roleMatrix: Record<Role, Capability[]> = {
  admin: [
    capabilityViewAlerts,
    capabilityProcessAlerts,
    capabilityViewConfig,
    capabilityManageConfig,
    capabilityManageUsers,
  ],
  operator: [capabilityViewAlerts, capabilityProcessAlerts, capabilityViewConfig],
  viewer: [capabilityViewAlerts, capabilityViewConfig],
}

export function can(role: Role | undefined, capability: Capability): boolean {
  if (!role) {
    return false
  }

  return roleMatrix[role]?.includes(capability) ?? false
}

export function canUser(user: User | null | undefined, capability: Capability): boolean {
  return can(user?.role, capability)
}

export function isAdmin(user: User | null | undefined): boolean {
  return canUser(user, capabilityManageUsers)
}

export function isReadOnlyConfigUser(user: User | null | undefined): boolean {
  return canUser(user, capabilityViewConfig) && !canUser(user, capabilityManageConfig)
}

export function canProcessAlerts(user: User | null | undefined): boolean {
  return canUser(user, capabilityProcessAlerts)
}

export function canRecoverDeliveries(user: User | null | undefined): boolean {
  return canUser(user, capabilityProcessAlerts)
}
