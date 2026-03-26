import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import NeuralBackground from "@/components/ui/flow-field-background";

const DEV_USERS = [
  { label: "Admin (Anna Admin)", token: "dev-admin-token", role: "admin" },
  { label: "Teamleiter (Lars Leiter)", token: "dev-leader-token", role: "team_leader" },
  { label: "Benutzer (Udo User)", token: "dev-user-token", role: "user" },
];

export default function LoginPage() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";
  const isDev = process.env.NODE_ENV === "development" || apiUrl.includes("localhost");

  return (
    <main className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden">
      {/* Animated flow-field background */}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <NeuralBackground
          color="#00d4ff"
          trailOpacity={0.08}
          speed={0.8}
          particleCount={600}
        />
      </div>

      {/* Login Card */}
      <div className="relative z-10 mx-auto flex max-w-md flex-col items-center gap-8 px-6">
        <div className="flex flex-col items-center gap-3 text-center">
          <Badge variant="outline" className="border-primary/20 bg-primary/5 text-primary">
            NEWA Zeiterfassung
          </Badge>
          <h1 className="font-heading text-3xl font-bold tracking-tight">
            <span className="text-glow-cyan text-primary">Anmelden</span>
          </h1>
        </div>

        <Card className="glass-card w-full">
          <CardHeader className="text-center">
            <CardTitle className="font-heading text-xl">Willkommen</CardTitle>
            <CardDescription>
              Melden Sie sich mit Ihrem Microsoft-Konto an, um fortzufahren.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <a href={`${apiUrl}/auth/login`} className="w-full">
              <Button
                size="lg"
                className="btn-glow w-full font-heading font-semibold"
              >
                <svg
                  className="mr-2 size-5"
                  viewBox="0 0 21 21"
                  fill="none"
                  aria-hidden="true"
                >
                  <rect x="1" y="1" width="9" height="9" fill="#f25022" />
                  <rect x="11" y="1" width="9" height="9" fill="#7fba00" />
                  <rect x="1" y="11" width="9" height="9" fill="#00a4ef" />
                  <rect x="11" y="11" width="9" height="9" fill="#ffb900" />
                </svg>
                Anmelden mit Microsoft
              </Button>
            </a>
            <p className="text-center text-xs text-muted-foreground">
              Sichere Anmeldung ueber Microsoft 365 &middot; DSGVO-konform
            </p>
          </CardContent>
        </Card>

        {/* Dev-only quick login buttons */}
        {isDev && (
          <Card className="glass-card w-full border-amber-500/20">
            <CardHeader className="pb-2 text-center">
              <CardTitle className="flex items-center justify-center gap-2 text-sm font-medium text-amber-400">
                <Badge variant="outline" className="border-amber-500/30 bg-amber-500/10 text-amber-400">
                  DEV
                </Badge>
                Schnellanmeldung
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {DEV_USERS.map((user) => (
                <a
                  key={user.token}
                  href={`${apiUrl}/auth/dev-login?token=${user.token}`}
                  className="w-full"
                >
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full justify-start border-border/50 text-sm"
                  >
                    <Badge
                      variant={
                        user.role === "admin"
                          ? "default"
                          : user.role === "team_leader"
                            ? "secondary"
                            : "outline"
                      }
                      className="mr-2 text-[10px]"
                    >
                      {user.role === "admin"
                        ? "Admin"
                        : user.role === "team_leader"
                          ? "TL"
                          : "User"}
                    </Badge>
                    {user.label}
                  </Button>
                </a>
              ))}
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
