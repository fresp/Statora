package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/security/pii"
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
	emailKey     []byte
}

type CreateUserRequest struct {
	Username  string
	Email     string
	Role      string
	Status    string
	InvitedBy *primitive.ObjectID
	Password  string
}

type CreateInvitationRequest struct {
	Email     string
	Role      string
	CreatedBy primitive.ObjectID
	TokenHash string
	ExpiresAt time.Time
}

func NewService(repo repository.UserRepository, jwtSecret, emailEncryptionKey string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, emailKey: []byte(emailEncryptionKey)}
}

func NewServiceWithSettings(repo repository.UserRepository, settingsRepo repository.SettingsRepository, jwtSecret, emailEncryptionKey string) *Service {
	return &Service{repo: repo, settingsRepo: settingsRepo, jwtSecret: jwtSecret, emailKey: []byte(emailEncryptionKey)}
}

func NewServiceFromDB(db *mongo.Database, jwtSecret, emailEncryptionKey string) *Service {
	return NewServiceWithSettings(repository.NewMongoUserRepository(db), repository.NewMongoSettingsRepository(db), jwtSecret, emailEncryptionKey)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	user, err := s.FindUserByEmail(ctx, req.Email)
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

	email := NormalizeEmail(claims.Email)
	if email == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.FindUserByEmail(ctx, email)
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

func (s *Service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.decryptUserEmail(user)
}

func (s *Service) EmailExists(ctx context.Context, email string) (bool, error) {
	normalized := pii.Normalize(email)
	if normalized == "" {
		return false, nil
	}

	return s.repo.EmailExistsByHash(ctx, pii.Hash(normalized))
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	processedEmail, err := pii.Process(req.Email, s.emailKey)
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     req.Username,
		Email:        processedEmail.Encrypted,
		EmailHash:    processedEmail.Hash,
		Role:         req.Role,
		Status:       req.Status,
		InvitedBy:    req.InvitedBy,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Email = processedEmail.Normalized
	return &user, nil
}

func (s *Service) EncryptForStorage(email string) (encrypted string, emailHash string, normalized string, err error) {
	processed, err := pii.Process(email, s.emailKey)
	if err != nil {
		return "", "", "", err
	}

	if processed.Normalized == "" {
		return "", "", "", nil
	}

	return processed.Encrypted, processed.Hash, processed.Normalized, nil
}

func (s *Service) DecryptStoredEmail(ciphertext string) (string, error) {
	return pii.Decrypt(ciphertext, s.emailKey)
}

func (s *Service) CreateInvitation(ctx context.Context, req CreateInvitationRequest) (*models.UserInvitation, error) {
	encryptedEmail, emailHash, _, err := s.EncryptForStorage(req.Email)
	if err != nil {
		return nil, err
	}

	invitation := &models.UserInvitation{
		ID:        primitive.NewObjectID(),
		TokenHash: req.TokenHash,
		Email:     encryptedEmail,
		EmailHash: emailHash,
		Role:      req.Role,
		ExpiresAt: req.ExpiresAt,
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
	}

	_ = ctx
	return invitation, nil
}

func (s *Service) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normalized := pii.Normalize(email)
	if normalized == "" {
		return nil, mongo.ErrNoDocuments
	}

	user, err := s.repo.FindByEmailHash(ctx, pii.Hash(normalized))
	if err != nil {
		return nil, err
	}

	return s.decryptUserEmail(user)
}

func (s *Service) decryptUserEmail(user *models.User) (*models.User, error) {
	decryptedEmail, err := pii.Decrypt(user.Email, s.emailKey)
	if err != nil {
		return nil, err
	}

	user.Email = decryptedEmail
	return user, nil
}
