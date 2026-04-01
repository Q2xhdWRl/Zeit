import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import LandingPage from "./page";

describe("LandingPage", () => {
  it("renders the main heading", () => {
    render(<LandingPage />);
    expect(screen.getByRole("heading", { level: 1 })).toBeInTheDocument();
    expect(screen.getByText("Zeit")).toBeInTheDocument();
    expect(screen.getByText(/Team im Blick/)).toBeInTheDocument();
  });

  it("renders all four feature card headings", () => {
    render(<LandingPage />);
    const headings = screen.getAllByRole("heading", { level: 3 });
    const titles = headings.map((h) => h.textContent);
    expect(titles).toContain("Zeiterfassung");
    expect(titles).toContain("Abwesenheiten");
    expect(titles).toContain("Teamübersicht");
    expect(titles).toContain("Überstunden");
  });

  it("renders the Microsoft login button", () => {
    render(<LandingPage />);
    const buttons = screen.getAllByRole("button", {
      name: "Anmelden mit Microsoft",
    });
    expect(buttons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders the NEWA badge text", () => {
    render(<LandingPage />);
    const elements = screen.getAllByText("NEWA Zeiterfassung");
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it("renders the DSGVO footer text", () => {
    render(<LandingPage />);
    const elements = screen.getAllByText(/DSGVO-konform/);
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });
});
