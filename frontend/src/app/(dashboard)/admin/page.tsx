"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Shield,
  Users,
  Plus,
  FolderKanban,
  CalendarOff,
  CalendarDays,
  Clock,
  Trash2,
  Save,
  UserPlus,
  UserMinus,
} from "lucide-react";
import type { User, UserRole, Team, TeamMember, Project, AbsenceType, WorkSchedule } from "@/lib/auth";
import {
  fetchUsers,
  updateUserRole,
  updateUserActive,
  fetchTeams,
  createTeam,
  deleteTeam,
  fetchAllProjects,
  createProject,
  updateProject,
  fetchAllAbsenceTypes,
  updateAbsenceType,
  upsertEntitlement,
  upsertWorkSchedule,
  fetchTeamMembers,
  addTeamMember,
  removeTeamMember,
} from "@/lib/api";

type AdminTab = "users" | "teams" | "projects" | "absence_types" | "entitlements" | "schedules";

const TABS: { key: AdminTab; label: string; icon: React.ReactNode }[] = [
  { key: "users", label: "Benutzer", icon: <Users className="size-4" /> },
  { key: "teams", label: "Teams", icon: <Users className="size-4" /> },
  { key: "projects", label: "Projekte", icon: <FolderKanban className="size-4" /> },
  { key: "absence_types", label: "Abwesenheitstypen", icon: <CalendarOff className="size-4" /> },
  { key: "entitlements", label: "Urlaubskontingente", icon: <CalendarDays className="size-4" /> },
  { key: "schedules", label: "Arbeitszeitmodelle", icon: <Clock className="size-4" /> },
];

