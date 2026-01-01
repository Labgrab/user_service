package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
)

type Service struct {
	proto.UnimplementedUserServiceServer
	repo *sqlc.Queries
}

func (s *Service) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	createdUUID, err := s.repo.CreateUser(ctx, userUUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.CreateUserResponse{
		User: &proto.User{
			Uuid: createdUUID.String(),
		},
	}, nil
}

// GetUserDetails retrieves user details by UUID
func (s *Service) GetUserDetails(ctx context.Context, req *proto.GetUserDetailsRequest) (*proto.GetUserDetailsResponse, error) {
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.repo.GetUserDetails(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
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

// GetUserContacts retrieves user contacts by UUID
func (s *Service) GetUserContacts(ctx context.Context, req *proto.GetUserContactsRequest) (*proto.GetUserContactsResponse, error) {
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.repo.GetUserContacts(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
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

// DeleteUser deletes a user by UUID
func (s *Service) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	userUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	if err := s.repo.DeleteUser(ctx, userUUID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &proto.DeleteUserResponse{}, nil
}

// CreateUserDetails creates user details with validation
func (s *Service) CreateUserDetails(ctx context.Context, req *proto.CreateUserDetailsRequest) (*proto.CreateUserDetailsResponse, error) {
	// Validate input
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	// Prepare parameters
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

	details, err := s.repo.CreateUserDetails(ctx, params)
	if err != nil {
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

// UpdateUserName updates user's name
func (s *Service) UpdateUserName(ctx context.Context, req *proto.UpdateUserNameRequest) (*proto.UpdateUserNameResponse, error) {
	if !ValidateAlphabeticString(req.Name) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid name format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.repo.UpdateUserName(ctx, sqlc.UpdateUserNameParams{
		UserUuid: userUUID,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
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

// UpdateUserSurname updates user's surname
func (s *Service) UpdateUserSurname(ctx context.Context, req *proto.UpdateUserSurnameRequest) (*proto.UpdateUserSurnameResponse, error) {
	if !ValidateAlphabeticString(req.Surname) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid surname format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.repo.UpdateUserSurname(ctx, sqlc.UpdateUserSurnameParams{
		UserUuid: userUUID,
		Surname:  req.Surname,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
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

// UpdateUserPatronymic updates user's patronymic
func (s *Service) UpdateUserPatronymic(ctx context.Context, req *proto.UpdateUserPatronymicRequest) (*proto.UpdateUserPatronymicResponse, error) {
	if req.Patronymic != nil && !ValidateAlphabeticString(*req.Patronymic) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserPatronymicParams{
		UserUuid: userUUID,
	}

	if req.Patronymic != nil {
		params.Patronymic = pgtype.Text(sql.NullString{
			String: *req.Patronymic,
			Valid:  true,
		})
	}

	details, err := s.repo.UpdateUserPatronymic(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
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

// UpdateUserGroupCode updates user's group code
func (s *Service) UpdateUserGroupCode(ctx context.Context, req *proto.UpdateUserGroupCodeRequest) (*proto.UpdateUserGroupCodeResponse, error) {
	if !ValidateGroupCode(req.GroupCode) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid group code format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	details, err := s.repo.UpdateUserGroupCode(ctx, sqlc.UpdateUserGroupCodeParams{
		UserUuid:  userUUID,
		GroupCode: req.GroupCode,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user details not found")
		}
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

// CreateUserContacts creates user contacts with validation
func (s *Service) CreateUserContacts(ctx context.Context, req *proto.CreateUserContactsRequest) (*proto.CreateUserContactsResponse, error) {
	if !ValidatePhoneNumber(req.PhoneNumber) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid phone number format")
	}

	if req.TelegramId != nil && *req.TelegramId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
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

	contacts, err := s.repo.CreateUserContacts(ctx, params)
	if err != nil {
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

// UpdateUserPhoneNumber updates user's phone number
func (s *Service) UpdateUserPhoneNumber(ctx context.Context, req *proto.UpdateUserPhoneNumberRequest) (*proto.UpdateUserPhoneNumberResponse, error) {
	if !ValidatePhoneNumber(req.PhoneNumber) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid phone number format")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	contacts, err := s.repo.UpdateUserPhoneNumber(ctx, sqlc.UpdateUserPhoneNumberParams{
		UserUuid:    userUUID,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
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

// UpdateUserEmail updates user's email
func (s *Service) UpdateUserEmail(ctx context.Context, req *proto.UpdateUserEmailRequest) (*proto.UpdateUserEmailResponse, error) {
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserEmailParams{
		UserUuid: userUUID,
	}

	if req.Email != nil {
		params.Email = pgtype.Text(sql.NullString{
			String: *req.Email,
			Valid:  true,
		})
	}

	contacts, err := s.repo.UpdateUserEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
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

// UpdateUserTelegramID updates user's Telegram ID
func (s *Service) UpdateUserTelegramID(ctx context.Context, req *proto.UpdateUserTelegramIDRequest) (*proto.UpdateUserTelegramIDResponse, error) {
	if req.TelegramId != nil && *req.TelegramId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "telegram ID must be positive")
	}

	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	params := sqlc.UpdateUserTelegramIDParams{
		UserUuid: userUUID,
	}

	if req.TelegramId != nil {
		params.TelegramID = pgtype.Int8(sql.NullInt64{
			Int64: *req.TelegramId,
			Valid: true,
		})
	}

	contacts, err := s.repo.UpdateUserTelegramID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user contacts not found")
		}
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
