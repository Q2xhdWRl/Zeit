"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Clock, Plus, ChevronLeft, ChevronRight, Trash2, Pencil } from "lucide-react";
import type { TimeEntry, Project, ArbZGViolation } from "@/lib/auth";
import {
  fetchTimeEntries,
  createTimeEntry,
  updateTimeEntry,
  deleteTimeEntry,
  fetchProjects,
} from "@/lib/api";
import ClockWidget from "@/components/clock-widget";

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("de-DE", {
    weekday: "short",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

function formatTime(timeStr: string): string {
  return timeStr.slice(0, 5);
}

function formatMinutes(m: number): string {
  const h = Math.floor(m / 60);
  const min = m % 60;
  return `${h}h ${min.toString().padStart(2, "0")}m`;
}

function getWorkMinutes(entry: TimeEntry): number {
  const [sh, sm] = entry.start_time.split(":").map(Number);
  const [eh, em] = entry.end_time.split(":").map(Number);
  return eh * 60 + em - (sh * 60 + sm) - entry.break_minutes;
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
  return d.toISOString().slice(0, 10);
}

function addDays(d: Date, n: number): Date {
  const result = new Date(d);
  result.setDate(result.getDate() + n);
  return result;
}

const EMPTY_FORM = {
  entry_date: "",
  start_time: "",
  end_time: "",
  break_minutes: 0,
  project_id: "",
  description: "",
};

export default function TimePage() {
  const [entries, setEntries] = useState<TimeEntry[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()));
  const [error, setError] = useState<string | null>(null);
  const [warnings, setWarnings] = useState<ArbZGViolation[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(EMPTY_FORM);

  const weekEnd = addDays(weekStart, 6);

  const loadData = useCallback(async () => {
    try {
      const [e, p] = await Promise.all([
        fetchTimeEntries(toDateString(weekStart), toDateString(weekEnd)),
        fetchProjects(),
      ]);
      setEntries(e ?? []);
      setProjects(p ?? []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    }
  }, [weekStart, weekEnd]);

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

  function openNewForm(date?: string) {
    setForm({
      ...EMPTY_FORM,
      entry_date: date || toDateString(new Date()),
    });
    setEditingId(null);
    setShowForm(true);
    setWarnings([]);
  }

  function openEditForm(entry: TimeEntry) {
    setForm({
      entry_date: entry.entry_date.slice(0, 10),
      start_time: formatTime(entry.start_time),
      end_time: formatTime(entry.end_time),
      break_minutes: entry.break_minutes,
      project_id: entry.project_id || "",
      description: entry.description,
    });
    setEditingId(entry.id);
    setShowForm(true);
    setWarnings([]);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setWarnings([]);

    try {
      const data = {
        entry_date: form.entry_date,
        start_time: form.start_time,
        end_time: form.end_time,
        break_minutes: form.break_minutes,
        project_id: form.project_id || undefined,
        description: form.description || undefined,
      };

      const res = editingId
        ? await updateTimeEntry(editingId, data)
        : await createTimeEntry(data);

      if (res.warnings && res.warnings.length > 0) {
        setWarnings(res.warnings);
      }

      setShowForm(false);
      setEditingId(null);
      setForm(EMPTY_FORM);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Speichern");
    }
  }

  async function handleDelete(entryId: string) {
    try {
      await deleteTimeEntry(entryId);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Loeschen");
    }
  }

  // Group entries by date
  const entriesByDate = new Map<string, TimeEntry[]>();
  for (let i = 0; i < 7; i++) {
    const dateStr = toDateString(addDays(weekStart, i));
    entriesByDate.set(dateStr, []);
  }
  for (const entry of entries) {
    const dateStr = entry.entry_date.slice(0, 10);
    const existing = entriesByDate.get(dateStr);
    if (existing) {
      existing.push(entry);
    }
  }

  // Weekly totals
  const weekTotalMinutes = entries.reduce((sum, e) => sum + getWorkMinutes(e), 0);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-3 mb-4">
            <Clock className="size-6 text-primary" />
            <h1 className="font-heading text-2xl font-bold tracking-tight">
              Zeiterfassung
            </h1>
          </div>
          <div className="w-72">
            <ClockWidget />
          </div>
        </div>
        <Button onClick={() => openNewForm()} className="btn-glow mt-1">
          <Plus className="mr-1 size-4" />
          Neuer Eintrag
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {warnings.length > 0 && (
        <div className="rounded-lg border border-yellow-500/20 bg-yellow-500/5 p-3 text-sm text-yellow-400">
          <p className="font-medium mb-1">ArbZG-Hinweise:</p>
          <ul className="list-disc list-inside">
            {warnings.map((w, i) => (
              <li key={i}>{w.message}</li>
            ))}
          </ul>
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
        <div className="text-sm text-muted-foreground">
          {formatDate(toDateString(weekStart))} &ndash;{" "}
          {formatDate(toDateString(weekEnd))}
        </div>
        <Badge variant="secondary" className="text-sm">
          Woche: {formatMinutes(weekTotalMinutes)}
        </Badge>
      </div>

      {/* Entry Form */}
      {showForm && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="font-heading text-lg">
              {editingId ? "Eintrag bearbeiten" : "Neuer Zeiteintrag"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <div>
                <label htmlFor="entry_date" className="mb-1 block text-sm font-medium">
                  Datum
                </label>
                <input
                  id="entry_date"
                  type="date"
                  required
                  value={form.entry_date}
                  onChange={(e) => setForm({ ...form, entry_date: e.target.value })}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label htmlFor="start_time" className="mb-1 block text-sm font-medium">
                  Beginn
                </label>
                <input
                  id="start_time"
                  type="time"
                  required
                  value={form.start_time}
                  onChange={(e) => setForm({ ...form, start_time: e.target.value })}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label htmlFor="end_time" className="mb-1 block text-sm font-medium">
                  Ende
                </label>
                <input
                  id="end_time"
                  type="time"
                  required
                  value={form.end_time}
                  onChange={(e) => setForm({ ...form, end_time: e.target.value })}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label htmlFor="break_minutes" className="mb-1 block text-sm font-medium">
                  Pause (Minuten)
                </label>
                <input
                  id="break_minutes"
                  type="number"
                  min={0}
                  value={form.break_minutes}
                  onChange={(e) =>
                    setForm({ ...form, break_minutes: parseInt(e.target.value) || 0 })
                  }
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label htmlFor="project_id" className="mb-1 block text-sm font-medium">
                  Projekt
                </label>
                <select
                  id="project_id"
                  value={form.project_id}
                  onChange={(e) => setForm({ ...form, project_id: e.target.value })}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                >
                  <option value="">Kein Projekt</option>
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                      {p.customer_name ? ` (${p.customer_name})` : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="description" className="mb-1 block text-sm font-medium">
                  Beschreibung
                </label>
                <input
                  id="description"
                  type="text"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  placeholder="Was wurde gemacht?"
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="flex items-end gap-2 sm:col-span-2 lg:col-span-3">
                <Button type="submit" className="btn-glow">
                  {editingId ? "Speichern" : "Eintrag erstellen"}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setShowForm(false);
                    setEditingId(null);
                    setForm(EMPTY_FORM);
                  }}
                >
                  Abbrechen
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Weekly View */}
      <div className="space-y-4">
        {Array.from(entriesByDate.entries()).map(([dateStr, dayEntries]) => {
          const dayTotal = dayEntries.reduce((sum, e) => sum + getWorkMinutes(e), 0);
          const isToday = dateStr === toDateString(new Date());

          return (
            <Card
              key={dateStr}
              className={`glass-card ${isToday ? "ring-1 ring-primary/30" : ""}`}
            >
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-base font-heading">
                    {formatDate(dateStr)}
                    {isToday && (
                      <Badge variant="default" className="text-xs">
                        Heute
                      </Badge>
                    )}
                  </CardTitle>
                  <div className="flex items-center gap-3">
                    {dayEntries.length > 0 && (
                      <span className="text-sm text-muted-foreground">
                        {formatMinutes(dayTotal)}
                      </span>
                    )}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => openNewForm(dateStr)}
                    >
                      <Plus className="size-3" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              {dayEntries.length > 0 && (
                <CardContent className="pt-0">
                  <div className="space-y-2">
                    {dayEntries.map((entry) => {
                      const workMin = getWorkMinutes(entry);
                      const project = projects.find(
                        (p) => p.id === entry.project_id,
                      );
                      return (
                        <div
                          key={entry.id}
                          className="flex items-center justify-between rounded-lg border border-border/50 p-3"
                        >
                          <div className="flex items-center gap-4">
                            <span className="font-mono text-sm font-medium">
                              {formatTime(entry.start_time)} &ndash;{" "}
                              {formatTime(entry.end_time)}
                            </span>
                            {entry.break_minutes > 0 && (
                              <Badge variant="outline" className="text-xs">
                                {entry.break_minutes}m Pause
                              </Badge>
                            )}
                            <span className="text-sm text-primary">
                              {formatMinutes(workMin)}
                            </span>
                            {project && (
                              <Badge variant="secondary" className="text-xs">
                                {project.name}
                              </Badge>
                            )}
                            {entry.description && (
                              <span className="text-sm text-muted-foreground">
                                {entry.description}
                              </span>
                            )}
                          </div>
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => openEditForm(entry)}
                              aria-label="Bearbeiten"
                            >
                              <Pencil className="size-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDelete(entry.id)}
                              aria-label="Loeschen"
                            >
                              <Trash2 className="size-3.5 text-destructive" />
                            </Button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </CardContent>
              )}
              {dayEntries.length === 0 && (
                <CardContent className="pt-0">
                  <p className="text-xs text-muted-foreground text-center py-2">
                    Keine Eintraege
                  </p>
                </CardContent>
              )}
            </Card>
          );
        })}
      </div>
    </div>
  );
}
