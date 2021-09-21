package usecase

import (
	"errors"
	"testing"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchAll(t *testing.T) {
	mockPsqlUserRepo := new(mocks.PsqlUserRepositoryInf)
	mockUser := models.User{
		ID:        1,
		Email:     "qwerty@gmail.com",
		Firstname: "Teeradet",
		Lastname:  "Phondetparinya",
		Age:       25,
	}

	mockUsers := make([]*models.User, 0)
	mockUsers = append(mockUsers, &mockUser)

	t.Run("success", func(t *testing.T) {
		mockPsqlUserRepo.On("FetchAll").Return(mockUsers, nil).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		list, err := u.FetchAll()

		assert.NoError(t, err)
		assert.NotNil(t, list)
		assert.Len(t, list, len(mockUsers))

		mockPsqlUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockPsqlUserRepo.On("FetchAll").Return(nil, errors.New("Unexpected")).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		list, err := u.FetchAll()
		assert.Error(t, err)
		assert.Nil(t, list)

		mockPsqlUserRepo.AssertExpectations(t)
	})
}

func TestFetchOneById(t *testing.T) {
	mockPsqlUserRepo := new(mocks.PsqlUserRepositoryInf)
	mockUser := models.User{
		ID:        1,
		Email:     "qwerty@gmail.com",
		Firstname: "Teeradet",
		Lastname:  "Phondetparinya",
		Age:       25,
	}

	t.Run("success", func(t *testing.T) {
		mockPsqlUserRepo.On("FetchOneById", mock.AnythingOfType("int64")).Return(&mockUser, nil).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		a, err := u.FetchOneById(mockUser.ID)

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockPsqlUserRepo.AssertExpectations(t)
	})
	t.Run("error-notfound", func(t *testing.T) {
		mockPsqlUserRepo.On("FetchOneById", mock.AnythingOfType("int64")).Return(nil, nil).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		a, err := u.FetchOneById(mockUser.ID)

		assert.NoError(t, err)
		assert.Nil(t, a)

		mockPsqlUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockPsqlUserRepo.On("FetchOneById", mock.AnythingOfType("int64")).Return(nil, errors.New("Unexpected")).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		a, err := u.FetchOneById(mockUser.ID)

		assert.Error(t, err)
		assert.Nil(t, a)

		mockPsqlUserRepo.AssertExpectations(t)
	})
}

func TestCreate(t *testing.T) {
	mockPsqlUserRepo := new(mocks.PsqlUserRepositoryInf)
	mockUser := models.User{
		Email:     "qwerty@gmail.com",
		Firstname: "Teeradet",
		Lastname:  "Phondetparinya",
		Age:       25,
	}

	t.Run("success", func(t *testing.T) {
		mockPsqlUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		err := u.Create(&mockUser)

		assert.NoError(t, err)

		mockPsqlUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockPsqlUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("Unexpected")).Once()
		u := NewUserUsecase(mockPsqlUserRepo)

		err := u.Create(&mockUser)

		assert.Error(t, err)

		mockPsqlUserRepo.AssertExpectations(t)
	})
}
