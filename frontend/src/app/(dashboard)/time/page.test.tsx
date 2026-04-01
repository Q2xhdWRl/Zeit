import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import TimePage from "./page";

vi.mock("@/lib/api", () => ({
  fetchTimeEntries: vi.fn().mockResolvedValue([]),
  fetchProjects: vi.fn().mockResolvedValue([]),
  createTimeEntry: vi.fn(),
  updateTimeEntry: vi.fn(),
  deleteTimeEntry: vi.fn(),
  fetchActiveStamp: vi.fn().mockResolvedValue(null),
  stampIn: vi.fn(),
  stampOut: vi.fn(),
  toggleBreak: vi.fn(),
}));

describe("TimePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<TimePage />);
    expect(
      screen.getByRole("heading", { name: /Zeiterfassung/i }),
    ).toBeInTheDocument();
  });

  it("renders the new entry button", () => {
    render(<TimePage />);
    const buttons = screen.getAllByRole("button", { name: /Neuer Eintrag/i });
    expect(buttons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders week navigation controls", () => {
    render(<TimePage />);
    const buttons = screen.getAllByRole("button", { name: /Heute/i });
    expect(buttons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders day cards with empty messages", async () => {
    render(<TimePage />);
    const emptyMessages = await screen.findAllByText("Keine Einträge");
    // At least 7 (one per day, may render more in strict mode)
    expect(emptyMessages.length).toBeGreaterThanOrEqual(7);
  });

  it("shows the weekly total badge", () => {
    render(<TimePage />);
    const badges = screen.getAllByText(/Woche:/);
    expect(badges.length).toBeGreaterThanOrEqual(1);
  });
});
