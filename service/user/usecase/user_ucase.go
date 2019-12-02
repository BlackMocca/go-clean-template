package usecase

import (
	"gitlab.com/km/go-kafka-playground/service/user"
)

type userUsecase struct {
	psqlUserRepo user.PsqlUserRepositoryInf
}

func NewUserUsecase(uRepo user.PsqlUserRepositoryInf) user.UserUsecaseInf {
	return &userUsecase{
		psqlUserRepo: uRepo,
	}
}
