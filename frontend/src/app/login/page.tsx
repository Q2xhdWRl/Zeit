import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function LoginPage() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

  return (
    <main className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden">
      {/* Background gradient effect */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 overflow-hidden"
      >
        <div className="absolute -top-1/2 left-1/2 h-[800px] w-[800px] -translate-x-1/2 rounded-full bg-primary/5 blur-[120px]" />
        <div className="absolute -bottom-1/4 -right-1/4 h-[600px] w-[600px] rounded-full bg-accent/5 blur-[100px]" />
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
      </div>
    </main>
  );
}
