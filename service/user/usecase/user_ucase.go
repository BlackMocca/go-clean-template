package usecase

import (
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user"
)

type userUsecase struct {
	psqlUserRepo user.PsqlUserRepositoryInf
}

func NewUserUsecase(uRepo user.PsqlUserRepositoryInf) user.UserUsecaseInf {
	return &userUsecase{
		psqlUserRepo: uRepo,
	}
}

func (u userUsecase) FetchAll() ([]*models.User, error) {
	return u.psqlUserRepo.FetchAll()
}

func (u userUsecase) Create(user *models.User) error {
	return u.psqlUserRepo.Create(user)
}
