const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

export interface User {
  id: string;
  email: string;
  display_name: string;
  is_active: boolean;
  created_at: string;
  last_login_at?: string;
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

    const res = await fetch(`${API_URL}/auth/me`, {
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
