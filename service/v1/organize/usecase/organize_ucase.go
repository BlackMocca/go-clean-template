package usecase

import (
	"context"
	"sync"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/Blackmocca/go-clean-template/service/v1/organize"
	"github.com/gofrs/uuid"
)

type organizeUsecase struct {
	organizeRepo organize.OrganizeRepository
}

func NewOrganizeUsecase(orgRepo organize.OrganizeRepository) organize.OrganizeUsecase {
	return &organizeUsecase{
		organizeRepo: orgRepo,
	}
}

func (o organizeUsecase) FetchAll(ctx context.Context, args *sync.Map, paginator *helperModel.Paginator) ([]*models.Organize, error) {
	return o.organizeRepo.FetchAll(ctx, args, paginator)
}

func (o organizeUsecase) FetchOneById(ctx context.Context, orgId *uuid.UUID) (*models.Organize, error) {
	return o.organizeRepo.FetchOneById(ctx, orgId)
}

func (o organizeUsecase) Create(ctx context.Context, organize *models.Organize) error {
	return o.organizeRepo.Create(ctx, organize)
}
