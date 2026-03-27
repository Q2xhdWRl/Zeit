"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { TrendingUp, ChevronLeft, ChevronRight } from "lucide-react";
import type { OvertimeSummary, TeamMember } from "@/lib/auth";
import {
  fetchOvertimeSummary,
  fetchOvertimeTrend,
  fetchMyTeams,
  fetchTeamOvertimeSummary,
  fetchAdminOvertimeSummary,
  fetchMe,
  type UserOvertimeSummary,
} from "@/lib/api";
import { isAdmin, isTeamLeaderOrAdmin } from "@/lib/rbac";

function formatMinutes(m: number): string {
  const sign = m < 0 ? "-" : "+";
  const abs = Math.abs(m);
  const h = Math.floor(abs / 60);
  const min = abs % 60;
  return `${sign}${h}:${min.toString().padStart(2, "0")}`;
}

function formatHours(m: number): string {
  const h = Math.floor(Math.abs(m) / 60);
  const min = Math.abs(m) % 60;
  return `${h}:${min.toString().padStart(2, "0")}`;
}

function toDateString(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function monthLabel(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("de-DE", { month: "long", year: "numeric" });
}

export default function OvertimePage() {
  const [month, setMonth] = useState(() => {
    const now = new Date();
    return new Date(now.getFullYear(), now.getMonth(), 1);
  });
  const [summary, setSummary] = useState<OvertimeSummary | null>(null);
  const [trend, setTrend] = useState<OvertimeSummary[]>([]);
  const [error, setError] = useState<string | null>(null);

  // Team / admin breakdown
  const [teamSummaries, setTeamSummaries] = useState<UserOvertimeSummary[]>([]);
  const [showTeam, setShowTeam] = useState(false);

  const monthStart = new Date(month.getFullYear(), month.getMonth(), 1);
  const monthEnd = new Date(month.getFullYear(), month.getMonth() + 1, 0);

  const loadSummary = useCallback(async () => {
    try {
      const data = await fetchOvertimeSummary(
        toDateString(monthStart),
        toDateString(monthEnd),
      );
      setSummary(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    }
  }, [monthStart.getTime(), monthEnd.getTime()]);

  useEffect(() => {
    loadSummary();
  }, [loadSummary]);

  useEffect(() => {
    fetchOvertimeTrend()
      .then((data) => setTrend(data ?? []))
      .catch(() => {});
  }, []);

  // Load team / admin overtime breakdown when month changes
  useEffect(() => {
    const from = toDateString(monthStart);
    const to = toDateString(monthEnd);

    fetchMe()
      .then(async (user) => {
        if (isAdmin(user)) {
          const data = await fetchAdminOvertimeSummary(from, to);
          setTeamSummaries(data ?? []);
          setShowTeam(true);
        } else if (isTeamLeaderOrAdmin(user)) {
          const memberships = await fetchMyTeams();
          if (memberships.length > 0) {
            const data = await fetchTeamOvertimeSummary(memberships[0].team_id, from, to);
            setTeamSummaries(data ?? []);
            setShowTeam(true);
          }
        }
      })
      .catch(() => {});
  }, [monthStart.getTime(), monthEnd.getTime()]);

  function prevMonth() {
    setMonth(new Date(month.getFullYear(), month.getMonth() - 1, 1));
  }
  function nextMonth() {
    setMonth(new Date(month.getFullYear(), month.getMonth() + 1, 1));
  }
  function goCurrentMonth() {
    const now = new Date();
    setMonth(new Date(now.getFullYear(), now.getMonth(), 1));
  }

  // Find max bar value for trend chart
  const maxBar = Math.max(
    ...trend.map((t) => Math.max(Math.abs(t.diff_minutes), t.target_minutes, t.actual_minutes)),
    1,
  );

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <TrendingUp className="size-6 text-primary" />
        <h1 className="font-heading text-2xl font-bold tracking-tight">
          Ueberstunden
        </h1>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Month Navigation */}
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={prevMonth}>
          <ChevronLeft className="size-4" />
        </Button>
        <Button variant="outline" size="sm" onClick={goCurrentMonth}>
          Aktuell
        </Button>
        <Button variant="outline" size="sm" onClick={nextMonth}>
          <ChevronRight className="size-4" />
        </Button>
        <span className="ml-2 font-medium">
          {monthStart.toLocaleDateString("de-DE", { month: "long", year: "numeric" })}
        </span>
      </div>

      {/* Current Month Summary */}
      {summary && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Card className="glass-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Soll
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold font-heading">
                {formatHours(summary.target_minutes)} h
              </div>
            </CardContent>
          </Card>

          <Card className="glass-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Ist
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold font-heading">
                {formatHours(summary.actual_minutes)} h
              </div>
            </CardContent>
          </Card>

          <Card className="glass-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Differenz
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div
                className={`text-2xl font-bold font-heading ${
                  summary.diff_minutes >= 0 ? "text-green-400" : "text-red-400"
                }`}
              >
                {formatMinutes(summary.diff_minutes)} h
              </div>
              <Badge
                variant={summary.diff_minutes >= 0 ? "default" : "destructive"}
                className="mt-1"
              >
                {summary.diff_minutes >= 0 ? "Ueberstunden" : "Minusstunden"}
              </Badge>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 6-Month Trend */}
      {trend.length > 0 && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-base">Trend (6 Monate)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {trend.map((t) => {
                const pct =
                  maxBar > 0 ? Math.abs(t.diff_minutes) / maxBar : 0;
                const isPositive = t.diff_minutes >= 0;
                return (
                  <div key={t.period_from} className="flex items-center gap-3">
                    <span className="w-28 text-sm text-muted-foreground shrink-0">
                      {monthLabel(t.period_from)}
                    </span>
                    <div className="flex-1 h-6 bg-muted/30 rounded overflow-hidden relative">
                      <div
                        className={`h-full rounded ${
                          isPositive ? "bg-green-500/40" : "bg-red-500/40"
                        }`}
                        style={{ width: `${Math.max(pct * 100, 2)}%` }}
                      />
                    </div>
                    <span
                      className={`w-16 text-sm text-right font-mono ${
                        isPositive ? "text-green-400" : "text-red-400"
                      }`}
                    >
                      {formatMinutes(t.diff_minutes)}
                    </span>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Team / admin overtime breakdown */}
      {showTeam && teamSummaries.length > 0 && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-base">Team-Ueberstunden</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {[...teamSummaries]
                .sort((a, b) => b.summary.diff_minutes - a.summary.diff_minutes)
                .map((u) => {
                  const positive = u.summary.diff_minutes >= 0;
                  return (
                    <div
                      key={u.user_id}
                      className="grid grid-cols-[1fr_auto_auto_auto] items-center gap-4 rounded-lg border border-border/40 px-4 py-2.5"
                    >
                      <span className="text-sm font-medium truncate">{u.display_name}</span>
                      <span className="text-xs text-muted-foreground text-right">
                        Soll: {formatHours(u.summary.target_minutes)} h
                      </span>
                      <span className="text-xs text-muted-foreground text-right">
                        Ist: {formatHours(u.summary.actual_minutes)} h
                      </span>
                      <span
                        className={`text-sm font-mono font-medium text-right ${
                          positive ? "text-green-400" : "text-red-400"
                        }`}
                      >
                        {formatMinutes(u.summary.diff_minutes)}
                      </span>
                    </div>
                  );
                })}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
