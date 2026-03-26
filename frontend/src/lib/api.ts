import type {
  User,
  UserRole,
  Team,
  TeamMember,
  Project,
  TimeEntry,
  TimeEntryResponse,
  DailySummary,
  Absence,
  AbsenceType,
  VacationBalance,
  OvertimeSummary,
  DayAvailability,
  WorkSchedule,
  DashboardStats,
} from "@/lib/auth";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || `Request failed: ${res.status}`);
  }
  return res.json() as Promise<T>;
}

// ── Current User ──

export function fetchMe(): Promise<import("@/lib/auth").User> {
  return apiFetch("/auth/me");
}

// ── Admin: Users ──

export function fetchUsers(): Promise<User[]> {
  return apiFetch("/admin/users");
}

export function fetchUser(id: string): Promise<User> {
  return apiFetch(`/admin/users/${id}`);
}

export function updateUserRole(id: string, role: UserRole): Promise<void> {
  return apiFetch(`/admin/users/${id}/role`, {
    method: "PUT",
    body: JSON.stringify({ role }),
  });
}

export function updateUserActive(id: string, active: boolean): Promise<void> {
  return apiFetch(`/admin/users/${id}/active`, {
    method: "PUT",
    body: JSON.stringify({ active }),
  });
}

// ── Teams ──

export function fetchTeams(): Promise<Team[]> {
  return apiFetch("/teams");
}

export function fetchMyTeams(): Promise<TeamMember[]> {
  return apiFetch("/teams/my");
}

export function createTeam(name: string, description: string): Promise<Team> {
  return apiFetch("/admin/teams", {
    method: "POST",
    body: JSON.stringify({ name, description }),
  });
}

export function updateTeam(id: string, name: string, description: string): Promise<Team> {
  return apiFetch(`/admin/teams/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name, description }),
  });
}

export function deleteTeam(id: string): Promise<void> {
  return apiFetch(`/admin/teams/${id}`, { method: "DELETE" });
}

// ── Team Members ──

export function fetchTeamMembers(teamId: string): Promise<TeamMember[]> {
  return apiFetch(`/teams/${teamId}/members`);
}

export function addTeamMember(teamId: string, userId: string, role: UserRole): Promise<void> {
  return apiFetch(`/teams/${teamId}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId, role }),
  });
}

export function removeTeamMember(teamId: string, userId: string): Promise<void> {
  return apiFetch(`/teams/${teamId}/members/${userId}`, { method: "DELETE" });
}

// ── Time Entries ──

export function createTimeEntry(data: {
  entry_date: string;
  start_time: string;
  end_time: string;
  break_minutes: number;
  project_id?: string;
  description?: string;
}): Promise<TimeEntryResponse> {
  return apiFetch("/time-entries", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function fetchTimeEntries(from: string, to: string): Promise<TimeEntry[]> {
  return apiFetch(`/time-entries?from=${from}&to=${to}`);
}

export function updateTimeEntry(
  id: string,
  data: {
    entry_date: string;
    start_time: string;
    end_time: string;
    break_minutes: number;
    project_id?: string;
    description?: string;
  },
): Promise<TimeEntryResponse> {
  return apiFetch(`/time-entries/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export function deleteTimeEntry(id: string): Promise<void> {
  return apiFetch(`/time-entries/${id}`, { method: "DELETE" });
}

export function fetchTimeEntrySummary(from: string, to: string): Promise<DailySummary[]> {
  return apiFetch(`/time-entries/summary?from=${from}&to=${to}`);
}

export function fetchTeamTimeEntries(
  teamId: string,
  from: string,
  to: string,
): Promise<TimeEntry[]> {
  return apiFetch(`/time-entries/team/${teamId}?from=${from}&to=${to}`);
}

// ── Projects ──

export function fetchProjects(): Promise<Project[]> {
  return apiFetch("/projects");
}

export function fetchAllProjects(): Promise<Project[]> {
  return apiFetch("/admin/projects");
}

export function createProject(name: string, customerName: string): Promise<Project> {
  return apiFetch("/admin/projects", {
    method: "POST",
    body: JSON.stringify({ name, customer_name: customerName }),
  });
}

export function updateProject(
  id: string,
  name: string,
  customerName: string,
  isActive: boolean,
): Promise<Project> {
  return apiFetch(`/admin/projects/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name, customer_name: customerName, is_active: isActive }),
  });
}

// ── Absences ──

