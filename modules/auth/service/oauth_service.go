package service

import (
	"context"
	"fmt"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	authDto "github.com/Caknoooo/go-gin-clean-starter/modules/auth/dto"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	userDto "github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	userRepo "github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==================== OAUTH SERVICE ====================
type OAuthService interface {
	HandleGoogleRegister(ctx context.Context, googleUser userDto.GoogleUserData) (userDto.UserResponse, error)
	GoogleLogin(ctx context.Context, googleUser userDto.GoogleUserData) (authDto.TokenResponse, error)
}

type oauthService struct {
	userRepository         userRepo.UserRepository
	refreshTokenRepository authRepo.RefreshTokenRepository
	jwtService             JWTService
	db                     *gorm.DB
}

func NewOAuthService(
	userRepository userRepo.UserRepository,
	refreshTokenRepository authRepo.RefreshTokenRepository,
	jwtService JWTService,
	db *gorm.DB,
) OAuthService {
	return &oauthService{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		jwtService:             jwtService,
		db:                     db,
	}
}

// HandleGoogleRegister: CEK email, kalau belum ada BARU register
func (s *oauthService) HandleGoogleRegister(ctx context.Context, googleUser userDto.GoogleUserData) (userDto.UserResponse, error) {
	// 1. Cek apakah email sudah terdaftar
	user, exists, err := s.userRepository.CheckEmail(ctx, nil, googleUser.Email)
	if err != nil {
		return userDto.UserResponse{}, fmt.Errorf("failed to check email: %w", err)
	}

	// 2. Jika belum ada, buat user baru
	if !exists {
		newUser := entities.User{
			ID:         uuid.New(),
			Name:       googleUser.Name,
			Email:      googleUser.Email,
			Password:   "", // OAuth tidak pakai password
			Avatar:     googleUser.Avatar,
			Role:       "user",
			IsVerified: true,
		}

		user, err = s.userRepository.Register(ctx, nil, newUser)
		if err != nil {
			return userDto.UserResponse{}, fmt.Errorf("failed to register user: %w", err)
		}
	}

	// 3. Return user response
	return userDto.UserResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Institution: user.Institution,
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsVerified:  user.IsVerified,
	}, nil
}

// GoogleLogin: Handle login via Google OAuth
func (s *oauthService) GoogleLogin(ctx context.Context, googleUser userDto.GoogleUserData) (authDto.TokenResponse, error) {
	// 1. Cek email & register otomatis jika belum ada
	userResp, err := s.HandleGoogleRegister(ctx, googleUser)
	if err != nil {
		return authDto.TokenResponse{}, fmt.Errorf("failed to handle google user: %w", err)
	}

	// 2. Ambil user dari database (untuk pastikan data terbaru & lengkap)
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, userResp.Email)
	if err != nil {
		return authDto.TokenResponse{}, fmt.Errorf("user not found after google register: %w", err)
	}

	// 3. Generate access & refresh token
	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	refreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, refreshToken)
	if err != nil {
		return authDto.TokenResponse{}, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return authDto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}
