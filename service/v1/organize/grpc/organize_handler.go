package grpc

import (
	"context"
	"encoding/json"

	"github.com/Blackmocca/go-clean-template/proto/proto_models"
	"github.com/Blackmocca/go-clean-template/service/v1/organize"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcOrganizeHandler struct {
	orgUs organize.OrganizeUsecase
}

func NewGRPCOrganizeHandler(orgUs organize.OrganizeUsecase) proto_models.OrganizeServer {
	return &grpcOrganizeHandler{
		orgUs: orgUs,
	}
}

func (g grpcOrganizeHandler) FetchOrganizeById(ctx context.Context, req *proto_models.FetchOrganizeByIdRequest) (*proto_models.FetchOrganizeByIdResponse, error) {
	orgId := uuid.FromStringOrNil(req.GetId())

	org, err := g.orgUs.FetchOneById(ctx, &orgId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if org == nil {
		return nil, status.Errorf(codes.NotFound, "organize not found")
	}

	orgData := &proto_models.OrganizeData{}

	bu, _ := json.Marshal(org)
	json.Unmarshal(bu, orgData)

	resp := &proto_models.FetchOrganizeByIdResponse{
		Organize: orgData,
	}
	return resp, nil
}
