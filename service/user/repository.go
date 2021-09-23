package user

import (
	"sync"

	"github.com/BlackMocca/go-clean-template/models"
)

type UserRepository interface {
	FetchAll(args *sync.Map) ([]*models.User, error)
	FetchOneById(id int64) (*models.User, error)
	Create(user *models.User) error
}
