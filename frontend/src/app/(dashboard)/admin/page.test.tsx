import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import AdminPage from "./page";

vi.mock("@/lib/api", () => ({
  fetchUsers: vi.fn().mockResolvedValue([
    {
      id: "u1",
      email: "admin@test.com",
      display_name: "Admin User",
      global_role: "admin",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "u2",
      email: "user@test.com",
      display_name: "Normal User",
      global_role: "user",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
    },
  ]),
  updateUserRole: vi.fn().mockResolvedValue(undefined),
  updateUserActive: vi.fn().mockResolvedValue(undefined),
  fetchTeams: vi.fn().mockResolvedValue([
    {
      id: "t1",
      name: "Entwicklung",
      description: "Dev team",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ]),
  createTeam: vi.fn().mockResolvedValue({}),
  deleteTeam: vi.fn().mockResolvedValue(undefined),
  fetchAllProjects: vi.fn().mockResolvedValue([
    {
      id: "p1",
      name: "Zeiterfassung",
      customer_name: "NEWA",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
    },
  ]),
  createProject: vi.fn().mockResolvedValue({}),
  updateProject: vi.fn().mockResolvedValue({}),
  fetchAllAbsenceTypes: vi.fn().mockResolvedValue([
    {
      id: "at1",
      name: "Urlaub",
      color: "#22c55e",
      requires_approval: true,
      counts_as_work: false,
      is_active: true,
      sort_order: 1,
      created_at: "2026-01-01T00:00:00Z",
    },
  ]),
  updateAbsenceType: vi.fn().mockResolvedValue({}),
  upsertEntitlement: vi.fn().mockResolvedValue({}),
  upsertWorkSchedule: vi.fn().mockResolvedValue({}),
  fetchTeamMembers: vi.fn().mockResolvedValue([]),
  addTeamMember: vi.fn().mockResolvedValue(undefined),
  removeTeamMember: vi.fn().mockResolvedValue(undefined),
}));

describe("AdminPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<AdminPage />);
    expect(
      screen.getByRole("heading", { name: /Verwaltung/i }),
    ).toBeInTheDocument();
  });

  it("renders all tab buttons", () => {
    render(<AdminPage />);
    const tabLabels = ["Benutzer", "Teams", "Projekte", "Abwesenheitstypen", "Urlaubskontingente", "Arbeitszeitmodelle"];
    for (const label of tabLabels) {
      const buttons = screen.getAllByRole("button", { name: new RegExp(label, "i") });
      expect(buttons.length).toBeGreaterThanOrEqual(1);
    }
  });

  it("shows users tab by default with user data", async () => {
    render(<AdminPage />);
    const rows = await screen.findAllByText(/admin@test.com/i);
    expect(rows.length).toBeGreaterThanOrEqual(1);
  });

  it("switches to projects tab on click", async () => {
    render(<AdminPage />);
    const tabs = screen.getAllByRole("button", { name: /Projekte/i });
    fireEvent.click(tabs[0]);
    const projectNames = await screen.findAllByText(/Zeiterfassung/i);
    expect(projectNames.length).toBeGreaterThanOrEqual(1);
  });

  it("switches to absence types tab on click", async () => {
    render(<AdminPage />);
    const tabs = screen.getAllByRole("button", { name: /Abwesenheitstypen/i });
    fireEvent.click(tabs[0]);
    const types = await screen.findAllByText(/Urlaub/i);
    expect(types.length).toBeGreaterThanOrEqual(1);
  });

  it("switches to entitlements tab and shows form", () => {
    render(<AdminPage />);
    const tabs = screen.getAllByRole("button", { name: /Urlaubskontingente/i });
    fireEvent.click(tabs[0]);
    expect(screen.getAllByLabelText(/Urlaubstage gesamt/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByLabelText(/Resturlaub Vorjahr/i).length).toBeGreaterThanOrEqual(1);
  });

  it("switches to schedules tab and shows weekday inputs", () => {
    render(<AdminPage />);
    const tabs = screen.getAllByRole("button", { name: /Arbeitszeitmodelle/i });
    fireEvent.click(tabs[0]);
    expect(screen.getAllByLabelText(/Wochenstunden/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByLabelText(/Gueltig ab/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Mo").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Fr").length).toBeGreaterThanOrEqual(1);
  });

  it("switches to teams tab and shows team list", async () => {
    render(<AdminPage />);
    const tabs = screen.getAllByRole("button", { name: /^Teams$/i });
    fireEvent.click(tabs[0]);
    const teamNames = await screen.findAllByText(/Entwicklung/i);
    expect(teamNames.length).toBeGreaterThanOrEqual(1);
  });
});
