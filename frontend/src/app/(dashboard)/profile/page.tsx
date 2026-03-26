"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { UserCircle, CalendarDays, Clock } from "lucide-react";
import type { User, WorkSchedule, VacationBalance } from "@/lib/auth";
import { fetchMe, fetchWorkSchedule, fetchVacationBalance } from "@/lib/api";

const roleLabels: Record<string, string> = {
  admin: "Admin",
  team_leader: "Teamleiter",
  user: "Benutzer",
};

const roleVariants: Record<string, "default" | "secondary" | "destructive"> = {
  admin: "destructive",
  team_leader: "default",
  user: "secondary",
};

const DAY_LABELS = [
  { key: "monday_hours" as const, label: "Mo" },
  { key: "tuesday_hours" as const, label: "Di" },
  { key: "wednesday_hours" as const, label: "Mi" },
  { key: "thursday_hours" as const, label: "Do" },
  { key: "friday_hours" as const, label: "Fr" },
  { key: "saturday_hours" as const, label: "Sa" },
  { key: "sunday_hours" as const, label: "So" },
];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("de-DE", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

function initials(name: string): string {
  return name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

export default function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [schedule, setSchedule] = useState<WorkSchedule | null>(null);
  const [balance, setBalance] = useState<VacationBalance | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      fetchMe(),
      fetchWorkSchedule(),
      fetchVacationBalance(),
    ])
      .then(([u, schedules, b]) => {
        setUser(u);
        // Most recent schedule (API returns ordered by valid_from DESC)
        setSchedule(schedules.length > 0 ? schedules[0] : null);
        setBalance(b);
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : "Fehler beim Laden"),
      );
  }, []);

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
        {error}
      </div>
    );
  }

  if (!user) {
    return (
      <div className="text-sm text-muted-foreground">Laden…</div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-2xl font-bold tracking-tight">
        Mein Profil
      </h1>

      {/* Identity card */}
      <Card className="glass-card">
        <CardContent className="flex items-center gap-5 pt-6">
          <div className="flex size-16 shrink-0 items-center justify-center rounded-full bg-primary/20 text-xl font-bold text-primary font-heading">
            {initials(user.display_name)}
          </div>
          <div className="min-w-0">
            <p className="text-lg font-semibold font-heading truncate">
              {user.display_name}
            </p>
            <p className="text-sm text-muted-foreground truncate">{user.email}</p>
            <div className="mt-2 flex items-center gap-2">
              <Badge variant={roleVariants[user.global_role] ?? "secondary"}>
                {roleLabels[user.global_role] ?? user.global_role}
              </Badge>
              {!user.is_active && (
                <Badge variant="destructive">Inaktiv</Badge>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Work schedule */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center gap-2 pb-3">
            <Clock className="size-4 text-muted-foreground" />
            <CardTitle className="text-base font-heading">Arbeitszeitmodell</CardTitle>
          </CardHeader>
          <CardContent>
            {schedule ? (
              <>
                <p className="mb-4 text-sm text-muted-foreground">
                  Gültig ab {formatDate(schedule.valid_from)} · {schedule.weekly_hours} h/Woche
                </p>
                <div className="grid grid-cols-7 gap-1 text-center text-xs">
                  {DAY_LABELS.map(({ key, label }) => (
                    <div key={key}>
                      <div className="mb-1 font-medium text-muted-foreground">{label}</div>
                      <div
                        className={`rounded py-1 font-mono ${
                          schedule[key] > 0
                            ? "bg-primary/10 text-primary"
                            : "bg-muted/30 text-muted-foreground"
                        }`}
                      >
                        {schedule[key] > 0 ? `${schedule[key]}h` : "–"}
                      </div>
                    </div>
                  ))}
                </div>
              </>
            ) : (
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Kein individuelles Modell hinterlegt.
                </p>
                <div className="grid grid-cols-7 gap-1 text-center text-xs">
                  {DAY_LABELS.map(({ key, label }) => {
                    const defaultHours = ["Sa", "So"].includes(label) ? 0 : 8;
                    return (
                      <div key={key}>
                        <div className="mb-1 font-medium text-muted-foreground">{label}</div>
                        <div
                          className={`rounded py-1 font-mono ${
                            defaultHours > 0
                              ? "bg-muted/20 text-muted-foreground"
                              : "bg-muted/10 text-muted-foreground/50"
                          }`}
                        >
                          {defaultHours > 0 ? `${defaultHours}h` : "–"}
                        </div>
                      </div>
                    );
                  })}
                </div>
                <p className="text-xs text-muted-foreground">
                  Standardmodell: 40 h/Woche (Mo–Fr je 8 h)
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Vacation balance */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center gap-2 pb-3">
            <CalendarDays className="size-4 text-muted-foreground" />
            <CardTitle className="text-base font-heading">
              Urlaubskonto {balance?.year ?? new Date().getFullYear()}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {balance ? (
              <div className="grid grid-cols-2 gap-3">
                <div className="rounded-lg bg-white/5 p-3 text-center">
                  <div className="text-2xl font-bold font-heading">
                    {balance.total_days + balance.carry_over_days}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">Anspruch gesamt</div>
                </div>
                <div className="rounded-lg bg-emerald-500/10 p-3 text-center">
                  <div className="text-2xl font-bold font-heading text-emerald-400">
                    {balance.remaining_days}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">Verbleibend</div>
                </div>
                <div className="rounded-lg bg-white/5 p-3 text-center">
                  <div className="text-2xl font-bold font-heading">
                    {balance.used_days}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">Genommen</div>
                </div>
                <div className="rounded-lg bg-yellow-500/10 p-3 text-center">
                  <div className="text-2xl font-bold font-heading text-yellow-400">
                    {balance.pending_days}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">Ausstehend</div>
                </div>
                {balance.carry_over_days > 0 && (
                  <div className="col-span-2 rounded-lg bg-cyan-500/10 p-2 text-center text-xs text-cyan-400">
                    Davon {balance.carry_over_days} Tage Übertrag aus dem Vorjahr
                  </div>
                )}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                Kein Urlaubsanspruch konfiguriert.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Account details */}
      <Card className="glass-card">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base font-heading">
            <UserCircle className="size-4 text-muted-foreground" />
            Kontoinformationen
          </CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 gap-3 sm:grid-cols-2 text-sm">
            <div>
              <dt className="text-muted-foreground">Konto angelegt</dt>
              <dd className="font-medium">{formatDate(user.created_at)}</dd>
            </div>
            {user.last_login_at && (
              <div>
                <dt className="text-muted-foreground">Letzter Login</dt>
                <dd className="font-medium">
                  {new Date(user.last_login_at).toLocaleString("de-DE", {
                    day: "2-digit",
                    month: "2-digit",
                    year: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>
    </div>
  );
}
