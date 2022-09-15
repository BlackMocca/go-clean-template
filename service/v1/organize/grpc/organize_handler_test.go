package grpc

import (
	"context"
	"testing"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/Blackmocca/go-clean-template/proto/proto_models"
	_organize_mock "github.com/Blackmocca/go-clean-template/service/v1/organize/mocks"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getSpan() context.Context {
	var ctx = context.Background()
	tracer := opentracing.NoopTracer{}
	span := tracer.StartSpan("test")
	return opentracing.ContextWithSpan(ctx, span)
}

func TestFetchOrganizeById_Success(t *testing.T) {
	var ctx = getSpan()
	orgId := uuid.FromStringOrNil("9c58d4df-a308-4a2e-9cc6-8a36f0b1a7ea")
	createdAt := helperModel.NewTimestampFromString("2019-07-18 07:06:13")
	updatedAt := helperModel.NewTimestampFromString("2020-11-24 04:27:55")

	var org = &models.Organize{
		Id:        &orgId,
		Name:      "สัตว์สี่ขาพร้อมลุย",
		AliasName: zero.StringFrom("สต.4"),
		OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	req := &proto_models.FetchOrganizeByIdRequest{
		Id: orgId.String(),
	}

	orgUs := new(_organize_mock.OrganizeUsecase)
	orgUs.On("FetchOneById", mock.Anything, mock.AnythingOfType("*uuid.UUID")).Return(org, nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		id := args.Get(1).(*uuid.UUID)

		assert.NotNil(t, ctx)
		assert.Equal(t, id.String(), orgId.String())
	})

	handler := NewGRPCOrganizeHandler(orgUs)
	resp, err := handler.FetchOrganizeById(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, resp.Organize.Id, org.Id.String())
	assert.Equal(t, resp.Organize.Name, org.Name)
	assert.Equal(t, resp.Organize.AliasName, org.AliasName.ValueOrZero())
	assert.Equal(t, resp.Organize.CreatedAt, org.CreatedAt.String())
	assert.Equal(t, resp.Organize.UpdatedAt, org.UpdatedAt.String())
}

func TestFetchOrganizeById_NOT_FOUND(t *testing.T) {
	var ctx = getSpan()
	orgId := uuid.FromStringOrNil("9c58d4df-a308-4a2e-9cc6-8a36f0b1a7ea")

	req := &proto_models.FetchOrganizeByIdRequest{
		Id: orgId.String(),
	}

	orgUs := new(_organize_mock.OrganizeUsecase)
	orgUs.On("FetchOneById", mock.Anything, mock.AnythingOfType("*uuid.UUID")).Return(nil, nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		id := args.Get(1).(*uuid.UUID)

		assert.NotNil(t, ctx)
		assert.Equal(t, id.String(), orgId.String())
	})

	handler := NewGRPCOrganizeHandler(orgUs)
	resp, err := handler.FetchOrganizeById(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, status.Code(err), codes.NotFound)
	assert.Nil(t, resp)
}
