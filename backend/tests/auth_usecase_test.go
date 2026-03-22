package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type mockAuthRepo struct {
	byEmail map[string]*modules.User
	byID    map[string]*modules.User
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		byEmail: make(map[string]*modules.User),
		byID:    make(map[string]*modules.User),
	}
}

func (m *mockAuthRepo) CreateUser(_ context.Context, u *modules.User) error {
	if _, exists := m.byEmail[u.Email]; exists {
		return errors.New("duplicate email")
	}
	m.byEmail[u.Email] = u
	m.byID[u.ID] = u
	return nil
}

func (m *mockAuthRepo) GetUserByEmail(_ context.Context, email string) (*modules.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *mockAuthRepo) GetUserByID(_ context.Context, id string) (*modules.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func newTestAuthUsecase() usecase.AuthUsecase {
	return usecase.NewAuthUsecase(newMockAuthRepo(), "test-secret-32-chars-long-key!!", 15*time.Minute, 168*time.Hour)
}

func TestRegister_Success(t *testing.T) {
	svc := newTestAuthUsecase()
	user, err := svc.Register(modules.RegisterRequest{
		Email: "test@example.com", Password: "password123", Role: modules.RoleClient,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
	if user.PasswordHash == "password123" {
		t.Error("password should be hashed, not stored as plaintext")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestAuthUsecase()
	req := modules.RegisterRequest{Email: "dup@example.com", Password: "pass123", Role: modules.RoleClient}
	if _, err := svc.Register(req); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err := svc.Register(req)
	if !errors.Is(err, usecase.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	svc := newTestAuthUsecase()
	_, err := svc.Register(modules.RegisterRequest{
		Email: "x@example.com", Password: "pass123", Role: "superuser",
	})
	if !errors.Is(err, usecase.ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestAuthUsecase()
	_, _ = svc.Register(modules.RegisterRequest{
		Email: "login@example.com", Password: "secret123", Role: modules.RoleClient,
	})
	tokens, err := svc.Login(modules.LoginRequest{Email: "login@example.com", Password: "secret123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestAuthUsecase()
	_, _ = svc.Register(modules.RegisterRequest{
		Email: "pw@example.com", Password: "correct123", Role: modules.RoleClient,
	})
	_, err := svc.Login(modules.LoginRequest{Email: "pw@example.com", Password: "wrong"})
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	svc := newTestAuthUsecase()
	_, _ = svc.Register(modules.RegisterRequest{
		Email: "refresh@example.com", Password: "pass123", Role: modules.RoleClient,
	})
	tokens, _ := svc.Login(modules.LoginRequest{Email: "refresh@example.com", Password: "pass123"})
	newTokens, err := svc.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected new access token")
	}
}

func TestValidateAccessToken_Valid(t *testing.T) {
	svc := newTestAuthUsecase()
	_, _ = svc.Register(modules.RegisterRequest{
		Email: "val@example.com", Password: "pass123", Role: modules.RoleAdmin,
	})
	tokens, _ := svc.Login(modules.LoginRequest{Email: "val@example.com", Password: "pass123"})

	claims, err := svc.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if claims.Email != "val@example.com" {
		t.Errorf("unexpected email: %s", claims.Email)
	}
	if claims.Role != modules.RoleAdmin {
		t.Errorf("unexpected role: %s", claims.Role)
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	svc := newTestAuthUsecase()
	_, err := svc.ValidateAccessToken("not.a.valid.token")
	if !errors.Is(err, usecase.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
