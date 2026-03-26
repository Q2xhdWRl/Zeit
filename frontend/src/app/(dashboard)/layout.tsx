import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { fetchCurrentUser } from "@/lib/auth";
import Sidebar from "@/components/sidebar";
import { GlowProvider } from "@/components/glow-provider";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const sessionCookie = cookieStore.get("zeit_session");

  if (!sessionCookie?.value) {
    redirect("/login");
  }

  const user = await fetchCurrentUser(
    `zeit_session=${sessionCookie.value}`,
  );

  if (!user) {
    redirect("/login");
  }

  return (
    <div className="flex min-h-screen">
      <GlowProvider />
      <Sidebar user={user} />

      <main className="flex-1 pt-14 lg:pt-0">
        <header className="hidden lg:flex h-16 items-center justify-between border-b border-border px-6">
          <h2 className="font-heading text-lg font-semibold">NEWA Zeiterfassung</h2>
          <div className="flex items-center gap-3">
            <span className="text-xs rounded-full border border-primary/20 bg-primary/5 px-2 py-0.5 text-primary">
              {user.global_role === "admin"
                ? "Admin"
                : user.global_role === "team_leader"
                  ? "Teamleiter"
                  : "Benutzer"}
            </span>
            <span className="text-sm text-muted-foreground">
              {user.email}
            </span>
          </div>
        </header>
        <div className="p-6">{children}</div>
      </main>
    </div>
  );
}
