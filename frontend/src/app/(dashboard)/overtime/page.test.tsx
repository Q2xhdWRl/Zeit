import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import OvertimePage from "./page";

vi.mock("@/lib/api", () => ({
  fetchOvertimeSummary: vi.fn().mockResolvedValue({
    period_from: "2026-03-01",
    period_to: "2026-03-31",
    target_minutes: 9600,
    actual_minutes: 10080,
    diff_minutes: 480,
  }),
  fetchOvertimeTrend: vi.fn().mockResolvedValue([]),
}));

describe("OvertimePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<OvertimePage />);
    expect(
      screen.getByRole("heading", { name: /Überstunden/i }),
    ).toBeInTheDocument();
  });

  it("renders month navigation controls", () => {
    render(<OvertimePage />);
    const buttons = screen.getAllByRole("button", { name: /Aktuell/i });
    expect(buttons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders Soll/Ist/Differenz cards", async () => {
    render(<OvertimePage />);
    const sollCards = await screen.findAllByText(/Soll/i);
    expect(sollCards.length).toBeGreaterThanOrEqual(1);

    const istCards = await screen.findAllByText(/Ist/i);
    expect(istCards.length).toBeGreaterThanOrEqual(1);

    const diffCards = await screen.findAllByText(/Differenz/i);
    expect(diffCards.length).toBeGreaterThanOrEqual(1);
  });
});
