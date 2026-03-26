package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
)

// RequireRole returns middleware that restricts access to users with one of the given global roles.
func RequireRole(roles ...model.UserRole) func(http.Handler) http.Handler {
	allowed := make(map[model.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, `{"error":"Unauthorized","message":"not authenticated"}`, http.StatusUnauthorized)
				return
			}

			if !allowed[user.GlobalRole] {
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("role", string(user.GlobalRole)).
					Str("path", r.URL.Path).
					Msg("access denied: insufficient role")
				http.Error(w, `{"error":"Forbidden","message":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamAccess returns middleware that checks if the user is an admin or a team_leader
// for the team identified by the {teamID} URL parameter.
func RequireTeamAccess(teamRepo *repository.TeamRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, `{"error":"Unauthorized","message":"not authenticated"}`, http.StatusUnauthorized)
				return
			}

			// Admins have full access to all teams
			if user.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			teamIDStr := chi.URLParam(r, "teamID")
			teamID, err := uuid.Parse(teamIDStr)
			if err != nil {
				http.Error(w, `{"error":"Bad Request","message":"invalid team ID"}`, http.StatusBadRequest)
				return
			}

			isLeader, err := teamRepo.IsTeamLeader(r.Context(), teamID, user.ID)
			if err != nil {
				log.Error().Err(err).Str("team_id", teamID.String()).Msg("failed to check team access")
				http.Error(w, `{"error":"Internal Server Error","message":"access check failed"}`, http.StatusInternalServerError)
				return
			}

			if !isLeader {
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("team_id", teamID.String()).
					Msg("access denied: not team leader")
				http.Error(w, `{"error":"Forbidden","message":"insufficient team permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
