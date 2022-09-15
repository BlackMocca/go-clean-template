package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"git.innovasive.co.th/backend/psql"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestFetchAll_WITHOUT_PAGINATOR_Success(t *testing.T) {
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

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"total_rows",
		"organizes.id", "organizes.name", "organizes.alias_name", "organizes.org_type", "organizes.created_at", "organizes.updated_at",
	})

	for _, org := range orgs {
		rows.AddRow(
			len(orgs),
			org.Id, org.Name, org.AliasName, org.OrgType, org.CreatedAt, org.UpdatedAt,
		)
	}

	sql := `
	SELECT
		(.+)
	FROM
		(.+)
	WHERE
		(.+)
	`

	sqlMock.ExpectPrepare(sql).ExpectQuery().WithArgs().WillReturnRows(rows)

	repo := NewPsqlOrganizeRepository(client)
	epOrgs, err := repo.FetchAll(context.Background(), new(sync.Map), nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, epOrgs)

	for index, _ := range epOrgs {
		assert.Equal(t, epOrgs[index].Id, orgs[index].Id)
		assert.Equal(t, epOrgs[index].Name, orgs[index].Name)
		assert.Equal(t, epOrgs[index].AliasName, orgs[index].AliasName)
		assert.Equal(t, epOrgs[index].OrgType, orgs[index].OrgType)
		assert.Equal(t, epOrgs[index].CreatedAt.String(), orgs[index].CreatedAt.String())
		assert.Equal(t, epOrgs[index].UpdatedAt.String(), orgs[index].UpdatedAt.String())
	}
}

func TestFetchAll_PAGINATOR_Success(t *testing.T) {
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

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"total_row",
		"organizes.id", "organizes.name", "organizes.alias_name", "organizes.org_type", "organizes.created_at", "organizes.updated_at",
	})

	for _, org := range orgs {
		rows.AddRow(
			len(orgs),
			org.Id, org.Name, org.AliasName, org.OrgType, org.CreatedAt, org.UpdatedAt,
		)
	}

	sql := `
	SELECT
		(.+)
	FROM
		(.+)
	WHERE
		(.+)
	`

	sqlMock.ExpectPrepare(sql).ExpectQuery().WithArgs().WillReturnRows(rows)

	repo := NewPsqlOrganizeRepository(client)

	paginator := helperModel.NewPaginator()
	paginator.PerPage = 2

	args := new(sync.Map)
	args.Store("show_disabled", "false")

	epOrgs, err := repo.FetchAll(context.Background(), args, &paginator)

	assert.NoError(t, err)
	assert.NotEmpty(t, epOrgs)
	assert.Equal(t, paginator.TotalPages, 2)
	assert.Equal(t, paginator.TotalEntrySizes, 3)

	for index, _ := range epOrgs {
		assert.Equal(t, epOrgs[index].Id, orgs[index].Id)
		assert.Equal(t, epOrgs[index].Name, orgs[index].Name)
		assert.Equal(t, epOrgs[index].AliasName, orgs[index].AliasName)
		assert.Equal(t, epOrgs[index].OrgType, orgs[index].OrgType)
		assert.Equal(t, epOrgs[index].CreatedAt.String(), orgs[index].CreatedAt.String())
		assert.Equal(t, epOrgs[index].UpdatedAt.String(), orgs[index].UpdatedAt.String())
	}
}

