package usecase

import (
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user"
)

type userUsecase struct {
	psqlUserRepo user.UserRepository
}

func NewUserUsecase(uRepo user.UserRepository) user.UserUsecase {
	return &userUsecase{
		psqlUserRepo: uRepo,
	}
}

func (u userUsecase) FetchAll() ([]*models.User, error) {
	return u.psqlUserRepo.FetchAll()
}

func (u userUsecase) FetchOneById(id int64) (*models.User, error) {
	return u.psqlUserRepo.FetchOneById(id)
}

func (u userUsecase) Create(user *models.User) error {
	return u.psqlUserRepo.Create(user)
}
