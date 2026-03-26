"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Clock,
  CalendarDays,
  Users,
  TrendingUp,
  LayoutDashboard,
  Shield,
  LogOut,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { User } from "@/lib/auth";
import { isAdmin } from "@/lib/rbac";
import { getLogoutUrl } from "@/lib/auth";

interface NavItem {
  href: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

const mainNav: NavItem[] = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/time", label: "Zeiterfassung", icon: Clock },
  { href: "/absences", label: "Abwesenheiten", icon: CalendarDays },
  { href: "/team", label: "Team", icon: Users },
  { href: "/overtime", label: "Ueberstunden", icon: TrendingUp },
];

const adminNav: NavItem[] = [
  { href: "/admin", label: "Verwaltung", icon: Shield },
];

function NavLink({ item, pathname }: { item: NavItem; pathname: string }) {
  const active = pathname === item.href || pathname.startsWith(item.href + "/");
  return (
    <Link
      href={item.href}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
        active
          ? "bg-primary/10 text-primary font-medium"
          : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
      )}
    >
      <item.icon className="size-4" />
      {item.label}
    </Link>
  );
}

export default function Sidebar({ user }: { user: User }) {
  const pathname = usePathname();

  return (
    <aside className="hidden w-64 border-r border-border bg-sidebar lg:flex lg:flex-col">
      <div className="flex h-16 items-center gap-2 border-b border-sidebar-border px-6">
        <span className="font-heading text-lg font-bold text-glow-cyan text-primary">
          Zeit
        </span>
      </div>

      <nav className="flex flex-1 flex-col gap-1 p-4">
        <div className="mb-2 px-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Navigation
        </div>
        {mainNav.map((item) => (
          <NavLink key={item.href} item={item} pathname={pathname} />
        ))}

        {isAdmin(user) && (
          <>
            <div className="mb-2 mt-6 px-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Administration
            </div>
            {adminNav.map((item) => (
              <NavLink key={item.href} item={item} pathname={pathname} />
            ))}
          </>
        )}
      </nav>

      <div className="border-t border-sidebar-border p-4">
        <div className="mb-3 px-3">
          <p className="text-sm font-medium text-sidebar-foreground truncate">
            {user.display_name}
          </p>
          <p className="text-xs text-muted-foreground truncate">{user.email}</p>
        </div>
        <a
          href={getLogoutUrl()}
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors"
        >
          <LogOut className="size-4" />
          Abmelden
        </a>
      </div>
    </aside>
  );
}