func TestFetchOneById_Success(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
	orgId1 := uuid.FromStringOrNil("042e37e5-3027-4499-9a02-91ade81f2d67")
	var admin1 = uuid.FromStringOrNil("62f63e0f-2009-4921-ad77-2bc8d3086c23")
	var admin2 = uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")

	org := &models.Organize{
		Id:        &orgId1,
		Name:      "หน่วยอนุรักษ์สัตว์น้ำ",
		AliasName: zero.StringFrom("อุ่งๆ"),
		OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
		OrderNo:   0,
		Configs: []*models.OrganizesConfig{
			&models.OrganizesConfig{
				OrganizeId:  &orgId1,
				ConfigKey:   "admin_1",
				ConfigValue: admin1.String(),
			},
			&models.OrganizesConfig{
				OrganizeId:  &orgId1,
				ConfigKey:   "admin_2",
				ConfigValue: admin2.String(),
			},
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"organizes.id", "organizes.name", "organizes.alias_name", "organizes.org_type", "organizes.created_at", "organizes.updated_at", "organize_configs.organize_id", "organize_configs.config_key", "organize_configs.config_value",
	})

	for _, orgConfig := range org.Configs {
		rows.AddRow(
			org.Id, org.Name, org.AliasName, org.OrgType, org.CreatedAt, org.UpdatedAt, orgConfig.OrganizeId, orgConfig.ConfigKey, orgConfig.ConfigValue,
		)
	}

	sql := `
	SELECT
		(.+)
	FROM
		(.+)
	WHERE
		(.+)
	`

	sqlMock.ExpectPrepare(sql).ExpectQuery().WithArgs().WillReturnRows(rows)

	repo := NewPsqlOrganizeRepository(client)
	epOrg, err := repo.FetchOneById(context.Background(), &orgId1)

	assert.NoError(t, err)
	assert.NotNil(t, epOrg)

	assert.Equal(t, epOrg.Id, org.Id)
	assert.Equal(t, epOrg.Name, org.Name)
	assert.Equal(t, epOrg.AliasName, org.AliasName)
	assert.Equal(t, epOrg.OrgType, org.OrgType)
	assert.Equal(t, epOrg.CreatedAt.String(), org.CreatedAt.String())
	assert.Equal(t, epOrg.UpdatedAt.String(), org.UpdatedAt.String())
}

func TestFetchOneById_NOT_FOUND(t *testing.T) {
	orgId1 := uuid.FromStringOrNil("907eefd8-181b-457b-8ca2-692c442b2b0b")

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"organizes.id", "organizes.name", "organizes.alias_name", "organizes.org_type", "organizes.created_at", "organizes.updated_at",
	})

	sql := `
	SELECT
		(.+)
	FROM
		(.+)
	WHERE
		(.+)
	`

	sqlMock.ExpectPrepare(sql).ExpectQuery().WithArgs().WillReturnRows(rows)

	repo := NewPsqlOrganizeRepository(client)

	epOrg, err := repo.FetchOneById(context.Background(), &orgId1)

	assert.NoError(t, err)
	assert.Nil(t, epOrg)

}

