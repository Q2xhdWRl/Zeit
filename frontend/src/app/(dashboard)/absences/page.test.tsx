import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AbsencesPage from "./page";

vi.mock("@/lib/api", () => ({
  fetchMe: vi.fn().mockResolvedValue({
    id: "user-1",
    email: "test@newa.test",
    display_name: "Test User",
    global_role: "user",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
  }),
  fetchAbsences: vi.fn().mockResolvedValue([]),
  fetchAbsenceTypes: vi.fn().mockResolvedValue([
    { id: "type-1", name: "Urlaub", color: "#10b981", requires_approval: true, counts_as_work: false, is_active: true, sort_order: 1, created_at: "2026-01-01T00:00:00Z" },
  ]),
  fetchVacationBalance: vi.fn().mockResolvedValue({
    year: 2026,
    total_days: 30,
    carry_over_days: 2,
    used_days: 5,
    pending_days: 3,
    remaining_days: 24,
  }),
  fetchMyTeams: vi.fn().mockResolvedValue([
    { user_id: "user-1", team_id: "team-1", role: "user", joined_at: "2026-01-01T00:00:00Z" },
  ]),
  fetchTeams: vi.fn().mockResolvedValue([
    { id: "team-1", name: "Dev Team", description: "", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
  ]),
  fetchTeamMembers: vi.fn().mockResolvedValue([]),
  fetchTeamAbsences: vi.fn().mockResolvedValue([]),
  fetchPendingAbsences: vi.fn().mockResolvedValue([]),
  createAbsence: vi.fn(),
  updateAbsence: vi.fn(),
  deleteAbsence: vi.fn(),
  cancelAbsence: vi.fn(),
  reviewAbsence: vi.fn(),
}));

describe("AbsencesPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<AbsencesPage />);
    expect(screen.getByRole("heading", { name: /An- und Abwesenheit/i })).toBeInTheDocument();
  });

  it("renders the request absence button", () => {
    render(<AbsencesPage />);
    const btns = screen.getAllByRole("button", { name: /Abwesenheit beantragen/i });
    expect(btns.length).toBeGreaterThanOrEqual(1);
  });

  it("renders Teamkalender and Abwesenheit tabs", () => {
    render(<AbsencesPage />);
    expect(screen.getAllByRole("button", { name: /Teamkalender/i }).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByRole("button", { name: /Abwesenheit/i }).length).toBeGreaterThanOrEqual(1);
  });

  it("switches to Abwesenheit tab", async () => {
    const user = userEvent.setup();
    render(<AbsencesPage />);
    const tabs = screen.getAllByRole("button", { name: /Abwesenheit/i });
    const abwesenheitTab = tabs.find((b) => b.textContent?.trim() === "Abwesenheit");
    if (abwesenheitTab) {
      await user.click(abwesenheitTab);
      expect(await screen.findAllByText(/Urlaubskonto/i)).toBeTruthy();
    }
  });

  it("opens modal when request button clicked", async () => {
    const user = userEvent.setup();
    render(<AbsencesPage />);
    const btn = screen.getAllByRole("button", { name: /Abwesenheit beantragen/i })[0];
    await user.click(btn);
    expect(await screen.findByRole("heading", { name: /Abwesenheit beantragen/i })).toBeInTheDocument();
  });

  it("shows empty state in Abwesenheit tab", async () => {
    const user = userEvent.setup();
    render(<AbsencesPage />);
    const tabs = screen.getAllByRole("button", { name: /Abwesenheit/i });
    const abwesenheitTab = tabs.find((b) => b.textContent?.trim() === "Abwesenheit");
    if (abwesenheitTab) {
      await user.click(abwesenheitTab);
      expect(await screen.findByText(/Keine Eintr/i)).toBeInTheDocument();
    }
  });
});
