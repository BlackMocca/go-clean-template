package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/models"
	_mock_organize "github.com/Blackmocca/go-clean-template/service/v1/organize/mocks"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchAll_Success(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
	orgId1 := uuid.FromStringOrNil("907eefd8-181b-457b-8ca2-692c442b2b0b")
	orgId2 := uuid.FromStringOrNil("97478e1b-2ebd-4dee-88da-49da3ca482f4")
	orgId3 := uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")
	orgs := []*models.Organize{
		&models.Organize{
			Id:        &orgId1,
			Name:      "จราจรการสื่อสาร",
			AliasName: zero.StringFrom("จส100"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		&models.Organize{
			Id:        &orgId2,
			Name:      "ตุ้ดซี่ review",
			AliasName: zero.StringFrom("ตซ"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		&models.Organize{
			Id:        &orgId3,
			Name:      "เกมถูกบอกด้วย",
			AliasName: zero.StringFrom("เกมถูกบอกด้วย"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	args := new(sync.Map)
	paginator := helperModel.NewPaginator()
	paginator.PerPage = 2

	orgRepo := new(_mock_organize.OrganizeRepository)
	us := NewOrganizeUsecase(orgRepo)
	orgRepo.On("FetchAll", mock.Anything, mock.AnythingOfType("*sync.Map"), mock.AnythingOfType("*models.Paginator")).Return(orgs, nil)
	organizes, err := us.FetchAll(context.Background(), args, &paginator)

	assert.NoError(t, err)
	assert.NotEmpty(t, organizes)
}

func TestFetchOneById_Success(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
	orgId1 := uuid.FromStringOrNil("042e37e5-3027-4499-9a02-91ade81f2d67")

	org := &models.Organize{
		Id:        &orgId1,
		Name:      "หน่วยทำลายเครื่องดื่มชานมเผือกแบบจู่โจม",
		AliasName: zero.StringFrom("ทชผจ."),
		OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	orgRepo := new(_mock_organize.OrganizeRepository)
	us := NewOrganizeUsecase(orgRepo)
	orgRepo.On("FetchOneById", mock.Anything, mock.AnythingOfType("*uuid.UUID")).Return(org, nil)
	reOrg, err := us.FetchOneById(context.Background(), &orgId1)
	assert.NoError(t, err)
	assert.NotEmpty(t, reOrg)
}

func TestCreate_Success(t *testing.T) {
	id := uuid.FromStringOrNil("81914106-9919-4c4f-be96-7d9258336556")
	name := "หน่วยทำลายเครื่องดืมมึนเมาแบบจู่โจม"
	aliasName := zero.StringFrom("ทม.")
	now := helperModel.NewTimestampFromTime(time.Now())
	var organize = &models.Organize{
		Id:        &id,
		Name:      name,
		AliasName: aliasName,
		OrgType:   "PUBLIC",
		OrderNo:   45,
		Configs:   []*models.OrganizesConfig{},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	orgRepo := new(_mock_organize.OrganizeRepository)
	us := NewOrganizeUsecase(orgRepo)
	orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(nil)
	err := us.Create(context.Background(), organize)
	assert.NoError(t, err)
}
