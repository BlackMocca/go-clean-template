// Package user defines the use-case layer for user operations.
// It depends on a Repository interface so that real and mock
// implementations can be swapped during testing.
package user

import (
	"context"
	"errors"

	"example/constants"
	"example/models"

	"github.com/google/uuid"
)

// userUsecase implements the Usecase interface using a Repository for data access.
type userUsecase struct {
	repo Repository // injected dependency; nil-safe validation is done in each method
}

// NewUserUsecase constructs a new Usecase with the given repository.
// Pass a mock Repository in tests to isolate business logic from data access.
func NewUserUsecase(repo Repository) Usecase {
	return &userUsecase{repo: repo}
}

// GetUserByID validates the ID and retrieves a single user via the repository.
func (s *userUsecase) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(ctx, id.String())
}

// ListUsers delegates to the repository and returns all users.
func (s *userUsecase) ListUsers(ctx context.Context, params ListParams) (models.Paginator[[]*models.User], error) {
	return s.repo.List(ctx, params)
}

func (s *userUsecase) Save(ctx context.Context, user *models.User) error {
	existsUser, err := s.repo.FindByID(ctx, user.Email.ValueOrZero())
	if err != nil {
		return err
	}
	if existsUser != nil {
		return errors.New(string(constants.ApplicationCodeConflictUserEmailWasDuplicate))
	}
	return s.repo.Save(ctx, user)
}

// Delete removes a user by their ID. Returns an error if the user does not exist.
func (s *userUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, id.String())
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(string(constants.ApplicationCodeConflictUserAlreadyRemoved))
	}
	return s.repo.Delete(ctx, id)
}
