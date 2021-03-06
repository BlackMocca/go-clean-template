package user

import (
	"github.com/BlackMocca/go-clean-template/models"
)

type PsqlUserRepositoryInf interface {
	FetchAll() ([]*models.User, error)
	FetchOneById(id int64) (*models.User, error)
	Create(user *models.User) error
}
