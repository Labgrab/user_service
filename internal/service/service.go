package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
	"labgrab/user_service/pkg/logger"
)

var tracer = otel.Tracer("user-service")

type Service struct {
	proto.UnimplementedUserServiceServer
	Logger *zap.Logger
	Repo   *sqlc.Queries
}

func (s *Service) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	ctx, span := tracer.Start(ctx, "CreateUser")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.Uuid),
		zap.String("method", "CreateUser"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.Uuid),
		attribute.String("grpc.method", "CreateUser"),
	)

	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	createdUUID, err := s.Repo.CreateUser(ctx, userUUID)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to create user: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	return &proto.CreateUserResponse{
		User: &proto.User{
			Uuid: createdUUID.String(),
		},
	}, nil
}

func (s *Service) GetUserDetails(ctx context.Context, req *proto.GetUserDetailsRequest) (*proto.GetUserDetailsResponse, error) {
	ctx, span := tracer.Start(ctx, "GetUserDetails")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "GetUserDetails"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "GetUserDetails"),
	)

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.GetUserDetails(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user details not found")
		}
		log.Error("Failed to get user details", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to get user details: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.GetUserDetailsResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) GetUserContacts(ctx context.Context, req *proto.GetUserContactsRequest) (*proto.GetUserContactsResponse, error) {
	ctx, span := tracer.Start(ctx, "GetUserContacts")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "GetUserContacts"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "GetUserContacts"),
	)

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.Repo.GetUserContacts(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user contacts not found")
		}
		log.Error("Failed to get user contacts", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to get user contacts: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.GetUserContactsResponse{
		Contacts: &proto.UserContacts{
			PhoneNumber: contacts.PhoneNumber,
			UserUuid:    contacts.UserUuid.String(),
		},
	}

	if contacts.Email.Valid {
		response.Contacts.Email = &contacts.Email.String
	}

	if contacts.TelegramID.Valid {
		response.Contacts.TelegramId = &contacts.TelegramID.Int64
	}

	return response, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	ctx, span := tracer.Start(ctx, "DeleteUser")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.Uuid),
		zap.String("method", "DeleteUser"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.Uuid),
		attribute.String("grpc.method", "DeleteUser"),
	)

	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	if err := s.Repo.DeleteUser(ctx, userUUID); err != nil {
		log.Error("Failed to delete user", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to delete user: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	return &proto.DeleteUserResponse{}, nil
}

func (s *Service) CreateUserDetails(ctx context.Context, req *proto.CreateUserDetailsRequest) (*proto.CreateUserDetailsResponse, error) {
	ctx, span := tracer.Start(ctx, "CreateUserDetails")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "CreateUserDetails"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "CreateUserDetails"),
		attribute.String("user.group_code", req.GroupCode),
	)

	if !ValidateAlphabeticString(req.Name) {
		span.SetStatus(codes.Error, "invalid name format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid name format")
	}
	if !ValidateAlphabeticString(req.Surname) {
		span.SetStatus(codes.Error, "invalid surname format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid surname format")
	}
	if req.Patronymic != nil && !ValidateAlphabeticString(*req.Patronymic) {
		span.SetStatus(codes.Error, "invalid patronymic format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid patronymic format")
	}
	if !ValidateGroupCode(req.GroupCode) {
		span.SetStatus(codes.Error, "invalid group code format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid group code format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.CreateUserDetailsParams{
		Name:      req.Name,
		Surname:   req.Surname,
		GroupCode: req.GroupCode,
		UserUuid:  userUUID,
	}

	if req.Patronymic != nil {
		params.Patronymic = pgtype.Text(sql.NullString{
			String: *req.Patronymic,
			Valid:  true,
		})
	}

	details, err := s.Repo.CreateUserDetails(ctx, params)
	if err != nil {
		log.Error("Failed to create user details", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to create user details: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.CreateUserDetailsResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) UpdateUserName(ctx context.Context, req *proto.UpdateUserNameRequest) (*proto.UpdateUserNameResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserName")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserName"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserName"),
	)

	if !ValidateAlphabeticString(req.Name) {
		span.SetStatus(codes.Error, "invalid name format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid name format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserName(ctx, sqlc.UpdateUserNameParams{
		UserUuid: userUUID,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user details not found")
		}
		log.Error("Failed to update user name", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update user name: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserNameResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) UpdateUserSurname(ctx context.Context, req *proto.UpdateUserSurnameRequest) (*proto.UpdateUserSurnameResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserSurname")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserSurname"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserSurname"),
	)

	if !ValidateAlphabeticString(req.Surname) {
		span.SetStatus(codes.Error, "invalid surname format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid surname format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserSurname(ctx, sqlc.UpdateUserSurnameParams{
		UserUuid: userUUID,
		Surname:  req.Surname,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user details not found")
		}
		log.Error("Failed to update user surname", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update user surname: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserSurnameResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) UpdateUserPatronymic(ctx context.Context, req *proto.UpdateUserPatronymicRequest) (*proto.UpdateUserPatronymicResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserPatronymic")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserPatronymic"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserPatronymic"),
	)

	if !ValidateAlphabeticString(req.Patronymic) {
		span.SetStatus(codes.Error, "invalid patronymic format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid patronymic format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserPatronymicParams{
		UserUuid: userUUID,
		Patronymic: pgtype.Text(sql.NullString{
			String: req.Patronymic,
			Valid:  true,
		}),
	}

	details, err := s.Repo.UpdateUserPatronymic(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user details not found")
		}
		log.Error("Failed to update user patronymic", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update user patronymic: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserPatronymicResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) UpdateUserGroupCode(ctx context.Context, req *proto.UpdateUserGroupCodeRequest) (*proto.UpdateUserGroupCodeResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserGroupCode")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserGroupCode"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserGroupCode"),
		attribute.String("user.group_code", req.GroupCode),
	)

	if !ValidateGroupCode(req.GroupCode) {
		span.SetStatus(codes.Error, "invalid group code format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid group code format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserGroupCode(ctx, sqlc.UpdateUserGroupCodeParams{
		UserUuid:  userUUID,
		GroupCode: req.GroupCode,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user details not found")
		}
		log.Error("Failed to update user group code", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update user group code: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserGroupCodeResponse{
		Details: &proto.UserDetails{
			Name:      details.Name,
			Surname:   details.Surname,
			GroupCode: details.GroupCode,
			UserUuid:  details.UserUuid.String(),
		},
	}

	if details.Patronymic.Valid {
		response.Details.Patronymic = &details.Patronymic.String
	}

	return response, nil
}

func (s *Service) CreateUserContacts(ctx context.Context, req *proto.CreateUserContactsRequest) (*proto.CreateUserContactsResponse, error) {
	ctx, span := tracer.Start(ctx, "CreateUserContacts")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "CreateUserContacts"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "CreateUserContacts"),
	)

	if !ValidatePhoneNumber(req.PhoneNumber) {
		span.SetStatus(codes.Error, "invalid phone number format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid phone number format")
	}

	if req.TelegramId != nil && *req.TelegramId <= 0 {
		span.SetStatus(codes.Error, "invalid telegram ID")
		return nil, status.Errorf(grpccodes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.CreateUserContactsParams{
		PhoneNumber: req.PhoneNumber,
		UserUuid:    userUUID,
	}

	if req.Email != nil {
		params.Email = pgtype.Text(sql.NullString{
			String: *req.Email,
			Valid:  true,
		})
	}

	if req.TelegramId != nil {
		params.TelegramID = pgtype.Int8(sql.NullInt64{
			Int64: *req.TelegramId,
			Valid: true,
		})
	}

	contacts, err := s.Repo.CreateUserContacts(ctx, params)
	if err != nil {
		log.Error("Failed to create user contacts", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to create user contacts: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.CreateUserContactsResponse{
		Contacts: &proto.UserContacts{
			PhoneNumber: contacts.PhoneNumber,
			UserUuid:    contacts.UserUuid.String(),
		},
	}

	if contacts.Email.Valid {
		response.Contacts.Email = &contacts.Email.String
	}

	if contacts.TelegramID.Valid {
		response.Contacts.TelegramId = &contacts.TelegramID.Int64
	}

	return response, nil
}

func (s *Service) UpdateUserPhoneNumber(ctx context.Context, req *proto.UpdateUserPhoneNumberRequest) (*proto.UpdateUserPhoneNumberResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserPhoneNumber")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserPhoneNumber"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserPhoneNumber"),
	)

	if !ValidatePhoneNumber(req.PhoneNumber) {
		span.SetStatus(codes.Error, "invalid phone number format")
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid phone number format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.Repo.UpdateUserPhoneNumber(ctx, sqlc.UpdateUserPhoneNumberParams{
		UserUuid:    userUUID,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update phone number", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update phone number: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserPhoneNumberResponse{
		Contacts: &proto.UserContacts{
			PhoneNumber: contacts.PhoneNumber,
			UserUuid:    contacts.UserUuid.String(),
		},
	}

	if contacts.Email.Valid {
		response.Contacts.Email = &contacts.Email.String
	}

	if contacts.TelegramID.Valid {
		response.Contacts.TelegramId = &contacts.TelegramID.Int64
	}

	return response, nil
}

func (s *Service) UpdateUserEmail(ctx context.Context, req *proto.UpdateUserEmailRequest) (*proto.UpdateUserEmailResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserEmail")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserEmail"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserEmail"),
	)

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserEmailParams{
		UserUuid: userUUID,
		Email: pgtype.Text(sql.NullString{
			String: req.Email,
			Valid:  true,
		}),
	}

	contacts, err := s.Repo.UpdateUserEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update email", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update email: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserEmailResponse{
		Contacts: &proto.UserContacts{
			PhoneNumber: contacts.PhoneNumber,
			UserUuid:    contacts.UserUuid.String(),
		},
	}

	if contacts.Email.Valid {
		response.Contacts.Email = &contacts.Email.String
	}

	if contacts.TelegramID.Valid {
		response.Contacts.TelegramId = &contacts.TelegramID.Int64
	}

	return response, nil
}

func (s *Service) UpdateUserTelegramID(ctx context.Context, req *proto.UpdateUserTelegramIDRequest) (*proto.UpdateUserTelegramIDResponse, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserTelegramID")
	defer span.End()

	log := logger.WithTraceContext(ctx, s.Logger).With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("method", "UpdateUserTelegramID"),
	)

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUuid),
		attribute.String("grpc.method", "UpdateUserTelegramID"),
	)

	if !ValidateTelegramID(int(req.TelegramId)) {
		span.SetStatus(codes.Error, "invalid telegram ID")
		return nil, status.Errorf(grpccodes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		span.SetStatus(codes.Error, "invalid UUID format")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserTelegramIDParams{
		UserUuid: userUUID,
		TelegramID: pgtype.Int8(sql.NullInt64{
			Int64: req.TelegramId,
			Valid: true,
		}),
	}

	contacts, err := s.Repo.UpdateUserTelegramID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			span.SetStatus(codes.Error, "not found")
			return nil, status.Errorf(grpccodes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update telegram ID", zap.Error(err))
		span.SetStatus(codes.Error, "database error")
		span.RecordError(err)
		return nil, status.Errorf(grpccodes.Internal, "failed to update telegram ID: %v", err)
	}

	span.SetStatus(codes.Ok, "")
	response := &proto.UpdateUserTelegramIDResponse{
		Contacts: &proto.UserContacts{
			PhoneNumber: contacts.PhoneNumber,
			UserUuid:    contacts.UserUuid.String(),
		},
	}

	if contacts.Email.Valid {
		response.Contacts.Email = &contacts.Email.String
	}

	if contacts.TelegramID.Valid {
		response.Contacts.TelegramId = &contacts.TelegramID.Int64
	}

	return response, nil
}
