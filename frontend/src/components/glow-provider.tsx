"use client";
import { useEffect } from "react";

export function GlowProvider() {
  useEffect(() => {
    const sync = (e: PointerEvent) => {
      const root = document.documentElement;
      root.style.setProperty("--x", e.clientX.toFixed(2));
      root.style.setProperty("--y", e.clientY.toFixed(2));
      root.style.setProperty("--xp", (e.clientX / window.innerWidth).toFixed(2));
    };
    document.addEventListener("pointermove", sync);
    return () => document.removeEventListener("pointermove", sync);
  }, []);
  return null;
}
