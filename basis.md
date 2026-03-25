## Stack: Go Backend · Docker · Traefik

## 1. Design 
Kundenname: NEWA
erstelle ein Professionelles Design mit folgenden skills : /frontend-design und /ui-ux-pro-max skill.
tech art, Buttons glow Effekt beim drüber gehen.

## 2. Projektübersicht Zeiterfassungssystem mit Abwesenheitsverwaltung, Übersicht über Teamverfügbarkeit, Überstundenübersicht
Lösung: Maßgeschneiderte Web-App, gehostet auf Kundenserver 
Nutzer: 50–200 Mitarbeiter + Admin(s) 
Zugriff: Browser, kein Install – URL aufrufen, fertig 
M365: Hybrid (on-premise AD + Azure AD), Business Premium 

Das Projekt soll beim Kunden genutzt werden muss somit möglichst sicher sein, nutze den bug-hunter skill und den security-audit skill.

M365 Einbindung sodass die User ihren microsoft account nutzen können für Login, SSO ist wichtig!
Login an landingpage muss case-sensitiv sein, das heisst es wird nicht auf groß/klein schreibweise bei der Email geachtet.
RBAC System beachten, es gibt einen Admin der die anderen berechtigen kann.
Es gibt Teams, Teamleiter die ihre Teams managen und deren Buchungen einsehen können. Normale Benutzer können nur ihr eigenes Team sehen und ihre Zeiten. 

### Migrationsprinzipien

- Jede Phase ist ein eigenständiger PR/Branch
- Keine Phase erfordert die vorherige (aber empfohlen in Reihenfolge)
- Bestehende Funktionalität darf nie brechen
- Tests vor und nach jeder Phase
- Rollback-Plan für jede Phase


