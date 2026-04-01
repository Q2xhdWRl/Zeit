"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  CalendarDays,
  Plus,
  ChevronLeft,
  ChevronRight,
  X,
} from "lucide-react";
import type { Absence, AbsenceType, VacationBalance, User, TeamMember } from "@/lib/auth";
import {
  fetchAbsences,
  fetchAbsenceTypes,
  fetchVacationBalance,
  createAbsence,
  updateAbsence,
  deleteAbsence,
  cancelAbsence,
  fetchMe,
  fetchMyTeams,
  fetchTeamMembers,
  fetchTeamAbsences,
  fetchPendingAbsences,
  reviewAbsence,
  fetchTeams,
} from "@/lib/api";
import type { Team } from "@/lib/auth";

// ─── constants & helpers ──────────────────────────────────────────────────────

const DAY_W = 32;
const LEFT_W = 220;

const DAY_ABBR = ["So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"];
const MONTH_NAMES = [
  "Januar","Februar","März","April","Mai","Juni",
  "Juli","August","September","Oktober","November","Dezember",
];

const statusLabels: Record<string, string> = {
  pending: "Ausstehend",
  approved: "Genehmigt",
  rejected: "Abgelehnt",
  cancelled: "Storniert",
};
const statusColors: Record<string, string> = {
  pending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  approved: "bg-green-500/10 text-green-400 border-green-500/20",
  rejected: "bg-red-500/10 text-red-400 border-red-500/20",
  cancelled: "bg-gray-500/10 text-gray-400 border-gray-500/20",
};

