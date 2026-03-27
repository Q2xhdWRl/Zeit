"use client";

import { useEffect, useRef, useState } from "react";
import { Clock, LogIn, LogOut, Coffee, Play, X } from "lucide-react";
import type { ActiveStamp } from "@/lib/api";
import { fetchActiveStamp, stampIn, stampOut, toggleBreak, discardStamp } from "@/lib/api";

function formatDuration(startedAt: string, breakMinutes: number, breakStart?: string): string {
  const start = new Date(startedAt);
  const now = new Date();
  let elapsedMs = now.getTime() - start.getTime();
  let totalBreakMs = breakMinutes * 60_000;
  if (breakStart) {
    totalBreakMs += now.getTime() - new Date(breakStart).getTime();
  }
  elapsedMs = Math.max(0, elapsedMs - totalBreakMs);
  const totalMinutes = Math.floor(elapsedMs / 60_000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

export function StampFab() {
  const [stamp, setStamp] = useState<ActiveStamp | null | undefined>(undefined);
  const [elapsed, setElapsed] = useState("");
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const tickRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchActiveStamp().then(setStamp).catch(() => setStamp(null));
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      fetchActiveStamp().then(setStamp).catch(() => {});
    }, 30_000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (stamp) {
      const tick = () =>
        setElapsed(formatDuration(stamp.started_at, stamp.break_minutes, stamp.break_start));
      tick();
      tickRef.current = setInterval(tick, 1000);
    } else {
      setElapsed("");
    }
    return () => {
      if (tickRef.current) clearInterval(tickRef.current);
    };
  }, [stamp]);

  // Close menu when clicking outside
  useEffect(() => {
    function onPointerDown(e: PointerEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    if (open) document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  async function run(fn: () => Promise<unknown>, onSuccess: (result: unknown) => void) {
    setLoading(true);
    setError(null);
    try {
      const result = await fn();
      onSuccess(result);
      setOpen(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler");
    } finally {
      setLoading(false);
    }
  }

  const onBreak = !!stamp?.break_start;

  // Button colour based on state (glass effect with tinted backgrounds)
  const btnClass = stamp
    ? onBreak
      ? "bg-yellow-600/70 hover:bg-yellow-500/80 shadow-yellow-500/30 backdrop-blur-md"
      : "bg-emerald-600/70 hover:bg-emerald-500/80 shadow-emerald-500/30 backdrop-blur-md"
    : "bg-cyan-700/70 hover:bg-cyan-600/80 shadow-cyan-500/30 backdrop-blur-md";

  if (stamp === undefined) return null;

  return (
    <div ref={menuRef} className="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-2">
      {/* Popover menu */}
      {open && (
        <div className="mb-1 rounded-xl border border-white/10 stamp-popup-glass shadow-2xl p-3 w-52 space-y-1">
          {error && <p className="text-xs text-red-400 pb-1">{error}</p>}

          {!stamp && (
            <button
              disabled={loading}
              onClick={() => run(() => stampIn(), (s) => setStamp(s as ActiveStamp))}
              className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-white hover:bg-white/10 transition-colors disabled:opacity-50"
            >
              <LogIn className="size-4 text-cyan-400" />
              Kommen
            </button>
          )}

          {stamp && (
            <>
              <button
                disabled={loading}
                onClick={() => run(() => toggleBreak(), (s) => setStamp(s as ActiveStamp))}
                className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-white hover:bg-white/10 transition-colors disabled:opacity-50"
              >
                {onBreak ? (
                  <>
                    <Play className="size-4 text-emerald-400" />
                    Weiter
                  </>
                ) : (
                  <>
                    <Coffee className="size-4 text-yellow-400" />
                    Pause
                  </>
                )}
              </button>

              <button
                disabled={loading}
                onClick={() => run(() => stampOut(), () => setStamp(null))}
                className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-white hover:bg-white/10 transition-colors disabled:opacity-50"
              >
                <LogOut className="size-4 text-red-400" />
                Gehen
              </button>

              <div className="border-t border-white/10 my-1" />

              <button
                disabled={loading}
                onClick={() => {
                  if (!window.confirm("Stempel verwerfen? Es wird kein Zeiteintrag erstellt.")) return;
                  run(() => discardStamp(), () => setStamp(null));
                }}
                className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-red-400 hover:bg-white/5 transition-colors disabled:opacity-50"
              >
                <X className="size-4" />
                Verwerfen
              </button>
            </>
          )}
        </div>
      )}

      {/* FAB button */}
      <button
        onClick={() => setOpen((o) => !o)}
        aria-label="Stempeluhr"
        aria-expanded={open}
        className={`flex items-center gap-2 rounded-full px-4 py-3 text-white font-medium shadow-lg transition-all ${btnClass} disabled:opacity-50`}
        disabled={loading}
      >
        <Clock className="size-5" />
        {stamp ? (
          <span className="font-mono tabular-nums text-sm">{elapsed}</span>
        ) : (
          <span className="text-sm">Kommen</span>
        )}
      </button>
    </div>
  );
}
