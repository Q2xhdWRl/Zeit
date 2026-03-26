"use client";

import { useEffect, useRef, useState } from "react";
import type { ActiveStamp } from "@/lib/api";
import { fetchActiveStamp, stampIn, stampOut, toggleBreak, discardStamp } from "@/lib/api";

function formatDuration(
  startedAt: string,
  breakMinutes: number,
  breakStart?: string,
  now?: Date,
): string {
  const start = new Date(startedAt);
  const current = now ?? new Date();
  let elapsedMs = current.getTime() - start.getTime();

  let totalBreakMs = breakMinutes * 60_000;
  if (breakStart) {
    totalBreakMs += current.getTime() - new Date(breakStart).getTime();
  }
  elapsedMs = Math.max(0, elapsedMs - totalBreakMs);

  const totalMinutes = Math.floor(elapsedMs / 60_000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString("de-DE", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function ClockWidget() {
  const [stamp, setStamp] = useState<ActiveStamp | null | undefined>(undefined);
  const [elapsed, setElapsed] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const tickRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Load active stamp on mount.
  useEffect(() => {
    fetchActiveStamp()
      .then(setStamp)
      .catch(() => setStamp(null));
  }, []);

  // Poll every 30s to sync across tabs.
  useEffect(() => {
    const interval = setInterval(() => {
      fetchActiveStamp().then(setStamp).catch(() => {});
    }, 30_000);
    return () => clearInterval(interval);
  }, []);

  // Tick every second while stamped in.
  useEffect(() => {
    if (stamp) {
      const tick = () => {
        setElapsed(
          formatDuration(stamp.started_at, stamp.break_minutes, stamp.break_start),
        );
      };
      tick();
      tickRef.current = setInterval(tick, 1000);
    } else {
      setElapsed("");
    }
    return () => {
      if (tickRef.current) clearInterval(tickRef.current);
    };
  }, [stamp]);

  async function handleKommen() {
    setLoading(true);
    setError(null);
    try {
      const s = await stampIn();
      setStamp(s);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler beim Kommen");
    } finally {
      setLoading(false);
    }
  }

  async function handleGehen() {
    setLoading(true);
    setError(null);
    try {
      await stampOut();
      setStamp(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler beim Gehen");
    } finally {
      setLoading(false);
    }
  }

  async function handleDiscard() {
    if (!window.confirm("Stempel verwerfen? Es wird kein Zeiteintrag erstellt.")) return;
    setLoading(true);
    setError(null);
    try {
      await discardStamp();
      setStamp(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler beim Verwerfen");
    } finally {
      setLoading(false);
    }
  }

  async function handleToggleBreak() {
    setLoading(true);
    setError(null);
    try {
      const updated = await toggleBreak();
      setStamp(updated);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler bei der Pause");
    } finally {
      setLoading(false);
    }
  }

  if (stamp === undefined) {
    return (
      <div className="rounded-xl border border-white/10 bg-white/5 p-4 text-sm text-gray-400">
        Laden…
      </div>
    );
  }

  const onBreak = !!stamp?.break_start;

  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-4 space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-widest text-gray-400">
          Stempeluhr
        </span>
        {stamp && (
          <span
            className={`text-xs px-2 py-0.5 rounded-full font-medium ${
              onBreak
                ? "bg-yellow-500/20 text-yellow-300"
                : "bg-emerald-500/20 text-emerald-300"
            }`}
          >
            {onBreak ? "Pause" : "Aktiv"}
          </span>
        )}
      </div>

      {stamp ? (
        <>
          <div className="text-center">
            <div className="text-3xl font-mono font-bold text-white tabular-nums">
              {elapsed}
            </div>
            <div className="text-xs text-gray-400 mt-1">
              Kommen: {formatTime(stamp.started_at)}
              {stamp.break_minutes > 0 && ` · ${stamp.break_minutes} min Pause`}
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleToggleBreak}
              disabled={loading}
              className={`flex-1 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                onBreak
                  ? "bg-emerald-600 hover:bg-emerald-500 text-white"
                  : "bg-yellow-600 hover:bg-yellow-500 text-white"
              } disabled:opacity-50`}
            >
              {onBreak ? "Weiter" : "Pause"}
            </button>
            <button
              onClick={handleGehen}
              disabled={loading}
              className="flex-1 rounded-lg bg-red-700 hover:bg-red-600 text-white px-3 py-2 text-sm font-medium transition-colors disabled:opacity-50"
            >
              Gehen
            </button>
          </div>
        </>
      ) : (
        <button
          onClick={handleKommen}
          disabled={loading}
          className="w-full rounded-lg bg-cyan-700 hover:bg-cyan-600 text-white px-3 py-2 text-sm font-medium transition-colors disabled:opacity-50"
        >
          Kommen
        </button>
      )}

      {error && <p className="text-xs text-red-400">{error}</p>}

      {stamp && (
        <button
          onClick={handleDiscard}
          disabled={loading}
          className="w-full text-xs text-gray-500 hover:text-red-400 transition-colors mt-1 disabled:opacity-50"
        >
          Stempel verwerfen
        </button>
      )}
    </div>
  );
}
