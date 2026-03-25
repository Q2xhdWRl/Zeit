import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { fetchCurrentUser } from "@/lib/auth";

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
      {/* Sidebar placeholder — will be built in Phase 3 */}
      <aside className="hidden w-64 border-r border-border bg-sidebar lg:block">
        <div className="flex h-16 items-center gap-2 border-b border-sidebar-border px-6">
          <span className="font-heading text-lg font-bold text-glow-cyan text-primary">
            Zeit
          </span>
        </div>
        <nav className="flex flex-col gap-1 p-4">
          <span className="px-3 py-2 text-sm text-sidebar-foreground">
            {user.display_name}
          </span>
        </nav>
      </aside>

      {/* Main content */}
      <main className="flex-1">
        <header className="flex h-16 items-center justify-between border-b border-border px-6">
          <h2 className="font-heading text-lg font-semibold">Dashboard</h2>
          <span className="text-sm text-muted-foreground">
            {user.email}
          </span>
        </header>
        <div className="p-6">{children}</div>
      </main>
    </div>
  );
}
