import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api";
const isDev = process.env.NODE_ENV === "development";

export async function GET() {
  const cookieStore = await cookies();
  const sessionCookie = cookieStore.get("zeit_session");

  // In dev mode, skip backend logout to preserve seeded dev sessions.
  if (sessionCookie?.value && !isDev) {
    await fetch(`${API_URL}/auth/logout`, {
      method: "POST",
      headers: { Cookie: `zeit_session=${sessionCookie.value}` },
    }).catch(() => {});
  }

  cookieStore.delete("zeit_session");
  redirect("/");
}