function toDateStr(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function addDays(d: Date, n: number): Date {
  const r = new Date(d);
  r.setDate(r.getDate() + n);
  return r;
}

function daysBetween(a: Date, b: Date): number {
  const ua = Date.UTC(a.getFullYear(), a.getMonth(), a.getDate());
  const ub = Date.UTC(b.getFullYear(), b.getMonth(), b.getDate());
  return Math.round((ub - ua) / 86400000);
}

function isoWeek(d: Date): number {
  const tmp = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
  tmp.setUTCDate(tmp.getUTCDate() + 4 - (tmp.getUTCDay() || 7));
  const yearStart = new Date(Date.UTC(tmp.getUTCFullYear(), 0, 1));
  return Math.ceil(((tmp.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

function workingDays(startStr: string, endStr: string): number {
  let count = 0;
  for (let d = new Date(startStr); d <= new Date(endStr); d.setDate(d.getDate() + 1)) {
    const wd = d.getDay();
    if (wd !== 0 && wd !== 6) count++;
  }
  return count;
}

function formatDateShort(s: string): string {
  return new Date(s).toLocaleDateString("de-DE", { day: "2-digit", month: "2-digit" });
}

function formatDateLong(s: string): string {
  return new Date(s).toLocaleDateString("de-DE", { day: "2-digit", month: "2-digit", year: "numeric" });
}

function initials(name: string): string {
  return name.split(" ").map((w) => w[0] ?? "").join("").slice(0, 2).toUpperCase();
}

// ─── AbsenceFormModal ─────────────────────────────────────────────────────────

interface FormValues {
  absence_type_id: string;
  start_date: string;
  end_date: string;
  note: string;
}

interface AbsenceFormModalProps {
  absenceTypes: AbsenceType[];
  editingAbsence?: Absence | null;
  initialStartDate?: string;
  onClose: () => void;
  onSaved: () => void;
}

function AbsenceFormModal({
  absenceTypes,
  editingAbsence,
  initialStartDate,
  onClose,
  onSaved,
}: AbsenceFormModalProps) {
  const today = toDateStr(new Date());
  const [form, setForm] = useState<FormValues>({
    absence_type_id: editingAbsence?.absence_type_id ?? "",
    start_date: editingAbsence?.start_date.slice(0, 10) ?? initialStartDate ?? today,
    end_date: editingAbsence?.end_date.slice(0, 10) ?? initialStartDate ?? today,
    note: editingAbsence?.note ?? "",
  });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const days =
    form.start_date && form.end_date
      ? workingDays(form.start_date, form.end_date)
      : 0;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const data = {
        absence_type_id: form.absence_type_id,
        start_date: form.start_date,
        end_date: form.end_date,
        note: form.note || undefined,
      };
      if (editingAbsence) {
        await updateAbsence(editingAbsence.id, data);
      } else {
        await createAbsence(data);
      }
      onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Speichern");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-white/10 bg-[#0f1117] p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-heading text-lg font-bold">
            {editingAbsence ? "Abwesenheit bearbeiten" : "Abwesenheit beantragen"}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <X className="size-5" />
          </button>
        </div>
        {error && <p className="mb-3 text-sm text-red-400">{error}</p>}
        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label htmlFor="abs_type" className="mb-1 block text-sm font-medium">
              Art
            </label>
            <select
              id="abs_type"
              required
              value={form.absence_type_id}
              onChange={(e) => setForm({ ...form, absence_type_id: e.target.value })}
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
            >
              <option value="">Bitte wählen</option>
              {absenceTypes.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label htmlFor="abs_start" className="mb-1 block text-sm font-medium">
                Von
              </label>
              <input
                id="abs_start"
                type="date"
                required
                value={form.start_date}
                onChange={(e) => setForm({ ...form, start_date: e.target.value })}
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label htmlFor="abs_end" className="mb-1 block text-sm font-medium">
                Bis
              </label>
              <input
                id="abs_end"
                type="date"
                required
                value={form.end_date}
                onChange={(e) => setForm({ ...form, end_date: e.target.value })}
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
          </div>
          {days > 0 && (
            <p className="text-xs text-muted-foreground">
              {days} Arbeitstag{days !== 1 ? "e" : ""}
            </p>
          )}
          <div>
            <label htmlFor="abs_note" className="mb-1 block text-sm font-medium">
              Bemerkung (optional)
            </label>
            <input
              id="abs_note"
              type="text"
              value={form.note}
              onChange={(e) => setForm({ ...form, note: e.target.value })}
              placeholder="Optional"
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
            />
          </div>
          <div className="flex gap-2 pt-1">
            <Button type="submit" disabled={loading} className="btn-glow flex-1">
              {editingAbsence ? "Speichern" : "Beantragen"}
            </Button>
            <Button type="button" variant="outline" onClick={onClose}>
              Abbrechen
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── ReviewModal ──────────────────────────────────────────────────────────────

interface ReviewModalProps {
  absence: Absence;
  absenceTypes: AbsenceType[];
  memberName: string;
  onClose: () => void;
  onReviewed: () => void;
}

function ReviewModal({ absence, absenceTypes, memberName, onClose, onReviewed }: ReviewModalProps) {
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const typeName =
    absenceTypes.find((t) => t.id === absence.absence_type_id)?.name ?? "Abwesenheit";

  async function handleReview(approve: boolean) {
    setLoading(true);
    setError(null);
    try {
      await reviewAbsence(absence.id, approve, note);
      onReviewed();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-white/10 bg-[#0f1117] p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-heading text-base font-bold">Abwesenheit prüfen</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <X className="size-4" />
          </button>
        </div>
        <p className="mb-1 text-sm font-medium">{memberName}</p>
        <p className="mb-1 text-sm text-muted-foreground">{typeName}</p>
        <p className="mb-3 text-sm text-muted-foreground">
          {formatDateShort(absence.start_date.slice(0, 10))}
          {" – "}
          {formatDateShort(absence.end_date.slice(0, 10))}
          {" · "}
          {workingDays(absence.start_date.slice(0, 10), absence.end_date.slice(0, 10))} Arbeitstage
        </p>
        {absence.note && (
          <p className="mb-3 text-xs text-gray-400">Notiz: {absence.note}</p>
        )}
        {error && <p className="mb-2 text-xs text-red-400">{error}</p>}
        <div>
          <label className="mb-1 block text-sm font-medium">Kommentar (optional)</label>
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Begründung..."
            className="mb-3 w-full rounded border border-border bg-background px-3 py-2 text-sm"
          />
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => handleReview(true)}
            disabled={loading}
            className="flex-1 rounded-lg bg-green-700 py-2 text-sm font-medium text-white transition-colors hover:bg-green-600 disabled:opacity-50"
          >
            Genehmigen
          </button>
          <button
            onClick={() => handleReview(false)}
            disabled={loading}
            className="flex-1 rounded-lg bg-red-700 py-2 text-sm font-medium text-white transition-colors hover:bg-red-600 disabled:opacity-50"
          >
            Ablehnen
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── GanttTimeline ──────────────────────────────────────────────────────────────

interface GanttProps {
  members: TeamMember[];
  absences: Absence[];
  pendingAbsences: Absence[];
  absenceTypes: AbsenceType[];
  canReview: boolean;
  viewStart: Date;
  numDays: number;
  onReviewed: () => void;
  onRequestAbsence: () => void;
}

function GanttTimeline({
  members,
  absences,
  pendingAbsences,
  absenceTypes,
  canReview,
  viewStart,
  numDays,
  onReviewed,
  onRequestAbsence,
}: GanttProps) {
  const [reviewing, setReviewing] = useState<Absence | null>(null);
  const reviewingMember = reviewing
    ? members.find((m) => m.user_id === reviewing.user_id)
    : null;

  const today = toDateStr(new Date());
  const todayIndex = daysBetween(viewStart, new Date());

  // Build days array.
  const days: Date[] = [];
  for (let i = 0; i < numDays; i++) {
    days.push(addDays(viewStart, i));
  }

  // Build month header spans.
  type Span = { label: string; count: number };
  const monthSpans: Span[] = [];
  days.forEach((d) => {
    const label = `${MONTH_NAMES[d.getMonth()]} ${d.getFullYear()}`;
    if (!monthSpans.length || monthSpans[monthSpans.length - 1]?.label !== label) {
      monthSpans.push({ label, count: 1 });
    } else {
      monthSpans[monthSpans.length - 1]!.count++;
    }
  });

  // Build week header spans.
  const weekSpans: Span[] = [];
  days.forEach((d) => {
    const label = `KW${isoWeek(d)}`;
    if (!weekSpans.length || weekSpans[weekSpans.length - 1]?.label !== label) {
      weekSpans.push({ label, count: 1 });
    } else {
      weekSpans[weekSpans.length - 1]!.count++;
    }
  });

  function getAbsenceBars(userId: string) {
    return absences.filter((a) => a.user_id === userId);
  }

  function getAbsenceStyle(absence: Absence): React.CSSProperties | null {
    const start = new Date(absence.start_date);
    const end = new Date(absence.end_date);
    const si = daysBetween(viewStart, start);
    const ei = daysBetween(viewStart, end);
    if (ei < 0 || si >= numDays) return null;
    const clampedSi = Math.max(0, si);
    const clampedEi = Math.min(numDays - 1, ei);
    const w = (clampedEi - clampedSi + 1) * DAY_W;
    const l = clampedSi * DAY_W;
    return { left: l, width: w };
  }

  function getTypeColor(typeId: string): string {
    return absenceTypes.find((t) => t.id === typeId)?.color ?? "#6b7280";
  }

  function getTypeName(typeId: string): string {
    return absenceTypes.find((t) => t.id === typeId)?.name ?? "";
  }

  const pendingCount = pendingAbsences.length;

  return (
    <>
      {reviewing && reviewingMember && (
        <ReviewModal
          absence={reviewing}
          absenceTypes={absenceTypes}
          memberName={reviewingMember.display_name ?? reviewingMember.email ?? "Benutzer"}
          onClose={() => setReviewing(null)}
          onReviewed={onReviewed}
        />
      )}

      {/* Pending notice for team leaders */}
      {canReview && pendingCount > 0 && (
        <div className="mb-4 flex items-center gap-2 rounded-lg border border-yellow-500/20 bg-yellow-500/5 px-4 py-2 text-sm text-yellow-300">
          <span className="font-medium">
            {pendingCount} {pendingCount === 1 ? "ausstehender" : "ausstehende"}
          </span>
          <span>
            {pendingCount === 1
              ? "Abwesenheitsantrag — klicke auf den Balken zum Prüfen"
              : "Abwesenheitsanträge — klicke auf einen Balken zum Prüfen"}
          </span>
        </div>
      )}

      {/* Gantt container */}
      <div className="overflow-x-auto rounded-xl border border-white/10">
        <div style={{ minWidth: LEFT_W + numDays * DAY_W }}>

          {/* Month header */}
          <div className="flex border-b border-white/10 bg-white/5">
            <div
              className="shrink-0 border-r border-white/10 px-3 py-1.5 text-xs text-gray-400"
              style={{ width: LEFT_W }}
            >
              {members.length} Person{members.length !== 1 ? "en" : ""}
            </div>
            {monthSpans.map((s, i) => (
              <div
                key={i}
                className="border-r border-white/10 px-2 py-1.5 text-center text-xs font-semibold text-gray-200"
                style={{ width: s.count * DAY_W }}
              >
                {s.label}
              </div>
            ))}
          </div>

          {/* Week header */}
          <div className="flex border-b border-white/10 bg-white/5">
            <div className="shrink-0 border-r border-white/10" style={{ width: LEFT_W }} />
            {weekSpans.map((s, i) => (
              <div
                key={i}
                className="border-r border-white/10 px-1 py-1 text-center text-xs text-gray-400"
                style={{ width: s.count * DAY_W }}
              >
                {s.label}
              </div>
            ))}
          </div>

          {/* Day header */}
          <div className="flex border-b border-white/10 bg-white/5">
            <div className="shrink-0 border-r border-white/10" style={{ width: LEFT_W }} />
            {days.map((d, i) => {
              const isToday = toDateStr(d) === today;
              const isWeekend = d.getDay() === 0 || d.getDay() === 6;
              return (
                <div
                  key={i}
                  className={`border-r border-white/5 text-center text-[10px] py-1 ${
                    isToday
                      ? "bg-cyan-500/20 font-bold text-cyan-300"
                      : isWeekend
                        ? "text-gray-600"
                        : "text-gray-400"
                  }`}
                  style={{ width: DAY_W }}
                >
                  {DAY_ABBR[d.getDay()]}
                  <br />
                  <span className="text-[9px]">{d.getDate()}</span>
                </div>
              );
            })}
          </div>

          {/* Person rows */}
          {members.map((member) => {
            const bars = getAbsenceBars(member.user_id);
            return (
              <div
                key={member.user_id}
                className="flex border-b border-white/5 hover:bg-white/3"
              >
                {/* Person info (sticky) */}
                <div
                  className="sticky left-0 z-10 flex shrink-0 items-center gap-2 border-r border-white/10 bg-[#0a0d14] px-3 py-2"
                  style={{ width: LEFT_W }}
                >
                  <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/20 text-xs font-bold text-primary">
                    {initials(member.display_name ?? member.email ?? "?")}
                  </div>
                  <div className="min-w-0">
                    <div className="truncate text-xs font-medium text-white">
                      {member.display_name ?? member.email}
                    </div>
                    <div className="truncate text-[10px] text-gray-500">
                      {member.role === "team_leader"
                        ? "Teamleiter"
                        : member.role === "admin"
                          ? "Admin"
                          : "Benutzer"}
                    </div>
                  </div>
                </div>

                {/* Timeline */}
                <div
                  className="relative"
                  style={{ width: numDays * DAY_W, height: 48 }}
                >
                  {/* Weekend shading */}
                  {days.map((d, i) => {
                    const isWeekend = d.getDay() === 0 || d.getDay() === 6;
                    if (!isWeekend) return null;
                    return (
                      <div
                        key={i}
                        className="absolute inset-y-0 bg-white/[0.02]"
                        style={{ left: i * DAY_W, width: DAY_W }}
                      />
                    );
                  })}

                  {/* Today line */}
                  {todayIndex >= 0 && todayIndex < numDays && (
                    <div
                      className="absolute inset-y-0 w-px bg-cyan-500/60"
                      style={{ left: todayIndex * DAY_W + DAY_W / 2 }}
                    />
                  )}

                  {/* Absence bars */}
                  {bars.map((absence) => {
                    const style = getAbsenceStyle(absence);
                    if (!style) return null;
                    const color = getTypeColor(absence.absence_type_id);
                    const name = getTypeName(absence.absence_type_id);
                    const durationDays = daysBetween(
                      new Date(absence.start_date),
                      new Date(absence.end_date),
                    ) + 1;
                    const isPending = absence.status === "pending";
                    const isSingle = durationDays <= 1;

                    return (
                      <button
                        key={absence.id}
                        onClick={() => canReview && isPending && setReviewing(absence)}
                        className={`absolute top-1/2 -translate-y-1/2 rounded transition-opacity ${
                          canReview && isPending
                            ? "cursor-pointer hover:opacity-90"
                            : "cursor-default"
                        }`}
                        style={{
                          left: (style.left as number) + 2,
                          width: Math.max((style.width as number) - 4, DAY_W - 4),
                          height: isSingle ? 28 : 22,
                          backgroundColor: color + (isPending ? "80" : "cc"),
                          borderLeft: `3px solid ${color}`,
                          borderRadius: 4,
                        }}
                        title={`${name} (${formatDateShort(absence.start_date.slice(0, 10))} – ${formatDateShort(absence.end_date.slice(0, 10))})${isPending ? " · Ausstehend" : ""}`}
                      >
                        {!isSingle && (style.width as number) > 60 && (
                          <span className="truncate px-1 text-[10px] font-medium text-white">
                            {name}
                            {durationDays > 1 ? ` · ${durationDays} Tage` : ""}
                          </span>
                        )}
                        {isPending && (
                          <span
                            className="absolute -top-1 -right-1 size-2 rounded-full bg-yellow-400"
                            title="Ausstehend"
                          />
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}

          {members.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">
              Keine Teammitglieder gefunden
            </div>
          )}
        </div>
      </div>
    </>
  );
}

// ─── MonthCalendar ──────────────────────────────────────────────────────────────

interface MonthCalendarProps {
  absences: Absence[];
  absenceTypes: AbsenceType[];
  balance: VacationBalance | null;
  monthStart: Date;
  onEdit: (a: Absence) => void;
  onDelete: (id: string) => void;
  onCancel: (id: string) => void;
}

function MonthCalendar({
  absences,
  absenceTypes,
  balance,
  monthStart,
  onEdit,
  onDelete,
  onCancel,
}: MonthCalendarProps) {
  const year = monthStart.getFullYear();
  const month = monthStart.getMonth();

  // Build calendar grid.
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  // Start on Monday (ISO)
  const startOffset = (firstDay.getDay() + 6) % 7;
  const gridStart = addDays(firstDay, -startOffset);
  const totalCells = Math.ceil((startOffset + lastDay.getDate()) / 7) * 7;
  const cells: Date[] = [];
  for (let i = 0; i < totalCells; i++) {
    cells.push(addDays(gridStart, i));
  }
  const weeks: Date[][] = [];
  for (let i = 0; i < cells.length; i += 7) {
    weeks.push(cells.slice(i, i + 7));
  }

  function absencesOnDay(d: Date): Absence[] {
    const ds = toDateStr(d);
    return absences.filter((a) => {
      const s = a.start_date.slice(0, 10);
      const e = a.end_date.slice(0, 10);
      return ds >= s && ds <= e;
    });
  }

  function getTypeColor(typeId: string): string {
    return absenceTypes.find((t) => t.id === typeId)?.color ?? "#6b7280";
  }

  function getTypeName(typeId: string): string {
    return absenceTypes.find((t) => t.id === typeId)?.name ?? "Unbekannt";
  }

  const today = toDateStr(new Date());
  const sortedAbsences = [...absences].sort(
    (a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime(),
  );

  return (
    <div className="space-y-6">
      {/* Vacation balance */}
      {balance && (
        <Card className="glass-card">
          <CardHeader className="pb-3">
            <CardTitle className="font-heading text-base">Urlaubskonto {balance.year}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
              {[
                { label: "Anspruch", value: balance.total_days },
                { label: "Übertrag", value: balance.carry_over_days },
                { label: "Genommen", value: balance.used_days },
                { label: "Beantragt", value: balance.pending_days },
                { label: "Verbleibend", value: balance.remaining_days, highlight: balance.remaining_days < 5 ? "text-yellow-400" : "text-green-400" },
              ].map(({ label, value, highlight }) => (
                <div key={label}>
                  <p className="text-xs text-muted-foreground">{label}</p>
                  <p className={`text-lg font-semibold ${highlight ?? ""}`}>{value}d</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Calendar grid */}
      <Card className="glass-card">
        <CardContent className="pt-4">
          <div className="grid grid-cols-7 text-center">
            {["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"].map((d) => (
              <div key={d} className="py-1 text-xs font-semibold text-muted-foreground">
                {d}
              </div>
            ))}
            {weeks.map((week, wi) =>
              week.map((day, di) => {
                const ds = toDateStr(day);
                const isCurrentMonth = day.getMonth() === month;
                const isToday = ds === today;
                const dayAbsences = absencesOnDay(day);
                const firstAbsence = dayAbsences[0];
                const color = firstAbsence ? getTypeColor(firstAbsence.absence_type_id) : null;

                return (
                  <div
                    key={`${wi}-${di}`}
                    className={`relative min-h-[40px] rounded p-1 text-center text-sm ${
                      isToday
                        ? "ring-1 ring-cyan-500"
                        : ""
                    } ${!isCurrentMonth ? "opacity-30" : ""}`}
                    style={
                      color
                        ? { backgroundColor: color + "33", borderRadius: 4 }
                        : undefined
                    }
                    title={
                      dayAbsences.map((a) => getTypeName(a.absence_type_id)).join(", ") || undefined
                    }
                  >
                    <span
                      className={`text-xs ${isToday ? "font-bold text-cyan-300" : "text-gray-300"}`}
                    >
                      {day.getDate()}
                    </span>
                    {firstAbsence && (
                      <div
                        className="mx-auto mt-0.5 h-1 w-1 rounded-full"
                        style={{ backgroundColor: color ?? "#6b7280" }}
                      />
                    )}
                  </div>
                );
              }),
            )}
          </div>
        </CardContent>
      </Card>

      {/* Absence list */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-muted-foreground">
          Abwesenheiten {MONTH_NAMES[month]} {year}
        </h3>
        {sortedAbsences.length === 0 && (
          <p className="text-sm text-muted-foreground">Keine Einträge in diesem Monat</p>
        )}
        {sortedAbsences.map((absence) => {
          const typeColor = getTypeColor(absence.absence_type_id);
          const typeName = getTypeName(absence.absence_type_id);
          const days = workingDays(absence.start_date.slice(0, 10), absence.end_date.slice(0, 10));
          const isPending = absence.status === "pending";
          const isApproved = absence.status === "approved";

          return (
            <Card key={absence.id} className="glass-card">
              <CardContent className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <div
                    className="h-8 w-1 rounded-full"
                    style={{ backgroundColor: typeColor }}
                  />
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{typeName}</span>
                      <Badge
                        variant="outline"
                        className={`text-xs ${statusColors[absence.status]}`}
                      >
                        {statusLabels[absence.status]}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {formatDateLong(absence.start_date.slice(0, 10))}
                      {absence.start_date.slice(0, 10) !== absence.end_date.slice(0, 10) && (
                        <> – {formatDateLong(absence.end_date.slice(0, 10))}</>
                      )}{" "}
                      · {days} Arbeitstag{days !== 1 ? "e" : ""}
                      {absence.note ? ` · ${absence.note}` : ""}
                    </p>
                  </div>
                </div>
                <div className="flex gap-1">
                  {isPending && (
                    <>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onEdit(absence)}
                        aria-label="Bearbeiten"
                      >
                        <CalendarDays className="size-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(absence.id)}
                        aria-label="Löschen"
                      >
                        <X className="size-3.5 text-destructive" />
                      </Button>
                    </>
                  )}
                  {isApproved && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onCancel(absence.id)}
                      aria-label="Stornieren"
                    >
                      <X className="size-3.5 text-destructive" />
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}

// ─── Main Page ────────────────────────────────────────────────────────────────

type Tab = "teamkalender" | "abwesenheit";

export default function AbsencesPage() {
  const [tab, setTab] = useState<Tab>("teamkalender");
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  // Shared data
  const [absenceTypes, setAbsenceTypes] = useState<AbsenceType[]>([]);
  const [error, setError] = useState<string | null>(null);

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [editingAbsence, setEditingAbsence] = useState<Absence | null>(null);

  // Teamkalender state
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeamId, setSelectedTeamId] = useState<string>("");
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [teamAbsences, setTeamAbsences] = useState<Absence[]>([]);
  const [pendingAbsences, setPendingAbsences] = useState<Absence[]>([]);
  const [ganttMonthStart, setGanttMonthStart] = useState(() => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 1);
  });

  // Own absences state
  const [ownAbsences, setOwnAbsences] = useState<Absence[]>([]);
  const [balance, setBalance] = useState<VacationBalance | null>(null);
  const [ownMonthStart, setOwnMonthStart] = useState(() => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 1);
  });

  // Gantt view covers 3 months starting from ganttMonthStart.
  const ganttViewStart = ganttMonthStart;
  const ganttNumDays = (() => {
    const end = new Date(ganttMonthStart.getFullYear(), ganttMonthStart.getMonth() + 3, 0);
    return daysBetween(ganttViewStart, end) + 1;
  })();
  const ganttViewEnd = addDays(ganttViewStart, ganttNumDays - 1);

  const canReview =
    currentUser?.global_role === "admin" || currentUser?.global_role === "team_leader";

  // Load current user + absence types once.
  useEffect(() => {
    Promise.all([fetchMe(), fetchAbsenceTypes()])
      .then(([user, types]) => {
        setCurrentUser(user);
        setAbsenceTypes(types ?? []);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Fehler beim Laden"));
  }, []);

  // Load teams when user is ready.
  useEffect(() => {
    if (!currentUser) return;
    fetchMyTeams()
      .then((memberships) => {
        const teamIds = [...new Set(memberships.map((m) => m.team_id))];
        return fetchTeams().then((allTeams) => {
          const myTeams = allTeams.filter((t) => teamIds.includes(t.id));
          setTeams(myTeams);
          if (myTeams.length > 0 && !selectedTeamId) {
            setSelectedTeamId(myTeams[0]!.id);
          }
        });
      })
      .catch(() => {});
  }, [currentUser]);

  // Load team data when team or gantt range changes.
  const loadTeamData = useCallback(async () => {
    if (!selectedTeamId) return;
    try {
      const from = toDateStr(ganttViewStart);
      const to = toDateStr(ganttViewEnd);
      const [members, absences] = await Promise.all([
        fetchTeamMembers(selectedTeamId),
        fetchTeamAbsences(selectedTeamId, from, to),
      ]);
      setTeamMembers(members ?? []);
      setTeamAbsences(absences ?? []);

      if (canReview) {
        const pending = await fetchPendingAbsences(selectedTeamId);
        setPendingAbsences(pending ?? []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    }
  }, [selectedTeamId, ganttViewStart.getTime(), canReview]);

  useEffect(() => {
    loadTeamData();
  }, [loadTeamData]);

  // Load own absences when month changes.
  const loadOwnAbsences = useCallback(async () => {
    try {
      const monthEnd = new Date(ownMonthStart.getFullYear(), ownMonthStart.getMonth() + 1, 0);
      const [absences, bal] = await Promise.all([
        fetchAbsences(toDateStr(ownMonthStart), toDateStr(monthEnd)),
        fetchVacationBalance(ownMonthStart.getFullYear()),
      ]);
      setOwnAbsences(absences ?? []);
      setBalance(bal);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    }
  }, [ownMonthStart]);

  useEffect(() => {
    loadOwnAbsences();
  }, [loadOwnAbsences]);

  function openNewAbsence() {
    setEditingAbsence(null);
    setShowModal(true);
  }

  function openEditAbsence(a: Absence) {
    setEditingAbsence(a);
    setShowModal(true);
  }

  async function handleDelete(id: string) {
    try {
      await deleteAbsence(id);
      await loadOwnAbsences();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Löschen");
    }
  }

  async function handleCancel(id: string) {
    try {
      await cancelAbsence(id);
      await loadOwnAbsences();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Stornieren");
    }
  }

  const ganttMonthLabel = ganttMonthStart.toLocaleDateString("de-DE", {
    month: "long",
    year: "numeric",
  });
  const ownMonthLabel = ownMonthStart.toLocaleDateString("de-DE", {
    month: "long",
    year: "numeric",
  });

  return (
    <div className="flex flex-col gap-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <CalendarDays className="size-6 text-primary" />
          <h1 className="font-heading text-2xl font-bold tracking-tight">
            An- und Abwesenheit
          </h1>
        </div>
        <Button onClick={openNewAbsence} className="btn-glow">
          <Plus className="mr-1 size-4" />
          Abwesenheit beantragen
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b border-white/10">
        {(["teamkalender", "abwesenheit"] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium transition-colors ${
              tab === t
                ? "border-b-2 border-primary text-primary"
                : "text-muted-foreground hover:text-white"
            }`}
          >
            {t === "teamkalender" ? "Teamkalender" : "Abwesenheit"}
            {t === "teamkalender" && canReview && pendingAbsences.length > 0 && (
              <span className="ml-2 rounded-full bg-yellow-500 px-1.5 py-0.5 text-[10px] font-bold text-black">
                {pendingAbsences.length}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Teamkalender tab */}
      {tab === "teamkalender" && (
        <>
          {/* Controls */}
          <div className="flex flex-wrap items-center gap-3">
            {teams.length > 1 && (
              <select
                value={selectedTeamId}
                onChange={(e) => setSelectedTeamId(e.target.value)}
                className="rounded border border-border bg-background px-3 py-1.5 text-sm"
              >
                {teams.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
            )}
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  setGanttMonthStart(
                    new Date(ganttMonthStart.getFullYear(), ganttMonthStart.getMonth() - 1, 1),
                  )
                }
              >
                <ChevronLeft className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const d = new Date();
                  setGanttMonthStart(new Date(d.getFullYear(), d.getMonth(), 1));
                }}
              >
                Heute
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  setGanttMonthStart(
                    new Date(ganttMonthStart.getFullYear(), ganttMonthStart.getMonth() + 1, 1),
                  )
                }
              >
                <ChevronRight className="size-4" />
              </Button>
            </div>
            <span className="text-sm capitalize text-muted-foreground">{ganttMonthLabel}</span>
          </div>

          <GanttTimeline
            members={teamMembers}
            absences={teamAbsences}
            pendingAbsences={pendingAbsences}
            absenceTypes={absenceTypes}
            canReview={canReview}
            viewStart={ganttViewStart}
            numDays={ganttNumDays}
            onReviewed={loadTeamData}
            onRequestAbsence={openNewAbsence}
          />
        </>
      )}

      {/* Own absences tab */}
      {tab === "abwesenheit" && (
        <>
          {/* Month navigation */}
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                setOwnMonthStart(
                  new Date(ownMonthStart.getFullYear(), ownMonthStart.getMonth() - 1, 1),
                )
              }
            >
              <ChevronLeft className="size-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                const d = new Date();
                setOwnMonthStart(new Date(d.getFullYear(), d.getMonth(), 1));
              }}
            >
              Heute
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                setOwnMonthStart(
                  new Date(ownMonthStart.getFullYear(), ownMonthStart.getMonth() + 1, 1),
                )
              }
            >
              <ChevronRight className="size-4" />
            </Button>
            <span className="text-sm capitalize text-muted-foreground">{ownMonthLabel}</span>
          </div>

          <MonthCalendar
            absences={ownAbsences}
            absenceTypes={absenceTypes}
            balance={balance}
            monthStart={ownMonthStart}
            onEdit={openEditAbsence}
            onDelete={handleDelete}
            onCancel={handleCancel}
          />
        </>
      )}

      {/* Absence form modal */}
      {showModal && (
        <AbsenceFormModal
          absenceTypes={absenceTypes}
          editingAbsence={editingAbsence}
          onClose={() => {
            setShowModal(false);
            setEditingAbsence(null);
          }}
          onSaved={() => {
            loadOwnAbsences();
            if (tab === "teamkalender") loadTeamData();
          }}
        />
      )}
    </div>
  );
}
