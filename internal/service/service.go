package service

import (
	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
)

type Service struct {
	proto.UnimplementedUserServiceServer
	repo *sqlc.Queries
}
