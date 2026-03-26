import { describe, it, expect } from "vitest";
import { hasRole, isAdmin, isTeamLeaderOrAdmin } from "@/lib/rbac";
import type { User } from "@/lib/auth";

function makeUser(role: "admin" | "team_leader" | "user"): User {
  return {
    id: "1",
    email: "test@test.com",
    display_name: "Test User",
    global_role: role,
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
  };
}

describe("hasRole", () => {
  it("returns true when user has the specified role", () => {
    expect(hasRole(makeUser("admin"), "admin")).toBe(true);
  });

  it("returns false when user has a different role", () => {
    expect(hasRole(makeUser("user"), "admin")).toBe(false);
  });

  it("returns false for null user", () => {
    expect(hasRole(null, "admin")).toBe(false);
  });

  it("accepts multiple roles", () => {
    expect(hasRole(makeUser("team_leader"), "admin", "team_leader")).toBe(true);
  });
});

describe("isAdmin", () => {
  it("returns true for admin", () => {
    expect(isAdmin(makeUser("admin"))).toBe(true);
  });

  it("returns false for non-admin", () => {
    expect(isAdmin(makeUser("user"))).toBe(false);
    expect(isAdmin(makeUser("team_leader"))).toBe(false);
  });
});

describe("isTeamLeaderOrAdmin", () => {
  it("returns true for admin", () => {
    expect(isTeamLeaderOrAdmin(makeUser("admin"))).toBe(true);
  });

  it("returns true for team_leader", () => {
    expect(isTeamLeaderOrAdmin(makeUser("team_leader"))).toBe(true);
  });

  it("returns false for regular user", () => {
    expect(isTeamLeaderOrAdmin(makeUser("user"))).toBe(false);
  });
});
