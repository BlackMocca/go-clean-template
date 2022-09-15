package organize

import (
	"context"
	"sync"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/gofrs/uuid"
)

type OrganizeUsecase interface {
	FetchAll(ctx context.Context, args *sync.Map, paginator *helperModel.Paginator) ([]*models.Organize, error)
	FetchOneById(ctx context.Context, orgId *uuid.UUID) (*models.Organize, error)
	Create(ctx context.Context, organize *models.Organize) error
}
