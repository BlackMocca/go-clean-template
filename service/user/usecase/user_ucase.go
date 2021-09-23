package usecase

import (
	"sync"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/gofrs/uuid"
)

type userUsecase struct {
	psqlUserRepo user.UserRepository
}

func NewUserUsecase(uRepo user.UserRepository) user.UserUsecase {
	return &userUsecase{
		psqlUserRepo: uRepo,
	}
}

func (u userUsecase) FetchAll(args *sync.Map) ([]*models.User, error) {
	return u.psqlUserRepo.FetchAll(args)
}

func (u userUsecase) FetchOneById(id *uuid.UUID) (*models.User, error) {
	return u.psqlUserRepo.FetchOneById(id)
}

func (u userUsecase) Create(user *models.User) error {
	return u.psqlUserRepo.Create(user)
}
