"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Clock, CalendarDays, Users, CheckCircle, XCircle, Home } from "lucide-react";
import type {
  User,
  DailySummary,
  Absence,
  VacationBalance,
  DayAvailability,
  TeamMember,
} from "@/lib/auth";
import {
  fetchMe,
  fetchMyTeams,
  fetchTeamAvailability,
  fetchAbsences,
  fetchVacationBalance,
  fetchTimeEntrySummary,
  fetchPendingAbsences,
} from "@/lib/api";
import { isTeamLeaderOrAdmin } from "@/lib/rbac";

// ── Helpers ──────────────────────────────────────────────────────────────────

function greeting(): string {
  const h = new Date().getHours();
  if (h >= 5 && h < 11) return "Guten Morgen";
  if (h >= 11 && h < 18) return "Guten Tag";
  return "Guten Abend";
}

function todayLabel(): string {
  return new Date().toLocaleDateString("de-DE", {
    weekday: "long",
    day: "numeric",
    month: "long",
  });
}

function getMonday(d: Date): Date {
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  const monday = new Date(d);
  monday.setDate(diff);
  monday.setHours(0, 0, 0, 0);
  return monday;
}

function toDateString(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function workingDaysBetween(startStr: string, endStr: string): number {
  let count = 0;
  const end = new Date(endStr);
  for (const d = new Date(startStr); d <= end; d.setDate(d.getDate() + 1)) {
    const wd = d.getDay();
    if (wd !== 0 && wd !== 6) count++;
  }
  return count;
}

function addDays(d: Date, n: number): Date {
  const r = new Date(d);
  r.setDate(r.getDate() + n);
  return r;
}

function formatMinutes(m: number): string {
  const h = Math.floor(m / 60);
  const min = m % 60;
  return `${h}:${min.toString().padStart(2, "0")}`;
}

function initials(name: string): string {
  return name
    .split(" ")
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}

function avatarColor(name: string): string {
  const colors = [
    "bg-cyan-700", "bg-violet-700", "bg-emerald-700",
    "bg-rose-700", "bg-amber-700", "bg-sky-700",
  ];
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) & 0xffff;
  return colors[hash % colors.length];
}

function roleLabel(role: string): string {
  if (role === "admin") return "Administrator";
  if (role === "team_leader") return "Teamleiter";
  return "Mitarbeiter";
}

function absenceStatusLabel(status: string): string {
  if (status === "pending") return "Ausstehend";
  if (status === "approved") return "Genehmigt";
  if (status === "rejected") return "Abgelehnt";
  return status;
}

// ── Sub-components ────────────────────────────────────────────────────────────