function RoleBadge({ role }: { role: UserRole }) {
  const variant = role === "admin" ? "default" : role === "team_leader" ? "secondary" : "outline";
  const label = role === "admin" ? "Admin" : role === "team_leader" ? "Teamleiter" : "Benutzer";
  return <Badge variant={variant}>{label}</Badge>;
}

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState<AdminTab>("users");
  const [users, setUsers] = useState<User[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [absenceTypes, setAbsenceTypes] = useState<AbsenceType[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [newTeamName, setNewTeamName] = useState("");
  const [newProjectName, setNewProjectName] = useState("");
  const [newProjectCustomer, setNewProjectCustomer] = useState("");

  // Team members state
  const [selectedTeam, setSelectedTeam] = useState<string | null>(null);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [addMemberUserId, setAddMemberUserId] = useState("");
  const [addMemberRole, setAddMemberRole] = useState<UserRole>("user");

  // Entitlement form state
  const [entUserID, setEntUserID] = useState("");
  const [entYear, setEntYear] = useState(new Date().getFullYear());
  const [entTotalDays, setEntTotalDays] = useState(30);
  const [entCarryOver, setEntCarryOver] = useState(0);

  // Work schedule form state
  const [wsUserID, setWsUserID] = useState("");
  const [wsValidFrom, setWsValidFrom] = useState("");
  const [wsWeekly, setWsWeekly] = useState(40);
  const [wsDays, setWsDays] = useState([8, 8, 8, 8, 8, 0, 0]);

  const loadData = useCallback(async () => {
    try {
      const [u, t, p, at] = await Promise.all([
        fetchUsers(),
        fetchTeams(),
        fetchAllProjects().catch(() => []),
        fetchAllAbsenceTypes().catch(() => []),
      ]);
      setUsers(u ?? []);
      setTeams(t ?? []);
      setProjects(p ?? []);
      setAbsenceTypes(at ?? []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  async function loadTeamMembers(teamId: string) {
    try {
      const members = await fetchTeamMembers(teamId);
      setTeamMembers(members ?? []);
      setSelectedTeam(teamId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Laden der Mitglieder");
    }
  }

  async function handleRoleChange(userId: string, newRole: UserRole) {
    try {
      await updateUserRole(userId, newRole);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Aendern der Rolle");
    }
  }

  async function handleToggleActive(userId: string, currentlyActive: boolean) {
    try {
      await updateUserActive(userId, !currentlyActive);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Aendern des Status");
    }
  }

  async function handleCreateTeam() {
    if (!newTeamName.trim()) return;
    try {
      await createTeam(newTeamName.trim(), "");
      setNewTeamName("");
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Erstellen");
    }
  }

  async function handleDeleteTeam(teamId: string) {
    if (!window.confirm("Team wirklich loeschen?")) return;
    try {
      await deleteTeam(teamId);
      if (selectedTeam === teamId) {
        setSelectedTeam(null);
        setTeamMembers([]);
      }
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Loeschen");
    }
  }

  async function handleAddMember() {
    if (!selectedTeam || !addMemberUserId) return;
    try {
      await addTeamMember(selectedTeam, addMemberUserId, addMemberRole);
      setAddMemberUserId("");
      await loadTeamMembers(selectedTeam);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Hinzufuegen");
    }
  }

  async function handleRemoveMember(userId: string) {
    if (!selectedTeam) return;
    if (!window.confirm("Mitglied wirklich entfernen?")) return;
    try {
      await removeTeamMember(selectedTeam, userId);
      await loadTeamMembers(selectedTeam);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Entfernen");
    }
  }

  async function handleCreateProject() {
    if (!newProjectName.trim()) return;
    try {
      await createProject(newProjectName.trim(), newProjectCustomer.trim());
      setNewProjectName("");
      setNewProjectCustomer("");
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Erstellen");
    }
  }

  async function handleToggleProject(p: Project) {
    try {
      await updateProject(p.id, p.name, p.customer_name, !p.is_active);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Aendern");
    }
  }

  async function handleUpdateAbsenceType(at: AbsenceType, updates: Partial<AbsenceType>) {
    try {
      await updateAbsenceType(at.id, {
        name: updates.name ?? at.name,
        color: updates.color ?? at.color,
        requires_approval: updates.requires_approval ?? at.requires_approval,
        counts_as_work: updates.counts_as_work ?? at.counts_as_work,
        is_active: updates.is_active ?? at.is_active,
        sort_order: updates.sort_order ?? at.sort_order,
      });
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Aendern");
    }
  }

  async function handleSaveEntitlement() {
    if (!entUserID) return;
    try {
      await upsertEntitlement({
        user_id: entUserID,
        year: entYear,
        total_days: entTotalDays,
        carry_over_days: entCarryOver,
      });
      setError(null);
      alert("Urlaubskontingent gespeichert");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Speichern");
    }
  }

  async function handleSaveSchedule() {
    if (!wsUserID || !wsValidFrom) return;
    try {
      await upsertWorkSchedule({
        user_id: wsUserID,
        valid_from: wsValidFrom,
        weekly_hours: wsWeekly,
        monday_hours: wsDays[0],
        tuesday_hours: wsDays[1],
        wednesday_hours: wsDays[2],
        thursday_hours: wsDays[3],
        friday_hours: wsDays[4],
        saturday_hours: wsDays[5],
        sunday_hours: wsDays[6],
      });
      setError(null);
      alert("Arbeitszeitmodell gespeichert");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Fehler beim Speichern");
    }
  }

  const dayLabels = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];

  if (loading) {
    return (
      <div className="flex flex-col gap-6 animate-pulse">
        <div className="h-8 w-36 rounded bg-muted/40" />
        <div className="flex gap-1 rounded-lg bg-muted/10 p-1">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-10 w-24 rounded-md bg-muted/20" />
          ))}
        </div>
        <div className="h-64 rounded-xl bg-muted/20" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <Shield className="size-6 text-primary" />
        <h1 className="font-heading text-2xl font-bold tracking-tight">
          Verwaltung
        </h1>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Tab Navigation */}
      <div className="flex flex-wrap gap-1 rounded-lg border border-border/50 bg-muted/30 p-1">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex items-center gap-1.5 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
              activeTab === tab.key
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            }`}
            aria-label={tab.label}
          >
            {tab.icon}
            <span className="hidden sm:inline">{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Users Tab */}
      {activeTab === "users" && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 font-heading">
              <Users className="size-5" />
              Benutzer ({users.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    <th className="pb-3 font-medium">Name</th>
                    <th className="pb-3 font-medium">E-Mail</th>
                    <th className="pb-3 font-medium">Rolle</th>
                    <th className="pb-3 font-medium">Status</th>
                    <th className="pb-3 font-medium">Aktionen</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u) => (
                    <tr key={u.id} className="border-b border-border/50">
                      <td className="py-3">{u.display_name}</td>
                      <td className="py-3 text-muted-foreground">{u.email}</td>
                      <td className="py-3">
                        <RoleBadge role={u.global_role} />
                      </td>
                      <td className="py-3">
                        <Badge variant={u.is_active ? "outline" : "destructive"}>
                          {u.is_active ? "Aktiv" : "Inaktiv"}
                        </Badge>
                      </td>
                      <td className="py-3">
                        <div className="flex gap-2">
                          <select
                            value={u.global_role}
                            onChange={(e) =>
                              handleRoleChange(u.id, e.target.value as UserRole)
                            }
                            className="rounded border border-border bg-background px-2 py-1 text-xs"
                            aria-label={`Rolle fuer ${u.display_name}`}
                          >
                            <option value="user">Benutzer</option>
                            <option value="team_leader">Teamleiter</option>
                            <option value="admin">Admin</option>
                          </select>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleActive(u.id, u.is_active)}
                          >
                            {u.is_active ? "Deaktivieren" : "Aktivieren"}
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Teams Tab */}
      {activeTab === "teams" && (
        <div className="flex flex-col gap-4">
          <Card className="glass-card">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 font-heading">
                <Users className="size-5" />
                Teams ({teams.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="mb-4 flex gap-2">
                <input
                  type="text"
                  value={newTeamName}
                  onChange={(e) => setNewTeamName(e.target.value)}
                  placeholder="Neuer Teamname"
                  className="flex-1 rounded border border-border bg-background px-3 py-2 text-sm"
                  onKeyDown={(e) => e.key === "Enter" && handleCreateTeam()}
                  aria-label="Neuer Teamname"
                />
                <Button onClick={handleCreateTeam} size="sm">
                  <Plus className="mr-1 size-4" />
                  Team erstellen
                </Button>
              </div>

              <div className="space-y-2">
                {teams.map((team) => (
                  <div
                    key={team.id}
                    className={`flex items-center justify-between rounded-lg border p-3 transition-colors cursor-pointer ${
                      selectedTeam === team.id
                        ? "border-primary/50 bg-primary/5"
                        : "border-border/50 hover:border-border"
                    }`}
                    onClick={() => loadTeamMembers(team.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        loadTeamMembers(team.id);
                      }
                    }}
                    aria-label={`Team ${team.name} auswaehlen`}
                  >
                    <div>
                      <p className="font-medium">{team.name}</p>
                      {team.description && (
                        <p className="text-xs text-muted-foreground">{team.description}</p>
                      )}
                    </div>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteTeam(team.id);
                      }}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </div>
                ))}
                {teams.length === 0 && (
                  <p className="text-sm text-muted-foreground py-4 text-center">
                    Noch keine Teams erstellt.
                  </p>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Team Members Section */}
          {selectedTeam && (
            <Card className="glass-card">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 font-heading text-base">
                  <Users className="size-4" />
                  Mitglieder: {teams.find((t) => t.id === selectedTeam)?.name}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="mb-4 flex flex-wrap gap-2">
                  <select
                    value={addMemberUserId}
                    onChange={(e) => setAddMemberUserId(e.target.value)}
                    className="flex-1 rounded border border-border bg-background px-3 py-2 text-sm"
                    aria-label="Benutzer auswaehlen"
                  >
                    <option value="">Benutzer waehlen...</option>
                    {users
                      .filter((u) => !teamMembers.some((m) => m.user_id === u.id))
                      .map((u) => (
                        <option key={u.id} value={u.id}>
                          {u.display_name} ({u.email})
                        </option>
                      ))}
                  </select>
                  <select
                    value={addMemberRole}
                    onChange={(e) => setAddMemberRole(e.target.value as UserRole)}
                    className="rounded border border-border bg-background px-2 py-2 text-sm"
                    aria-label="Rolle im Team"
                  >
                    <option value="user">Benutzer</option>
                    <option value="team_leader">Teamleiter</option>
                    <option value="admin">Admin</option>
                  </select>
                  <Button onClick={handleAddMember} size="sm" disabled={!addMemberUserId}>
                    <UserPlus className="mr-1 size-4" />
                    Hinzufuegen
                  </Button>
                </div>

                <div className="space-y-2">
                  {teamMembers.map((m) => (
                    <div
                      key={m.user_id}
                      className="flex items-center justify-between rounded-lg border border-border/50 p-3"
                    >
                      <div>
                        <p className="font-medium">{m.display_name || m.email || m.user_id}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <RoleBadge role={m.role} />
                          {m.email && (
                            <span className="text-xs text-muted-foreground">{m.email}</span>
                          )}
                        </div>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRemoveMember(m.user_id)}
                      >
                        <UserMinus className="size-4" />
                      </Button>
                    </div>
                  ))}
                  {teamMembers.length === 0 && (
                    <p className="text-sm text-muted-foreground py-4 text-center">
                      Keine Mitglieder in diesem Team.
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Projects Tab */}
      {activeTab === "projects" && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 font-heading">
              <FolderKanban className="size-5" />
              Projekte ({projects.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="mb-4 flex flex-wrap gap-2">
              <input
                type="text"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="Projektname"
                className="flex-1 min-w-[150px] rounded border border-border bg-background px-3 py-2 text-sm"
                aria-label="Projektname"
              />
              <input
                type="text"
                value={newProjectCustomer}
                onChange={(e) => setNewProjectCustomer(e.target.value)}
                placeholder="Kundenname"
                className="flex-1 min-w-[150px] rounded border border-border bg-background px-3 py-2 text-sm"
                onKeyDown={(e) => e.key === "Enter" && handleCreateProject()}
                aria-label="Kundenname"
              />
              <Button onClick={handleCreateProject} size="sm">
                <Plus className="mr-1 size-4" />
                Projekt erstellen
              </Button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    <th className="pb-3 font-medium">Name</th>
                    <th className="pb-3 font-medium">Kunde</th>
                    <th className="pb-3 font-medium">Status</th>
                    <th className="pb-3 font-medium">Aktionen</th>
                  </tr>
                </thead>
                <tbody>
                  {projects.map((p) => (
                    <tr key={p.id} className="border-b border-border/50">
                      <td className="py-3 font-medium">{p.name}</td>
                      <td className="py-3 text-muted-foreground">{p.customer_name}</td>
                      <td className="py-3">
                        <Badge variant={p.is_active ? "outline" : "destructive"}>
                          {p.is_active ? "Aktiv" : "Inaktiv"}
                        </Badge>
                      </td>
                      <td className="py-3">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleToggleProject(p)}
                        >
                          {p.is_active ? "Deaktivieren" : "Aktivieren"}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {projects.length === 0 && (
                <p className="text-sm text-muted-foreground py-4 text-center">
                  Noch keine Projekte erstellt.
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Absence Types Tab */}
      {activeTab === "absence_types" && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 font-heading">
              <CalendarOff className="size-5" />
              Abwesenheitstypen ({absenceTypes.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    <th className="pb-3 font-medium">Farbe</th>
                    <th className="pb-3 font-medium">Name</th>
                    <th className="pb-3 font-medium">Genehmigung</th>
                    <th className="pb-3 font-medium">Zaehlt als Arbeit</th>
                    <th className="pb-3 font-medium">Status</th>
                    <th className="pb-3 font-medium">Reihenfolge</th>
                  </tr>
                </thead>
                <tbody>
                  {absenceTypes.map((at) => (
                    <tr key={at.id} className="border-b border-border/50">
                      <td className="py-3">
                        <input
                          type="color"
                          value={at.color}
                          onChange={(e) =>
                            handleUpdateAbsenceType(at, { color: e.target.value })
                          }
                          className="h-8 w-8 cursor-pointer rounded border border-border"
                          aria-label={`Farbe fuer ${at.name}`}
                        />
                      </td>
                      <td className="py-3 font-medium">{at.name}</td>
                      <td className="py-3">
                        <button
                          onClick={() =>
                            handleUpdateAbsenceType(at, {
                              requires_approval: !at.requires_approval,
                            })
                          }
                          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                            at.requires_approval
                              ? "bg-amber-500/20 text-amber-400"
                              : "bg-muted text-muted-foreground"
                          }`}
                          aria-label={`Genehmigungspflicht fuer ${at.name}`}
                        >
                          {at.requires_approval ? "Ja" : "Nein"}
                        </button>
                      </td>
                      <td className="py-3">
                        <button
                          onClick={() =>
                            handleUpdateAbsenceType(at, {
                              counts_as_work: !at.counts_as_work,
                            })
                          }
                          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                            at.counts_as_work
                              ? "bg-green-500/20 text-green-400"
                              : "bg-muted text-muted-foreground"
                          }`}
                          aria-label={`Zaehlt als Arbeit fuer ${at.name}`}
                        >
                          {at.counts_as_work ? "Ja" : "Nein"}
                        </button>
                      </td>
                      <td className="py-3">
                        <button
                          onClick={() =>
                            handleUpdateAbsenceType(at, { is_active: !at.is_active })
                          }
                          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                            at.is_active
                              ? "bg-green-500/20 text-green-400"
                              : "bg-destructive/20 text-destructive"
                          }`}
                          aria-label={`Status fuer ${at.name}`}
                        >
                          {at.is_active ? "Aktiv" : "Inaktiv"}
                        </button>
                      </td>
                      <td className="py-3">
                        <input
                          type="number"
                          value={at.sort_order}
                          onChange={(e) =>
                            handleUpdateAbsenceType(at, {
                              sort_order: parseInt(e.target.value, 10) || 0,
                            })
                          }
                          className="w-16 rounded border border-border bg-background px-2 py-1 text-sm text-center"
                          min={0}
                          aria-label={`Reihenfolge fuer ${at.name}`}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {absenceTypes.length === 0 && (
                <p className="text-sm text-muted-foreground py-4 text-center">
                  Keine Abwesenheitstypen vorhanden.
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Entitlements Tab */}
      {activeTab === "entitlements" && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 font-heading">
              <CalendarDays className="size-5" />
              Urlaubskontingent zuweisen
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="ent-user" className="mb-1 block text-sm font-medium">
                  Benutzer
                </label>
                <select
                  id="ent-user"
                  value={entUserID}
                  onChange={(e) => setEntUserID(e.target.value)}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                >
                  <option value="">Benutzer waehlen...</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.display_name} ({u.email})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="ent-year" className="mb-1 block text-sm font-medium">
                  Jahr
                </label>
                <input
                  id="ent-year"
                  type="number"
                  value={entYear}
                  onChange={(e) => setEntYear(parseInt(e.target.value, 10) || new Date().getFullYear())}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                  min={2020}
                  max={2099}
                />
              </div>
              <div>
                <label htmlFor="ent-total" className="mb-1 block text-sm font-medium">
                  Urlaubstage gesamt
                </label>
                <input
                  id="ent-total"
                  type="number"
                  value={entTotalDays}
                  onChange={(e) => setEntTotalDays(parseInt(e.target.value, 10) || 0)}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                  min={0}
                  max={365}
                />
              </div>
              <div>
                <label htmlFor="ent-carry" className="mb-1 block text-sm font-medium">
                  Resturlaub Vorjahr
                </label>
                <input
                  id="ent-carry"
                  type="number"
                  value={entCarryOver}
                  onChange={(e) => setEntCarryOver(parseInt(e.target.value, 10) || 0)}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                  min={0}
                  max={365}
                />
              </div>
            </div>
            <div className="mt-4">
              <Button onClick={handleSaveEntitlement} disabled={!entUserID}>
                <Save className="mr-1 size-4" />
                Kontingent speichern
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Work Schedules Tab */}
      {activeTab === "schedules" && (
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 font-heading">
              <Clock className="size-5" />
              Arbeitszeitmodell zuweisen
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="ws-user" className="mb-1 block text-sm font-medium">
                  Benutzer
                </label>
                <select
                  id="ws-user"
                  value={wsUserID}
                  onChange={(e) => setWsUserID(e.target.value)}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                >
                  <option value="">Benutzer waehlen...</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.display_name} ({u.email})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="ws-from" className="mb-1 block text-sm font-medium">
                  Gueltig ab
                </label>
                <input
                  id="ws-from"
                  type="date"
                  value={wsValidFrom}
                  onChange={(e) => setWsValidFrom(e.target.value)}
                  className="w-full rounded border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="sm:col-span-2">
                <label htmlFor="ws-weekly" className="mb-1 block text-sm font-medium">
                  Wochenstunden
                </label>
                <input
                  id="ws-weekly"
                  type="number"
                  value={wsWeekly}
                  onChange={(e) => setWsWeekly(parseFloat(e.target.value) || 0)}
                  className="w-32 rounded border border-border bg-background px-3 py-2 text-sm"
                  min={0}
                  max={168}
                  step={0.5}
                />
              </div>
              <div className="sm:col-span-2">
                <p className="mb-2 text-sm font-medium">Stunden pro Tag</p>
                <div className="grid grid-cols-7 gap-2">
                  {dayLabels.map((label, idx) => (
                    <div key={label} className="flex flex-col items-center gap-1">
                      <label htmlFor={`ws-day-${idx}`} className="text-xs text-muted-foreground">
                        {label}
                      </label>
                      <input
                        id={`ws-day-${idx}`}
                        type="number"
                        value={wsDays[idx]}
                        onChange={(e) => {
                          const next = [...wsDays];
                          next[idx] = parseFloat(e.target.value) || 0;
                          setWsDays(next);
                        }}
                        className="w-full rounded border border-border bg-background px-1 py-2 text-sm text-center"
                        min={0}
                        max={24}
                        step={0.5}
                      />
                    </div>
                  ))}
                </div>
              </div>
            </div>
            <div className="mt-4">
              <Button onClick={handleSaveSchedule} disabled={!wsUserID || !wsValidFrom}>
                <Save className="mr-1 size-4" />
                Arbeitszeitmodell speichern
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
