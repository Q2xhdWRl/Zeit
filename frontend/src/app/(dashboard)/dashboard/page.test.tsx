import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import DashboardPage from "./page";

vi.mock("@/lib/api", () => ({
  fetchDashboardStats: vi.fn().mockResolvedValue({
    today_minutes: 480,
    week_minutes: 2400,
    month_overtime: {
      period_from: "2026-03-01",
      period_to: "2026-03-31",
      target_minutes: 9600,
      actual_minutes: 10080,
      diff_minutes: 480,
    },
    team_present_count: 3,
    team_total_count: 5,
  }),
  fetchVacationBalance: vi.fn().mockResolvedValue({
    year: 2026,
    total_days: 30,
    carry_over_days: 2,
    used_days: 5,
    pending_days: 2,
    remaining_days: 25,
  }),
  fetchActiveStamp: vi.fn().mockResolvedValue(null),
  fetchProjects: vi.fn().mockResolvedValue([]),
  stampIn: vi.fn(),
  stampOut: vi.fn(),
  toggleBreak: vi.fn(),
}));

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<DashboardPage />);
    expect(
      screen.getByRole("heading", { name: /Dashboard/i }),
    ).toBeInTheDocument();
  });

  it("renders all four stat cards", () => {
    render(<DashboardPage />);
    expect(screen.getAllByText(/Heute gebucht/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/Resturlaub/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/Team heute/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/Ueberstunden/i).length).toBeGreaterThanOrEqual(1);
  });

  it("displays today's hours after data loads", async () => {
    render(<DashboardPage />);
    const todayValues = await screen.findAllByText(/8:00 h/i);
    expect(todayValues.length).toBeGreaterThanOrEqual(1);
  });
});
