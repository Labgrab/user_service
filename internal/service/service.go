package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
)

type Service struct {
	proto.UnimplementedUserServiceServer
	Logger *zap.Logger
	Repo   *sqlc.Queries
}

func (s *Service) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	log := s.Logger.With(zap.String("user_uuid", req.Uuid), zap.String("method", "CreateUser"))
	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	createdUUID, err := s.Repo.CreateUser(ctx, userUUID)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.CreateUserResponse{
		User: &proto.User{
			Uuid: createdUUID.String(),
		},
	}, nil
}

func (s *Service) GetUserDetails(ctx context.Context, req *proto.GetUserDetailsRequest) (*proto.GetUserDetailsResponse, error) {
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "GetUserDetails"))
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.GetUserDetails(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
		log.Error("Failed to get user details", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user details: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "GetUserContacts"))
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.Repo.GetUserContacts(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
		log.Error("Failed to get user contacts", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user contacts: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.Uuid), zap.String("method", "DeleteUser"))
	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	if err := s.Repo.DeleteUser(ctx, userUUID); err != nil {
		log.Error("Failed to delete user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &proto.DeleteUserResponse{}, nil
}

func (s *Service) CreateUserDetails(ctx context.Context, req *proto.CreateUserDetailsRequest) (*proto.CreateUserDetailsResponse, error) {
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "CreateUserDetails"))
	if !ValidateAlphabeticString(req.Name) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid name format")
	}
	if !ValidateAlphabeticString(req.Surname) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid surname format")
	}
	if req.Patronymic != nil && !ValidateAlphabeticString(*req.Patronymic) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic format")
	}
	if !ValidateGroupCode(req.GroupCode) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid group code format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
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
		return nil, status.Errorf(codes.Internal, "failed to create user details: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserName"))
	if !ValidateAlphabeticString(req.Name) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid name format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserName(ctx, sqlc.UpdateUserNameParams{
		UserUuid: userUUID,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
		log.Error("Failed to update user name", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user name: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserSurname"))
	if !ValidateAlphabeticString(req.Surname) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid surname format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserSurname(ctx, sqlc.UpdateUserSurnameParams{
		UserUuid: userUUID,
		Surname:  req.Surname,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
		log.Error("Failed to update user surname", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user surname: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserPatronymic"))
	if !ValidateAlphabeticString(req.Patronymic) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserPatronymicParams{
		UserUuid: userUUID,
	}

	params.Patronymic = pgtype.Text(sql.NullString{
		String: req.Patronymic,
		Valid:  true,
	})

	details, err := s.Repo.UpdateUserPatronymic(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
		log.Error("Failed to update user patronymic", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user patronymic: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserGroupCode"))
	if !ValidateGroupCode(req.GroupCode) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid group code format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.Repo.UpdateUserGroupCode(ctx, sqlc.UpdateUserGroupCodeParams{
		UserUuid:  userUUID,
		GroupCode: req.GroupCode,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User details not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
		log.Error("Failed to update user group code", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user group code: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "CreateUserContacts"))
	if !ValidatePhoneNumber(req.PhoneNumber) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid phone number format")
	}

	if req.TelegramId != nil && *req.TelegramId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
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
		return nil, status.Errorf(codes.Internal, "failed to create user contacts: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserPhoneNumber"))
	if !ValidatePhoneNumber(req.PhoneNumber) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid phone number format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.Repo.UpdateUserPhoneNumber(ctx, sqlc.UpdateUserPhoneNumberParams{
		UserUuid:    userUUID,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update phone number", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update phone number: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserEmail"))
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserEmailParams{
		UserUuid: userUUID,
	}

	params.Email = pgtype.Text(sql.NullString{
		String: req.Email,
		Valid:  true,
	})

	contacts, err := s.Repo.UpdateUserEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update email", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update email: %v", err)
	}

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
	log := s.Logger.With(zap.String("user_uuid", req.UserUuid), zap.String("method", "UpdateUserTelegramID"))
	if !ValidateTelegramID(int(req.TelegramId)) {
		return nil, status.Errorf(codes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		log.Warn("Failed to parse uuid", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserTelegramIDParams{
		UserUuid: userUUID,
	}

	params.TelegramID = pgtype.Int8(sql.NullInt64{
		Int64: req.TelegramId,
		Valid: true,
	})

	contacts, err := s.Repo.UpdateUserTelegramID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User contacts not found", zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
		log.Error("Failed to update telegram ID", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update telegram ID: %v", err)
	}

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
