import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import LoginPage from "./page";

describe("LoginPage", () => {
  it("renders the login heading", () => {
    render(<LoginPage />);
    const headings = screen.getAllByRole("heading", { level: 1 });
    expect(headings.length).toBeGreaterThanOrEqual(1);
    const anmeldenElements = screen.getAllByText("Anmelden");
    expect(anmeldenElements.length).toBeGreaterThanOrEqual(1);
  });

  it("renders the Microsoft login link", () => {
    render(<LoginPage />);
    const links = screen.getAllByRole("link", { name: /Anmelden mit Microsoft/ });
    expect(links.length).toBeGreaterThanOrEqual(1);
    expect(links[0]).toHaveAttribute(
      "href",
      expect.stringContaining("/auth/login"),
    );
  });

  it("renders the NEWA badge", () => {
    render(<LoginPage />);
    const elements = screen.getAllByText("NEWA Zeiterfassung");
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it("renders the welcome card title", () => {
    render(<LoginPage />);
    const elements = screen.getAllByText("Willkommen");
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it("renders the DSGVO footer text", () => {
    render(<LoginPage />);
    const elements = screen.getAllByText(/DSGVO-konform/);
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });
});
