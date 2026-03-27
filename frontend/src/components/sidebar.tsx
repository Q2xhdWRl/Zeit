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
  UserCircle,
  Menu,
  X,
} from "lucide-react";
import { useState } from "react";
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
  { href: "/dashboard", label: "Startseite", icon: LayoutDashboard },
  { href: "/time", label: "Zeiterfassung", icon: Clock },
  { href: "/absences", label: "Abwesenheiten", icon: CalendarDays },
  { href: "/team", label: "Team", icon: Users },
  { href: "/overtime", label: "Ueberstunden", icon: TrendingUp },
];

const adminNav: NavItem[] = [
  { href: "/admin", label: "Verwaltung", icon: Shield },
];

function NavLink({ item, pathname, onClick }: { item: NavItem; pathname: string; onClick?: () => void }) {
  const active = pathname === item.href || pathname.startsWith(item.href + "/");
  return (
    <Link
      href={item.href}
      onClick={onClick}
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

function SidebarContent({ user, pathname, onClose }: { user: User; pathname: string; onClose?: () => void }) {
  return (
    <>
      <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-6">
        <span className="font-heading text-lg font-bold text-glow-cyan text-primary">
          Zeit
        </span>
        {onClose && (
          <button
            onClick={onClose}
            aria-label="Sidebar schließen"
            className="lg:hidden rounded p-1 text-muted-foreground hover:text-foreground"
          >
            <X className="size-5" />
          </button>
        )}
      </div>

      <nav className="flex flex-1 flex-col gap-1 p-4">
        <div className="mb-2 px-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Navigation
        </div>
        {mainNav.map((item) => (
          <NavLink key={item.href} item={item} pathname={pathname} onClick={onClose} />
        ))}

        {isAdmin(user) && (
          <>
            <div className="mb-2 mt-6 px-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Administration
            </div>
            {adminNav.map((item) => (
              <NavLink key={item.href} item={item} pathname={pathname} onClick={onClose} />
            ))}
          </>
        )}
      </nav>

      <div className="border-t border-sidebar-border p-4">
        <Link
          href="/profile"
          onClick={onClose}
          className={cn(
            "flex items-center gap-3 rounded-lg px-3 py-2 mb-1 text-sm transition-colors",
            pathname === "/profile"
              ? "bg-primary/10 text-primary font-medium"
              : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
          )}
        >
          <UserCircle className="size-4" />
          <div className="min-w-0">
            <p className="font-medium truncate">{user.display_name}</p>
            <p className="text-xs text-muted-foreground truncate">{user.email}</p>
          </div>
        </Link>
        <a
          href={getLogoutUrl()}
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors"
        >
          <LogOut className="size-4" />
          Abmelden
        </a>
      </div>
    </>
  );
}

export default function Sidebar({ user }: { user: User }) {
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden w-64 border-r border-border sidebar-glass lg:flex lg:flex-col">
        <SidebarContent user={user} pathname={pathname} />
      </aside>

      {/* Mobile: hamburger button (rendered in layout header area via portal-like approach) */}
      <div className="lg:hidden fixed top-0 left-0 z-50 flex h-14 items-center px-4 border-b border-border sidebar-glass w-full">
        <button
          onClick={() => setMobileOpen(true)}
          aria-label="Menü öffnen"
          className="rounded p-1 text-muted-foreground hover:text-foreground"
        >
          <Menu className="size-5" />
        </button>
        <span className="ml-3 font-heading text-base font-bold text-primary">Zeit</span>
      </div>

      {/* Mobile drawer overlay */}
      {mobileOpen && (
        <div
          className="lg:hidden fixed inset-0 z-40 bg-black/60"
          onClick={() => setMobileOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Mobile drawer */}
      <aside
        className={cn(
          "lg:hidden fixed top-0 left-0 z-50 flex h-full w-64 flex-col border-r border-border sidebar-glass transition-transform duration-200",
          mobileOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <SidebarContent
          user={user}
          pathname={pathname}
          onClose={() => setMobileOpen(false)}
        />
      </aside>
    </>
  );
}
