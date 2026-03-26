package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/config"
	"github.com/newa/zeiterfassung/internal/database"
	"github.com/newa/zeiterfassung/internal/handler"
	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
	"github.com/newa/zeiterfassung/internal/service"
	"github.com/newa/zeiterfassung/migrations"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	if cfg.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	db, err := database.New(cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Auto-migrate: run all pending .up.sql migrations.
	if err := database.Migrate(db, migrations.FS); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Dev seed: populate test data in development mode.
	if cfg.Env == "development" && os.Getenv("DEV_SEED") == "true" {
		if err := database.Seed(db, migrations.DevSeedSQL); err != nil {
			log.Warn().Err(err).Msg("dev seed failed (may already be applied)")
		}
	}

	healthHandler := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	timeEntryRepo := repository.NewTimeEntryRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	absenceRepo := repository.NewAbsenceRepository(db)
	scheduleRepo := repository.NewWorkScheduleRepository(db)
	stampRepo := repository.NewStampRepository(db)

	authService := service.NewAuthService(cfg, userRepo, sessionRepo)
	timeEntryService := service.NewTimeEntryService(timeEntryRepo, projectRepo)
	absenceService := service.NewAbsenceService(absenceRepo)
	overtimeService := service.NewOvertimeService(scheduleRepo, timeEntryRepo, absenceRepo, teamRepo, userRepo)

	authHandler := handler.NewAuthHandler(authService, cfg)
	adminHandler := handler.NewAdminHandler(userRepo, sessionRepo)
	teamHandler := handler.NewTeamHandler(teamRepo)
	timeEntryHandler := handler.NewTimeEntryHandler(timeEntryService, timeEntryRepo)
	projectHandler := handler.NewProjectHandler(projectRepo)
	absenceHandler := handler.NewAbsenceHandler(absenceService, absenceRepo, teamRepo)
	overtimeHandler := handler.NewOvertimeHandler(overtimeService, scheduleRepo)
	stampHandler := handler.NewStampHandler(stampRepo, timeEntryService)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.CORSAllowedURLs))
	r.Use(middleware.SecurityHeaders)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(w, req.Body, 1<<20) // 1 MB
			next.ServeHTTP(w, req)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler.Check)

		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", authHandler.Login)
			r.Get("/callback", authHandler.Callback)
			r.Post("/logout", authHandler.Logout)

			// Dev-only: quick login with pre-seeded tokens (never available in production).
			if cfg.Env == "development" {
				r.Get("/dev-login", authHandler.DevLogin)
			}
		})

		// Protected routes (any authenticated user)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Get("/auth/me", authHandler.Me)
			r.Get("/teams", teamHandler.ListTeams)
			r.Get("/teams/my", teamHandler.MyTeams)
			r.Get("/teams/{teamID}", teamHandler.GetTeam)
			r.Get("/teams/{teamID}/members", teamHandler.ListMembers)

			// Time entries (own)
			r.Post("/time-entries", timeEntryHandler.Create)
			r.Get("/time-entries", timeEntryHandler.ListMy)
			r.Get("/time-entries/summary", timeEntryHandler.Summary)
			r.Put("/time-entries/{entryID}", timeEntryHandler.Update)
			r.Delete("/time-entries/{entryID}", timeEntryHandler.Delete)

			// Projects (active only for regular users)
			r.Get("/projects", projectHandler.List)

			// Absences (own)
			r.Post("/absences", absenceHandler.Create)
			r.Get("/absences", absenceHandler.ListMy)
			r.Get("/absences/balance", absenceHandler.VacationBalance)
			r.Put("/absences/{absenceID}", absenceHandler.Update)
			r.Delete("/absences/{absenceID}", absenceHandler.Delete)
			r.Post("/absences/{absenceID}/cancel", absenceHandler.Cancel)

			// Absence types (active only)
			r.Get("/absence-types", absenceHandler.ListAbsenceTypes)

			// Overtime & Dashboard
			r.Get("/overtime", overtimeHandler.OvertimeSummary)
			r.Get("/overtime/trend", overtimeHandler.OvertimeTrend)
			r.Get("/dashboard", overtimeHandler.Dashboard)
			r.Get("/work-schedule", overtimeHandler.GetSchedule)

			// Stempeluhr
			r.Get("/stamp/active", stampHandler.GetActive)
			r.Post("/stamp/in", stampHandler.StampIn)
			r.Post("/stamp/out", stampHandler.StampOut)
			r.Post("/stamp/break", stampHandler.ToggleBreak)
		})

		// Team leader routes (team-scoped access — all have {teamID} in URL)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Use(middleware.RequireTeamAccess(teamRepo))
			r.Post("/teams/{teamID}/members", teamHandler.AddMember)
			r.Delete("/teams/{teamID}/members/{userID}", teamHandler.RemoveMember)
			r.Get("/time-entries/team/{teamID}", timeEntryHandler.ListByTeam)
			r.Get("/absences/team/{teamID}", absenceHandler.ListByTeam)
			r.Get("/absences/team/{teamID}/pending", absenceHandler.ListPending)
			r.Get("/teams/{teamID}/availability", overtimeHandler.TeamAvailability)
		})

		// Absence review — accessible to admins and team leaders; handler checks team membership.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Use(middleware.RequireRole(model.RoleAdmin, model.RoleTeamLeader))
			r.Put("/absences/{absenceID}/review", absenceHandler.Review)
		})

		// Admin-only routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Use(middleware.RequireRole(model.RoleAdmin))

			// User management
			r.Get("/users", adminHandler.ListUsers)
			r.Get("/users/{userID}", adminHandler.GetUser)
			r.Put("/users/{userID}/role", adminHandler.UpdateRole)
			r.Put("/users/{userID}/active", adminHandler.UpdateActive)

			// Team management (create/update/delete)
			r.Post("/teams", teamHandler.CreateTeam)
			r.Put("/teams/{teamID}", teamHandler.UpdateTeam)
			r.Delete("/teams/{teamID}", teamHandler.DeleteTeam)

			// Project management
			r.Get("/projects", projectHandler.ListAll)
			r.Post("/projects", projectHandler.Create)
			r.Put("/projects/{projectID}", projectHandler.Update)

			// Absence type management
			r.Get("/absence-types", absenceHandler.ListAllAbsenceTypes)
			r.Put("/absence-types/{typeID}", absenceHandler.UpdateAbsenceType)

			// Vacation entitlement management
			r.Put("/entitlements", absenceHandler.UpsertEntitlement)

			// Work schedule management
			r.Put("/work-schedules", overtimeHandler.UpsertSchedule)
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Int("port", cfg.Port).Str("env", cfg.Env).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server stopped gracefully")
}