func TestCreate_Success(t *testing.T) {
	id := uuid.FromStringOrNil("81914106-9919-4c4f-be96-7d9258336556")
	name := "หน่วยทำลายเครื่องดืมมึนเมาแบบจู่โจม"
	aliasName := zero.StringFrom("ทม.")
	var admin1 = uuid.FromStringOrNil("62f63e0f-2009-4921-ad77-2bc8d3086c23")
	var admin2 = uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")
	now := helperModel.NewTimestampFromTime(time.Now())
	var organize = &models.Organize{
		Id:        &id,
		Name:      name,
		AliasName: aliasName,
		OrgType:   "PUBLIC",
		OrderNo:   45,
		Configs: []*models.OrganizesConfig{
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_1",
				ConfigValue: admin1.String(),
			},
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_2",
				ConfigValue: admin2.String(),
			},
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	sqlInsertOrg := `INSERT INTO organizes`
	sqlInsertConfig := `INSERT INTO organize_configs`

	sqlMock.ExpectBegin()

	sqlMock.ExpectPrepare(sqlInsertOrg).ExpectExec().WithArgs(
		organize.Id,
		organize.Name,
		organize.AliasName,
		organize.OrgType,
		organize.CreatedAt,
		organize.UpdatedAt,
		organize.OrderNo,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	configM := sqlMock.ExpectPrepare(sqlInsertConfig)
	for _, config := range organize.GetOrganizeConfig() {
		configM.ExpectExec().WithArgs(
			organize.Id,
			config.ConfigKey,
			config.ConfigValue,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	sqlMock.ExpectCommit()

	repo := NewPsqlOrganizeRepository(client)
	epErr := repo.Create(context.Background(), organize)

	assert.NoError(t, epErr)
}

func TestCreate_ErrorOrgNameDuplicate(t *testing.T) {
	id := uuid.FromStringOrNil("81914106-9919-4c4f-be96-7d9258336556")
	name := "หน่วยทำลายเครื่องดืมมึนเมาแบบจู่โจม"
	aliasName := zero.StringFrom("ทม.")
	var admin1 = uuid.FromStringOrNil("62f63e0f-2009-4921-ad77-2bc8d3086c23")
	var admin2 = uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")
	now := helperModel.NewTimestampFromTime(time.Now())
	var organize = &models.Organize{
		Id:        &id,
		Name:      name,
		AliasName: aliasName,
		OrgType:   "PUBLIC",
		OrderNo:   45,
		Configs: []*models.OrganizesConfig{
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_1",
				ConfigValue: admin1.String(),
			},
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_2",
				ConfigValue: admin2.String(),
			},
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	sqlInsertOrg := `INSERT INTO organizes`

	sqlMock.ExpectBegin()
	sqlMock.ExpectPrepare(sqlInsertOrg).ExpectExec().WithArgs(
		organize.Id,
		organize.Name,
		organize.AliasName,
		organize.OrgType,
		organize.CreatedAt,
		organize.UpdatedAt,
		organize.OrderNo,
	).WillReturnError(errors.New(constants.ERROR_UNIQUE_ORGANIZE_NAME))

	sqlMock.ExpectCommit()
	repo := NewPsqlOrganizeRepository(client)
	epErr := repo.Create(context.Background(), organize)

	assert.Error(t, epErr)
	assert.Equal(t, epErr.Error(), constants.ERROR_ORGANIZE_NAME_WAS_DUPLICATE)
}

func TestCreate_ErrorOrgAliasNameDuplicate(t *testing.T) {
	id := uuid.FromStringOrNil("81914106-9919-4c4f-be96-7d9258336556")
	name := "หน่วยทำลายเครื่องดืมมึนเมาแบบจู่โจม"
	aliasName := zero.StringFrom("ทม.")
	var admin1 = uuid.FromStringOrNil("62f63e0f-2009-4921-ad77-2bc8d3086c23")
	var admin2 = uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")
	now := helperModel.NewTimestampFromTime(time.Now())
	var organize = &models.Organize{
		Id:        &id,
		Name:      name,
		AliasName: aliasName,
		OrgType:   "PUBLIC",
		OrderNo:   45,
		Configs: []*models.OrganizesConfig{
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_1",
				ConfigValue: admin1.String(),
			},
			&models.OrganizesConfig{
				OrganizeId:  &id,
				ConfigKey:   "admin_2",
				ConfigValue: admin2.String(),
			},
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	client := new(psql.Client)
	client.SetDB(sqlxDB)

	sqlInsertOrg := `INSERT INTO organizes`

	sqlMock.ExpectBegin()

	sqlMock.ExpectPrepare(sqlInsertOrg).ExpectExec().WithArgs(
		organize.Id,
		organize.Name,
		organize.AliasName,
		organize.OrgType,
		organize.CreatedAt,
		organize.UpdatedAt,
		organize.OrderNo,
	).WillReturnError(errors.New(constants.ERROR_UNIQUE_ORGANIZE_ALIAS_NAME))

	sqlMock.ExpectCommit()

	repo := NewPsqlOrganizeRepository(client)
	epErr := repo.Create(context.Background(), organize)

	assert.Error(t, epErr)
	assert.Equal(t, epErr.Error(), constants.ERROR_ORGANIZE_ALIAS_NAME_WAS_DUPLICATE)
}
