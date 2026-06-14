package user

import (
	"context"
	"example/constants"
	"example/models"

	"github.com/google/uuid"
)

// --- Service Interface (this is what mockery will mock) ---

// Repository defines the data access interface that both the in-memory
// implementation and real database implementations must satisfy.
type Repository interface {
	// FindByID retrieves a user by their unique ID. Returns an error if not found.
	FindByID(ctx context.Context, idOrEmail string) (*models.User, error)
	// Save persists a user to storage. If the user has no ID set (ID == 0),
	// the implementation auto-generates one.
	Save(ctx context.Context, user *models.User) error
	// Delete removes a user by their ID. Returns an error if the user does not exist.
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns all users currently stored. Never returns nil, may return an empty slice.
	List(ctx context.Context, params ListParams) (models.Paginator[[]*models.User], error)
}

// Usecase defines the business logic interface for user operations.
// Repository is injected at construction time and must satisfy
// the Repository interface from this same package.
type Usecase interface {
	// GetUserByID retrieves a single user by their unique ID.
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	// ListUsers retrieves all users. Returns an empty slice (not nil) when no users exist.
	ListUsers(ctx context.Context, params ListParams) (models.Paginator[[]*models.User], error)
	Save(ctx context.Context, user *models.User) error
	// Delete removes a user by their ID. Returns an error if the user does not exist.
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns all users currently stored. Never returns nil, may return an empty slice.
}

type ListParams struct {
	Page       int64          `json:"page"`
	PageSize   int64          `json:"page_size"`
	Searchword string         `json:"search_word"`
	Role       constants.Role `json:"role"`
}
