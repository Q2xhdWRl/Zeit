// NEXT_PUBLIC_API_URL is used by the browser (client-side).
// BACKEND_URL is used for server-side fetches inside the container (e.g. layout auth check).
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

export type UserRole = "admin" | "team_leader" | "user";

export interface User {
  id: string;
  email: string;
  display_name: string;
  global_role: UserRole;
  is_active: boolean;
  created_at: string;
  last_login_at?: string;
}

export interface Team {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  user_id: string;
  team_id: string;
  role: UserRole;
  joined_at: string;
  display_name?: string;
  email?: string;
}

export interface Project {
  id: string;
  name: string;
  customer_name: string;
  is_active: boolean;
  created_at: string;
}

export interface TimeEntry {
  id: string;
  user_id: string;
  entry_date: string;
  start_time: string;
  end_time: string;
  break_minutes: number;
  project_id?: string;
  description: string;
  insert_time: string;
  updated_at: string;
}

export interface ArbZGViolation {
  rule: string;
  message: string;
}

export interface TimeEntryResponse {
  entry: TimeEntry;
  warnings?: ArbZGViolation[];
}

export interface DailySummary {
  date: string;
  total_minutes: number;
  entry_count: number;
}

export type AbsenceStatus = "pending" | "approved" | "rejected" | "cancelled";

export interface AbsenceType {
  id: string;
  name: string;
  color: string;
  requires_approval: boolean;
  counts_as_work: boolean;
  is_active: boolean;
  sort_order: number;
  created_at: string;
}

export interface Absence {
  id: string;
  user_id: string;
  absence_type_id: string;
  start_date: string;
  end_date: string;
  note: string;
  status: AbsenceStatus;
  reviewed_by?: string;
  reviewed_at?: string;
  review_note: string;
  created_at: string;
  updated_at: string;
}

export interface VacationBalance {
  year: number;
  total_days: number;
  carry_over_days: number;
  used_days: number;
  pending_days: number;
  remaining_days: number;
}

export interface OvertimeSummary {
  period_from: string;
  period_to: string;
  target_minutes: number;
  actual_minutes: number;
  diff_minutes: number;
}

export interface DayAvailability {
  user_id: string;
  display_name: string;
  date: string;
  status: "present" | "absent" | "homeoffice" | "no_entry";
  absence_type?: string;
  work_minutes: number;
}

export interface WorkSchedule {
  id: string;
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
  created_at: string;
  updated_at: string;
}

export interface DashboardStats {
  today_minutes: number;
  week_minutes: number;
  month_overtime: OvertimeSummary;
  vacation_balance?: VacationBalance;
  team_present_count: number;
  team_total_count: number;
}

/** Fetches the current authenticated user. Returns null if not authenticated. */
export async function fetchCurrentUser(
  cookieHeader?: string,
): Promise<User | null> {
  try {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (cookieHeader) {
      headers["Cookie"] = cookieHeader;
    }

    // Use BACKEND_URL for server-side calls (container-internal), fall back to API_URL for local dev.
    const serverUrl = process.env.BACKEND_URL || API_URL;
    const res = await fetch(`${serverUrl}/auth/me`, {
      credentials: "include",
      headers,
      cache: "no-store",
    });

    if (!res.ok) {
      return null;
    }

    return (await res.json()) as User;
  } catch {
    return null;
  }
}

/** Returns the login URL for initiating Microsoft SSO. */
export function getLoginUrl(): string {
  return `${API_URL}/auth/login`;
}

/** Returns the logout URL. */
export function getLogoutUrl(): string {
  return `${API_URL}/auth/logout`;
}