function ProfileCard({ user }: { user: User }) {
  return (
    <Card className="glass-card">
      <CardContent className="pt-5 pb-4">
        <div className="flex items-start gap-4">
          <div
            className={`flex size-14 shrink-0 items-center justify-center rounded-xl text-lg font-bold text-white ${avatarColor(user.display_name)}`}
          >
            {initials(user.display_name)}
          </div>
          <div className="min-w-0">
            <p className="font-semibold font-heading truncate">{user.display_name}</p>
            <p className="text-sm text-muted-foreground truncate">{roleLabel(user.global_role)}</p>
          </div>
        </div>
        <div className="mt-4 flex gap-2">
          <Link
            href="/time"
            className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-border/50 px-3 py-2 text-xs text-muted-foreground hover:text-foreground hover:bg-muted/30 transition-colors"
          >
            <Clock className="size-3.5" />
            Zeiterfassung
          </Link>
          <Link
            href="/absences"
            className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-border/50 px-3 py-2 text-xs text-muted-foreground hover:text-foreground hover:bg-muted/30 transition-colors"
          >
            <CalendarDays className="size-3.5" />
            Abwesenheit
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}

function VacationCard({
  balance,
  upcomingAbsences,
}: {
  balance: VacationBalance | null;
  upcomingAbsences: Absence[];
}) {
  return (
    <Card className="glass-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium flex items-center gap-2">
          <CalendarDays className="size-4 text-muted-foreground" />
          Abwesenheit
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-3">
        {balance && (
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold font-heading">{balance.remaining_days}</span>
            <span className="text-sm text-muted-foreground">Tage Resturlaub</span>
          </div>
        )}
        {upcomingAbsences.length > 0 ? (
          <div className="space-y-1.5">
            {upcomingAbsences.slice(0, 3).map((a) => {
              const start = new Date(a.start_date);
              const end = new Date(a.end_date);
              const sameDay = a.start_date === a.end_date;
              return (
                <div key={a.id} className="flex items-center justify-between text-sm">
                  <span className="text-foreground truncate max-w-[140px]">
                    {sameDay
                      ? start.toLocaleDateString("de-DE", { day: "numeric", month: "short" })
                      : `${start.toLocaleDateString("de-DE", { day: "numeric", month: "short" })} – ${end.toLocaleDateString("de-DE", { day: "numeric", month: "short" })}`}
                  </span>
                  <Badge
                    variant={a.status === "approved" ? "default" : "secondary"}
                    className="text-xs"
                  >
                    {absenceStatusLabel(a.status)}
                  </Badge>
                </div>
              );
            })}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">Keine bevorstehenden Abwesenheiten</p>
        )}
        <Link
          href="/absences"
          className="block text-xs text-primary hover:underline"
        >
          Alle anzeigen →
        </Link>
      </CardContent>
    </Card>
  );
}

function TeamTodayCard({ availability }: { availability: DayAvailability[] }) {
  return (
    <Card className="glass-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium flex items-center gap-2">
          <Users className="size-4 text-muted-foreground" />
          Team heute
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-2">
        {availability.length === 0 ? (
          <p className="text-xs text-muted-foreground">Kein Team zugewiesen</p>
        ) : (
          availability.map((a) => (
            <div key={a.user_id} className="flex items-center gap-2.5">
              <div
                className={`flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${avatarColor(a.display_name)}`}
              >
                {initials(a.display_name)}
              </div>
              <span className="flex-1 text-sm truncate">{a.display_name}</span>
              {a.status === "present" && (
                <CheckCircle className="size-4 text-emerald-400 shrink-0" />
              )}
              {a.status === "homeoffice" && (
                <Home className="size-4 text-cyan-400 shrink-0" />
              )}
              {a.status === "absent" && (
                <XCircle className="size-4 text-red-400 shrink-0" />
              )}
              {a.status === "no_entry" && (
                <span className="text-xs text-muted-foreground shrink-0">–</span>
              )}
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function InboxCard({
  pendingAbsences,
  isLeader,
}: {
  pendingAbsences: Absence[];
  isLeader: boolean;
}) {
  return (
    <Card className="glass-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium">Inbox-Highlights</CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-2">
        {pendingAbsences.length === 0 ? (
          <p className="text-xs text-muted-foreground">Keine offenen Eintraege</p>
        ) : (
          pendingAbsences.slice(0, 5).map((a) => {
            const start = new Date(a.start_date);
            const end = new Date(a.end_date);
            const diff = workingDaysBetween(a.start_date, a.end_date);
            return (
              <div
                key={a.id}
                className="flex items-start gap-3 rounded-lg border border-border/40 p-2.5"
              >
                <div className={`size-2 mt-1.5 rounded-full shrink-0 ${isLeader ? "bg-amber-400" : "bg-cyan-400"}`} />
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">
                    {isLeader ? "Abwesenheitsantrag" : "Eigener Antrag"}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {diff} {diff === 1 ? "Tag" : "Tage"} – vom{" "}
                    {start.toLocaleDateString("de-DE", { day: "numeric", month: "short" })} bis{" "}
                    {end.toLocaleDateString("de-DE", { day: "numeric", month: "short", year: "numeric" })}
                  </p>
                </div>
              </div>
            );
          })
        )}
        {pendingAbsences.length > 0 && (
          <Link href="/absences" className="block text-xs text-primary hover:underline">
            {isLeader ? "Antraege pruefen →" : "Details anzeigen →"}
          </Link>
        )}
      </CardContent>
    </Card>
  );
}

function WeekCard({ summary }: { summary: DailySummary[] }) {
  const weekStart = getMonday(new Date());
  const days = Array.from({ length: 5 }, (_, i) => {
    const d = addDays(weekStart, i);
    const ds = toDateString(d);
    const dayData = summary.find((s) => s.date.slice(0, 10) === ds);
    const isToday = ds === toDateString(new Date());
    return { date: d, ds, dayData, isToday };
  });

  return (
    <Card className="glass-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium flex items-center justify-between">
          <span className="flex items-center gap-2">
            <Clock className="size-4 text-muted-foreground" />
            Diese Woche
          </span>
          <Badge variant="secondary">
            {formatMinutes(summary.reduce((s, d) => s + d.total_minutes, 0))} h
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-1">
        {days.map(({ date, ds, dayData, isToday }) => (
          <div
            key={ds}
            className={`flex items-center gap-3 py-1.5 px-2 rounded-lg ${isToday ? "bg-primary/5 border border-primary/20" : ""}`}
          >
            <span className="w-8 text-xs text-muted-foreground font-medium">
              {date.toLocaleDateString("de-DE", { weekday: "short" })}
            </span>
            <span className="text-xs text-muted-foreground w-14">
              {date.toLocaleDateString("de-DE", { day: "2-digit", month: "2-digit" })}
            </span>
            {dayData ? (
              <span className="text-sm text-primary font-medium ml-auto">
                {formatMinutes(dayData.total_minutes)} h
              </span>
            ) : (
              <span className="text-xs text-muted-foreground ml-auto">–</span>
            )}
            {isToday && <Badge className="text-xs">Heute</Badge>}
          </div>
        ))}
        <Link href="/time" className="block text-xs text-primary hover:underline pt-1">
          Zeiterfassung offnen →
        </Link>
      </CardContent>
    </Card>
  );
}

// ── Page ─────────────────────────────────────────────────────────────────────

export default function StartseiteePage() {
  const [user, setUser] = useState<User | null>(null);
  const [weekSummary, setWeekSummary] = useState<DailySummary[]>([]);
  const [pendingAbsences, setPendingAbsences] = useState<Absence[]>([]);
  const [upcomingAbsences, setUpcomingAbsences] = useState<Absence[]>([]);
  const [vacationBalance, setVacationBalance] = useState<VacationBalance | null>(null);
  const [availability, setAvailability] = useState<DayAvailability[]>([]);

  useEffect(() => {
    fetchMe().then(setUser).catch(() => {});
    fetchVacationBalance().then(setVacationBalance).catch(() => {});

    // Week summary
    const weekStart = getMonday(new Date());
    const weekEnd = addDays(weekStart, 6);
    fetchTimeEntrySummary(toDateString(weekStart), toDateString(weekEnd))
      .then((s) => setWeekSummary(s ?? []))
      .catch(() => {});

    // Upcoming absences (next 90 days)
    const today = toDateString(new Date());
    const future = toDateString(addDays(new Date(), 90));
    fetchAbsences(today, future)
      .then((a) => setUpcomingAbsences((a ?? []).filter((x) => x.status !== "rejected" && x.status !== "cancelled")))
      .catch(() => {});
  }, []);

  // Load team data once we have the user
  useEffect(() => {
    if (!user) return;

    // Team availability for today
    fetchMyTeams()
      .then(async (memberships: TeamMember[]) => {
        if (memberships.length === 0) return;
        const firstTeamId = memberships[0].team_id;
        const today = toDateString(new Date());
        const avail = await fetchTeamAvailability(firstTeamId, today, today);
        setAvailability(avail ?? []);
      })
      .catch(() => {});

    // Pending absences (own or team)
    if (isTeamLeaderOrAdmin(user)) {
      fetchMyTeams()
        .then(async (memberships: TeamMember[]) => {
          if (memberships.length === 0) return;
          const firstTeamId = memberships[0].team_id;
          const pending = await fetchPendingAbsences(firstTeamId);
          setPendingAbsences(pending ?? []);
        })
        .catch(() => {});
    } else {
      // Own pending absences
      const today = toDateString(new Date());
      const future = toDateString(addDays(new Date(), 365));
      fetchAbsences(today, future)
        .then((a) => setPendingAbsences((a ?? []).filter((x) => x.status === "pending")))
        .catch(() => {});
    }
  }, [user]);

  const isLeader = isTeamLeaderOrAdmin(user);

  return (
    <div className="flex flex-col gap-6">
      {/* Header greeting */}
      <div>
        <p className="text-sm text-muted-foreground">{todayLabel()}</p>
        <h1 className="font-heading text-3xl font-bold tracking-tight mt-0.5">
          {greeting()}, {user?.display_name.split(" ")[0] ?? "…"}
        </h1>
      </div>

      {/* Main grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
        {/* Left column */}
        <div className="flex flex-col gap-6">
          <InboxCard pendingAbsences={pendingAbsences} isLeader={isLeader} />
          <WeekCard summary={weekSummary} />
        </div>

        {/* Right column */}
        <div className="flex flex-col gap-4">
          {user && <ProfileCard user={user} />}
          <VacationCard balance={vacationBalance} upcomingAbsences={upcomingAbsences} />
          <TeamTodayCard availability={availability} />
        </div>
      </div>
    </div>
  );
}
