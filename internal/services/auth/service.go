package auth

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSSONotConfigured   = errors.New("sso_not_configured")
	ErrSSODisabled        = errors.New("sso_disabled")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrUserNotFound       = errors.New("user_not_found")
	ErrSSONotAllowed      = errors.New("sso_not_allowed")
)

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token string
	User  struct {
		ID       string
		Username string
		Email    string
		Role     string
	}
	MFARequired bool
}

type ssoVerifier interface {
	Verify(rawToken string) (*externalSSOClaims, error)
}

type Service struct {
	repo         repository.UserRepository
	settingsRepo repository.SettingsRepository
	jwtSecret    string
}

func NewService(repo repository.UserRepository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func NewServiceWithSettings(repo repository.UserRepository, settingsRepo repository.SettingsRepository, jwtSecret string) *Service {
	return &Service{repo: repo, settingsRepo: settingsRepo, jwtSecret: jwtSecret}
}

func NewServiceFromDB(db *mongo.Database, jwtSecret string) *Service {
	return NewServiceWithSettings(repository.NewMongoUserRepository(db), repository.NewMongoSettingsRepository(db), jwtSecret)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildLoginResult(user, !user.MFAEnabled)
}

func (s *Service) AuthenticateSSO(ctx context.Context, rawToken string) (*LoginResult, error) {
	if s.settingsRepo == nil {
		return nil, ErrSSONotConfigured
	}

	ssoSettings, err := s.settingsRepo.GetSSOSettings(ctx)
	if err != nil {
		return nil, err
	}

	if ssoSettings == nil {
		return nil, ErrSSONotConfigured
	}

	if strings.TrimSpace(ssoSettings.Issuer) == "" || strings.TrimSpace(ssoSettings.Audience) == "" || strings.TrimSpace(ssoSettings.Algorithm) == "" {
		return nil, ErrSSONotConfigured
	}

	if !ssoSettings.Enabled {
		return nil, ErrSSODisabled
	}

	verifier, err := newSSOVerifier(*ssoSettings)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, err := verifier.Verify(rawToken)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.SSO.Enabled {
		return nil, ErrSSONotAllowed
	}

	return s.buildLoginResult(user, true)
}

func (s *Service) buildLoginResult(user *models.User, mfaVerified bool) (*LoginResult, error) {
	role := user.Role
	if role == "" {
		role = "admin"
	}

	mfaRequired := !mfaVerified
	token, err := generateAccessToken(user.ID.Hex(), user.Username, role, mfaVerified, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	var result LoginResult
	result.Token = token
	result.User.ID = user.ID.Hex()
	result.User.Username = user.Username
	result.User.Email = user.Email
	result.User.Role = role
	result.MFARequired = mfaRequired

	return &result, nil
}
