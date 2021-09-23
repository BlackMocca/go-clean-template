package user

import (
	"sync"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/gofrs/uuid"
)

type UserUsecase interface {
	FetchAll(args *sync.Map) ([]*models.User, error)
	FetchOneById(id *uuid.UUID) (*models.User, error)
	Create(user *models.User) error
}
