import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import TeamPage from "./page";

vi.mock("@/lib/api", () => ({
  fetchMyTeams: vi.fn().mockResolvedValue([]),
  fetchTeamAvailability: vi.fn().mockResolvedValue([]),
}));

describe("TeamPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page heading", () => {
    render(<TeamPage />);
    expect(
      screen.getByRole("heading", { name: /Teamübersicht/i }),
    ).toBeInTheDocument();
  });

  it("renders week navigation controls", () => {
    render(<TeamPage />);
    const buttons = screen.getAllByRole("button", { name: /Heute/i });
    expect(buttons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders empty state when no team members", async () => {
    render(<TeamPage />);
    const emptyMsg = await screen.findAllByText(/Keine Teammitglieder/i);
    expect(emptyMsg.length).toBeGreaterThanOrEqual(1);
  });

  it("renders legend items", () => {
    render(<TeamPage />);
    expect(screen.getAllByText(/Anwesend/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/Abwesend/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/Homeoffice/i).length).toBeGreaterThanOrEqual(1);
  });
});
