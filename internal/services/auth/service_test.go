package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/fresp/Statora/internal/middleware"
	"github.com/fresp/Statora/internal/models"
)

type stubUserRepo struct {
	user           *models.User
	err            error
	created        *models.User
	lastEmailHash  string
	lastExistsHash string
}

func (r *stubUserRepo) FindByEmailHash(_ context.Context, emailHash string) (*models.User, error) {
	r.lastEmailHash = emailHash
	if r.err != nil {
		return nil, r.err
	}
	return r.user, nil
}

func (r *stubUserRepo) EmailExistsByHash(_ context.Context, emailHash string) (bool, error) {
	r.lastExistsHash = emailHash
	if r.err != nil {
		return false, r.err
	}
	return r.user != nil, nil
}

func (r *stubUserRepo) Create(_ context.Context, user models.User) error {
	r.created = &user
	return r.err
}

func (r *stubUserRepo) FindByID(_ context.Context, _ string) (*models.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.user, nil
}

func (r *stubUserRepo) UpdateProfile(_ context.Context, _ string, _ string, _ *string) error {
	return r.err
}

func (r *stubUserRepo) BeginMFAEnrollment(_ context.Context, _ string, _ string, _ []string) error {
	return r.err
}

func (r *stubUserRepo) EnableMFA(_ context.Context, _ string) error {
	return r.err
}

func (r *stubUserRepo) DisableMFA(_ context.Context, _ string) error {
	return r.err
}

func (r *stubUserRepo) ReplaceRecoveryCodes(_ context.Context, _ string, _ []string) error {
	return r.err
}

func TestLoginIncludesRoleAndMFAVerifiedClaims(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	repo := &stubUserRepo{
		user: &models.User{
			ID:           primitive.NewObjectID(),
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: string(hash),
			Role:         "admin",
			MFAEnabled:   false,
		},
	}

	enc, err := EncryptEmail("admin@example.com", []byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	repo.user.Email = enc

	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")

	result, err := svc.Login(context.Background(), LoginRequest{Email: "admin@example.com", Password: "secret123"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "admin", result.User.Role)
	assert.Equal(t, false, result.MFARequired)

	parsed := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(result.Token, parsed, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsed.MFAVerified)
}

func TestLoginDefaultsRoleToAdminWhenMissing(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	repo := &stubUserRepo{
		user: &models.User{
			ID:           primitive.NewObjectID(),
			Username:     "legacy",
			Email:        "legacy@example.com",
			PasswordHash: string(hash),
			Role:         "",
			MFAEnabled:   false,
		},
	}

	enc, err := EncryptEmail("legacy@example.com", []byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	repo.user.Email = enc

	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")

	result, err := svc.Login(context.Background(), LoginRequest{Email: "legacy@example.com", Password: "secret123"})
	assert.NoError(t, err)
	assert.Equal(t, "admin", result.User.Role)
	assert.False(t, result.MFARequired)
}

func TestLoginReturnsMFARequiredForUnenrolledUsers(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	repo := &stubUserRepo{user: &models.User{
		ID:           primitive.NewObjectID(),
		Username:     "user-one",
		Email:        "user-one@example.com",
		PasswordHash: string(hash),
		Role:         "operator",
		MFAEnabled:   false,
	}}

	enc, err := EncryptEmail("user-one@example.com", []byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	repo.user.Email = enc

	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")
	result, err := svc.Login(context.Background(), LoginRequest{Email: "user-one@example.com", Password: "secret123"})
	assert.NoError(t, err)
	assert.False(t, result.MFARequired)
}

func TestLoginReturnsMFARequiredForEnrolledUsers(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	repo := &stubUserRepo{user: &models.User{
		ID:           primitive.NewObjectID(),
		Username:     "user-two",
		Email:        "user-two@example.com",
		PasswordHash: string(hash),
		Role:         "admin",
		MFAEnabled:   true,
	}}

	enc, err := EncryptEmail("user-two@example.com", []byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	repo.user.Email = enc

	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")
	result, err := svc.Login(context.Background(), LoginRequest{Email: "user-two@example.com", Password: "secret123"})
	assert.NoError(t, err)
	assert.True(t, result.MFARequired)
}

func TestCreateUserEncryptsEmailAndStoresHash(t *testing.T) {
	repo := &stubUserRepo{}
	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")

	created, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Username: "alice",
		Email:    " Alice@Example.com ",
		Role:     "admin",
		Status:   "active",
		Password: "secret123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "alice@example.com", created.Email)
	assert.NotNil(t, repo.created)
	assert.NotEqual(t, "alice@example.com", repo.created.Email)
	assert.Equal(t, HashEmail("alice@example.com"), repo.created.EmailHash)
}

func TestFindUserByEmailNormalizesHashesAndDecrypts(t *testing.T) {
	repo := &stubUserRepo{user: &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
		Role:     "admin",
		Status:   "active",
	}}

	enc, err := EncryptEmail("alice@example.com", []byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	repo.user.Email = enc

	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")

	user, err := svc.FindUserByEmail(context.Background(), "  Alice@Example.com ")
	assert.NoError(t, err)
	assert.Equal(t, HashEmail("alice@example.com"), repo.lastEmailHash)
	assert.Equal(t, "alice@example.com", user.Email)
}

func TestFindUserByEmailReturnsNoDocumentsForBlankInput(t *testing.T) {
	repo := &stubUserRepo{}
	svc := NewService(repo, "test-secret", "12345678901234567890123456789012")

	user, err := svc.FindUserByEmail(context.Background(), "   ")
	assert.Nil(t, user)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	assert.Empty(t, repo.lastEmailHash)
}
