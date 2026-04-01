"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Users, ChevronLeft, ChevronRight } from "lucide-react";
import type { DayAvailability, TeamMember } from "@/lib/auth";
import { fetchTeamAvailability, fetchMyTeams } from "@/lib/api";

function toDateString(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function getMonday(d: Date): Date {
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  const monday = new Date(d);
  monday.setDate(diff);
  monday.setHours(0, 0, 0, 0);
  return monday;
}

function addDays(d: Date, n: number): Date {
  const result = new Date(d);
  result.setDate(result.getDate() + n);
  return result;
}

function isoWeek(d: Date): number {
  const tmp = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
  tmp.setUTCDate(tmp.getUTCDate() + 4 - (tmp.getUTCDay() || 7));
  const yearStart = new Date(Date.UTC(tmp.getUTCFullYear(), 0, 1));
  return Math.ceil(((tmp.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

function formatDayHeader(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("de-DE", { weekday: "short", day: "2-digit", month: "2-digit" });
}

function formatMinutes(m: number): string {
  const h = Math.floor(m / 60);
  const min = m % 60;
  return `${h}:${min.toString().padStart(2, "0")}`;
}

const statusConfig: Record<string, { label: string; color: string; bg: string }> = {
  present: { label: "Anwesend", color: "text-green-400", bg: "bg-green-500/20" },
  absent: { label: "Abwesend", color: "text-red-400", bg: "bg-red-500/20" },
  homeoffice: { label: "Homeoffice", color: "text-blue-400", bg: "bg-blue-500/20" },
  no_entry: { label: "Kein Eintrag", color: "text-gray-500", bg: "bg-gray-500/10" },
};

export default function TeamPage() {
  const [availability, setAvailability] = useState<DayAvailability[]>([]);
  const [teams, setTeams] = useState<TeamMember[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<string>("");
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()));
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const weekEnd = addDays(weekStart, 4); // Mo-Fr

  useEffect(() => {
    fetchMyTeams()
      .then((t) => {
        setTeams(t ?? []);
        if (t && t.length > 0 && !selectedTeam) {
          setSelectedTeam(t[0].team_id);
        }
      })
      .catch(() => setError("Teams konnten nicht geladen werden"));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const loadData = useCallback(async () => {
    if (!selectedTeam) return;
    try {
      const data = await fetchTeamAvailability(
        selectedTeam,
        toDateString(weekStart),
        toDateString(weekEnd),
      );
      setAvailability(data ?? []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    } finally {
      setLoading(false);
    }
  }, [selectedTeam, weekStart, weekEnd]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  function prevWeek() {
    setWeekStart(addDays(weekStart, -7));
  }
  function nextWeek() {
    setWeekStart(addDays(weekStart, 7));
  }
  function goToday() {
    setWeekStart(getMonday(new Date()));
  }

  // Get unique users and dates
  const userMap = new Map<string, string>();
  for (const a of availability) {
    if (!userMap.has(a.user_id)) {
      userMap.set(a.user_id, a.display_name);
    }
  }
  const users = Array.from(userMap.entries());

  const dates: string[] = [];
  for (let i = 0; i < 5; i++) {
    dates.push(toDateString(addDays(weekStart, i)));
  }

  // Build lookup
  const lookup = new Map<string, DayAvailability>();
  for (const a of availability) {
    lookup.set(`${a.user_id}_${a.date}`, a);
  }

  if (loading) {
    return (
      <div className="flex flex-col gap-6 animate-pulse">
        <div className="h-8 w-48 rounded bg-muted/40" />
        <div className="flex items-center gap-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-9 w-20 rounded bg-muted/30" />
          ))}
        </div>
        <div className="h-64 rounded-xl bg-muted/20" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Users className="size-6 text-primary" />
          <h1 className="font-heading text-2xl font-bold tracking-tight">
            Teamuebersicht
          </h1>
        </div>
        {teams.length > 1 && (
          <select
            value={selectedTeam}
            onChange={(e) => setSelectedTeam(e.target.value)}
            className="rounded border border-border bg-background px-3 py-2 text-sm"
            aria-label="Team auswaehlen"
          >
            {teams.map((t) => (
              <option key={t.team_id} value={t.team_id}>
                {t.display_name || `Team ${t.team_id.slice(0, 8)}`}
              </option>
            ))}
          </select>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Week Navigation */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={prevWeek}>
            <ChevronLeft className="size-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={goToday}>
            Heute
          </Button>
          <Button variant="outline" size="sm" onClick={nextWeek}>
            <ChevronRight className="size-4" />
          </Button>
        </div>
        <span className="text-sm text-muted-foreground">
          KW {isoWeek(weekStart)}
        </span>
        <div className="flex gap-3 text-xs">
          {Object.entries(statusConfig).map(([key, cfg]) => (
            <span key={key} className={`flex items-center gap-1 ${cfg.color}`}>
              <span className={`inline-block h-2.5 w-2.5 rounded-full ${cfg.bg}`} />
              {cfg.label}
            </span>
          ))}
        </div>
      </div>

      {/* Availability Grid */}
      {users.length === 0 ? (
        <Card className="glass-card">
          <CardContent className="py-8 text-center">
            <p className="text-sm text-muted-foreground">
              Keine Teammitglieder gefunden
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card className="glass-card">
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                      Mitarbeiter
                    </th>
                    {dates.map((d) => (
                      <th
                        key={d}
                        className={`px-4 py-3 text-center font-medium text-muted-foreground ${d === toDateString(new Date()) ? "bg-primary/5" : ""}`}
                      >
                        {formatDayHeader(d)}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {users.map(([userId, displayName]) => (
                    <tr key={userId} className="border-b border-border/50">
                      <td className="px-4 py-3 font-medium">{displayName}</td>
                      {dates.map((d) => {
                        const entry = lookup.get(`${userId}_${d}`);
                        const status = entry?.status || "no_entry";
                        const cfg = statusConfig[status];
                        return (
                          <td
                            key={d}
                            className={`px-4 py-3 text-center ${d === toDateString(new Date()) ? "bg-primary/5" : ""}`}
                          >
                            <div className="flex flex-col items-center gap-1">
                              <span
                                className={`inline-block h-3 w-3 rounded-full ${cfg.bg}`}
                                title={cfg.label}
                              />
                              {entry?.absence_type && (
                                <span className="text-xs text-muted-foreground">
                                  {entry.absence_type}
                                </span>
                              )}
                              {status === "present" && entry?.work_minutes ? (
                                <span className="text-xs text-green-400">
                                  {formatMinutes(entry.work_minutes)}
                                </span>
                              ) : null}
                            </div>
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