export function createAbsence(data: {
  absence_type_id: string;
  start_date: string;
  end_date: string;
  note?: string;
}): Promise<Absence> {
  return apiFetch("/absences", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function fetchAbsences(from: string, to: string): Promise<Absence[]> {
  return apiFetch(`/absences?from=${from}&to=${to}`);
}

export function updateAbsence(
  id: string,
  data: {
    absence_type_id: string;
    start_date: string;
    end_date: string;
    note?: string;
  },
): Promise<Absence> {
  return apiFetch(`/absences/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export function deleteAbsence(id: string): Promise<void> {
  return apiFetch(`/absences/${id}`, { method: "DELETE" });
}

export function cancelAbsence(id: string): Promise<Absence> {
  return apiFetch(`/absences/${id}/cancel`, { method: "POST" });
}

export function fetchTeamAbsences(
  teamId: string,
  from: string,
  to: string,
): Promise<Absence[]> {
  return apiFetch(`/absences/team/${teamId}?from=${from}&to=${to}`);
}

export function fetchPendingAbsences(teamId: string): Promise<Absence[]> {
  return apiFetch(`/absences/team/${teamId}/pending`);
}

export function reviewAbsence(
  id: string,
  approve: boolean,
  reviewNote?: string,
): Promise<Absence> {
  return apiFetch(`/absences/${id}/review`, {
    method: "PUT",
    body: JSON.stringify({ approve, review_note: reviewNote || "" }),
  });
}

export function fetchVacationBalance(year?: number): Promise<VacationBalance> {
  const params = year ? `?year=${year}` : "";
  return apiFetch(`/absences/balance${params}`);
}

// ── Absence Types ──

export function fetchAbsenceTypes(): Promise<AbsenceType[]> {
  return apiFetch("/absence-types");
}

export function fetchAllAbsenceTypes(): Promise<AbsenceType[]> {
  return apiFetch("/admin/absence-types");
}

export function updateAbsenceType(
  id: string,
  data: {
    name: string;
    color: string;
    requires_approval: boolean;
    counts_as_work: boolean;
    is_active: boolean;
    sort_order: number;
  },
): Promise<AbsenceType> {
  return apiFetch(`/admin/absence-types/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

// ── Vacation Entitlements ──

export function upsertEntitlement(data: {
  user_id: string;
  year: number;
  total_days: number;
  carry_over_days: number;
}): Promise<unknown> {
  return apiFetch("/admin/entitlements", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

// ── Overtime & Dashboard ──

export function fetchOvertimeSummary(
  from?: string,
  to?: string,
): Promise<OvertimeSummary> {
  const params = from && to ? `?from=${from}&to=${to}` : "";
  return apiFetch(`/overtime${params}`);
}

export function fetchOvertimeTrend(): Promise<OvertimeSummary[]> {
  return apiFetch("/overtime/trend");
}

export function fetchDashboardStats(): Promise<DashboardStats> {
  return apiFetch("/dashboard");
}

export function fetchTeamAvailability(
  teamId: string,
  from: string,
  to: string,
): Promise<DayAvailability[]> {
  return apiFetch(`/teams/${teamId}/availability?from=${from}&to=${to}`);
}

export function fetchWorkSchedule(): Promise<WorkSchedule[]> {
  return apiFetch("/work-schedule");
}

// ── Stempeluhr ──

export interface ActiveStamp {
  user_id: string;
  started_at: string;
  break_start?: string;
  break_minutes: number;
  project_id?: string;
  description: string;
}

export function fetchActiveStamp(): Promise<ActiveStamp | null> {
  return apiFetch<ActiveStamp | null>("/stamp/active");
}

export function stampIn(data?: { project_id?: string; description?: string }): Promise<ActiveStamp> {
  return apiFetch("/stamp/in", {
    method: "POST",
    body: JSON.stringify(data ?? {}),
  });
}

export function stampOut(): Promise<{ entry: unknown; warnings?: unknown[] }> {
  return apiFetch("/stamp/out", { method: "POST" });
}

export function toggleBreak(): Promise<ActiveStamp> {
  return apiFetch("/stamp/break", { method: "POST" });
}

export function upsertWorkSchedule(data: {
  user_id: string;
  valid_from: string;
  weekly_hours: number;
  monday_hours: number;
  tuesday_hours: number;
  wednesday_hours: number;
  thursday_hours: number;
  friday_hours: number;
  saturday_hours: number;
  sunday_hours: number;
}): Promise<WorkSchedule> {
  return apiFetch("/admin/work-schedules", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}
