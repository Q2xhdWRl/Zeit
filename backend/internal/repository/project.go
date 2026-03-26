package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/newa/zeiterfassung/internal/model"
)

// ProjectRepository handles database operations for projects.
type ProjectRepository struct {
	db *pgxpool.Pool
}

// NewProjectRepository creates a new ProjectRepository.
func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

const projectColumns = `id, name, customer_name, is_active, created_at`

func scanProject(row pgx.Row) (*model.Project, error) {
	var p model.Project
	err := row.Scan(&p.ID, &p.Name, &p.CustomerName, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, name, customerName string) (*model.Project, error) {
	return scanProject(r.db.QueryRow(ctx,
		`INSERT INTO projects (name, customer_name) VALUES ($1, $2) RETURNING `+projectColumns,
		name, customerName,
	))
}

// FindByID returns a project by its UUID.
func (r *ProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	return scanProject(r.db.QueryRow(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE id = $1`, id,
	))
}

// ListAll returns all projects ordered by name.
func (r *ProjectRepository) ListAll(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT `+projectColumns+` FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CustomerName, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// ListActive returns all active projects ordered by name.
func (r *ProjectRepository) ListActive(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT `+projectColumns+` FROM projects WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CustomerName, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// Update modifies a project.
func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, name, customerName string, isActive bool) (*model.Project, error) {
	return scanProject(r.db.QueryRow(ctx,
		`UPDATE projects SET name = $1, customer_name = $2, is_active = $3 WHERE id = $4 RETURNING `+projectColumns,
		name, customerName, isActive, id,
	))
}
