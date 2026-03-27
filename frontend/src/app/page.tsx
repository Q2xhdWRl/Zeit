import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import NeuralBackground from "@/components/ui/flow-field-background";
import { Clock, Users, CalendarDays, TrendingUp } from "lucide-react";

function FeatureCard({
  icon: Icon,
  title,
  description,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description: string;
}) {
  return (
    <Card className="glass-card relative overflow-hidden">
      <CardContent className="flex flex-col gap-3 p-6">
        <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10">
          <Icon className="size-5 text-primary" />
        </div>
        <h3 className="font-heading text-lg font-semibold tracking-tight">
          {title}
        </h3>
        <p className="text-sm text-muted-foreground leading-relaxed">
          {description}
        </p>
      </CardContent>
    </Card>
  );
}

export default function LandingPage() {
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

      {/* Content */}
      <div className="relative z-10 mx-auto flex max-w-5xl flex-col items-center gap-12 px-6 py-24">
        {/* Hero */}
        <div className="flex flex-col items-center gap-6 text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-4 py-1.5 text-xs font-medium text-primary">
            <span className="relative flex size-2">
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-primary opacity-75" />
              <span className="relative inline-flex size-2 rounded-full bg-primary" />
            </span>
            NEWA Zeiterfassung
          </div>

          <h1 className="font-heading text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl">
            <span className="text-glow-cyan text-primary">Zeit</span> im Griff.
            <br />
            <span className="text-muted-foreground">Team im Blick.</span>
          </h1>

          <Link href="/login">
            <Button size="lg" className="btn-glow font-heading font-semibold">
              Anmelden mit Microsoft
            </Button>
          </Link>
        </div>

        {/* Features */}
        <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <FeatureCard
            icon={Clock}
            title="Zeiterfassung"
            description="Arbeitszeiten einfach buchen. Tages- und Wochenansicht mit ArbZG-Pruefung."
          />
          <FeatureCard
            icon={CalendarDays}
            title="Abwesenheiten"
            description="Urlaub, Krankheit und Sonderurlaub beantragen und genehmigen."
          />
          <FeatureCard
            icon={Users}
            title="Teamuebersicht"
            description="Verfuegbarkeit im Team auf einen Blick. Wer ist da, wer fehlt."
          />
          <FeatureCard
            icon={TrendingUp}
            title="Ueberstunden"
            description="Soll-Ist-Vergleich und Ueberstundentrends pro Mitarbeiter."
          />
        </div>

        {/* Footer hint */}
        <p className="text-xs text-muted-foreground/60">
          Sichere Anmeldung ueber Microsoft 365 &middot; DSGVO-konform &middot;
          Made for NEWA
        </p>
      </div>
    </main>
  );
}
