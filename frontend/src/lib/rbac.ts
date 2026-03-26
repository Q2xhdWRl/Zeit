import type { User, UserRole } from "@/lib/auth";

/** Check if the user has one of the specified roles. */
export function hasRole(user: User | null, ...roles: UserRole[]): boolean {
  if (!user) return false;
  return roles.includes(user.global_role);
}

/** Check if the user is an admin. */
export function isAdmin(user: User | null): boolean {
  return hasRole(user, "admin");
}

/** Check if the user is a team leader or admin. */
export function isTeamLeaderOrAdmin(user: User | null): boolean {
  return hasRole(user, "admin", "team_leader");
}
