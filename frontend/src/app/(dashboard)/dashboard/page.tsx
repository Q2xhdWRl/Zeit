"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Clock, CalendarDays, Users, TrendingUp, CalendarOff } from "lucide-react";
import type { DashboardStats } from "@/lib/auth";
import { fetchDashboardStats, fetchVacationBalance } from "@/lib/api";
import ClockWidget from "@/components/clock-widget";
import { GlowCard } from "@/components/ui/spotlight-card";

function formatMinutes(m: number): string {
  const h = Math.floor(Math.abs(m) / 60);
  const min = Math.abs(m) % 60;
  return `${h}:${min.toString().padStart(2, "0")}`;
}

function formatDiff(m: number): string {
  const sign = m >= 0 ? "+" : "-";
  return `${sign}${formatMinutes(m)}`;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [vacationDays, setVacationDays] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDashboardStats()
      .then(setStats)
      .catch((err) =>
        setError(err instanceof Error ? err.message : "Fehler beim Laden"),
      );

    fetchVacationBalance()
      .then((b) => {
        if (b) setVacationDays(b.remaining_days);
      })
      .catch(() => {});
  }, []);

  const todayH = stats ? formatMinutes(stats.today_minutes) : "0:00";
  const weekH = stats ? formatMinutes(stats.week_minutes) : "0:00";
  const overtime = stats?.month_overtime;
  const overtimeStr = overtime ? formatDiff(overtime.diff_minutes) : "+0:00";
  const overtimePositive = overtime ? overtime.diff_minutes >= 0 : true;

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-2xl font-bold tracking-tight">
        Dashboard
      </h1>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-1">
          <ClockWidget />
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Today */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Heute gebucht</CardTitle>
            <Clock className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-heading">{todayH} h</div>
            <p className="text-xs text-muted-foreground">
              Diese Woche: {weekH} h
            </p>
          </CardContent>
        </Card>

        {/* Vacation */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Resturlaub</CardTitle>
            <CalendarDays className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-heading">
              {vacationDays !== null ? `${vacationDays} Tage` : "-- Tage"}
            </div>
            <p className="text-xs text-muted-foreground">
              {vacationDays !== null
                ? `Verbleibend ${new Date().getFullYear()}`
                : "Noch nicht konfiguriert"}
            </p>
          </CardContent>
        </Card>

        {/* Team */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Team heute</CardTitle>
            <Users className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-heading">
              {stats
                ? `${stats.team_present_count}/${stats.team_total_count}`
                : "--"}
            </div>
            <p className="text-xs text-muted-foreground">Anwesend</p>
          </CardContent>
        </Card>

        {/* Overtime */}
        <Card className="glass-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Ueberstunden</CardTitle>
            <TrendingUp className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div
              className={`text-2xl font-bold font-heading ${
                overtimePositive ? "text-green-400" : "text-red-400"
              }`}
            >
              {overtimeStr} h
            </div>
            <Badge
              variant={overtimePositive ? "default" : "destructive"}
              className="mt-1"
            >
              {overtimePositive ? "Ueberstunden" : "Minusstunden"}
            </Badge>
          </CardContent>
        </Card>
      </div>

      <div className="flex flex-wrap gap-6">
        <GlowCard glowColor="blue" size="sm" className="bg-black/30">
          <div className="flex flex-col items-center justify-center gap-2 text-center">
            <Clock className="size-8 text-cyan-400" />
            <span className="font-heading text-sm font-semibold text-white">Zeiterfassung</span>
            <span className="text-xs text-gray-400">Stunden buchen</span>
          </div>
        </GlowCard>

        <GlowCard glowColor="purple" size="sm" className="bg-black/30">
          <div className="flex flex-col items-center justify-center gap-2 text-center">
            <CalendarOff className="size-8 text-violet-400" />
            <span className="font-heading text-sm font-semibold text-white">Abwesenheit</span>
            <span className="text-xs text-gray-400">Urlaub & Krankmeldung</span>
          </div>
        </GlowCard>

        <GlowCard glowColor="green" size="sm" className="bg-black/30">
          <div className="flex flex-col items-center justify-center gap-2 text-center">
            <Users className="size-8 text-emerald-400" />
            <span className="font-heading text-sm font-semibold text-white">Team</span>
            <span className="text-xs text-gray-400">Teammitglieder</span>
          </div>
        </GlowCard>
      </div>
    </div>
  );
}
