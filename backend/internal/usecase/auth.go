package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidRole        = errors.New("invalid role")
)

type authUsecase struct {
	repo            repository.AuthRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthUsecase(repo repository.AuthRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) AuthUsecase {
	return &authUsecase{
		repo:            repo,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (u *authUsecase) Register(req modules.RegisterRequest) (*modules.User, error) {
	if !isValidRole(req.Role) {
		return nil, ErrInvalidRole
	}

	ctx := context.Background()
	existing, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) && !isNotFoundErr(err) {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &modules.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		CreatedAt:    time.Now().UTC(),
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

func (u *authUsecase) Login(req modules.LoginRequest) (*modules.TokenPair, error) {
	ctx := context.Background()
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u.generateTokenPair(user)
}

func (u *authUsecase) Refresh(refreshToken string) (*modules.TokenPair, error) {
	claims, err := u.parseToken(refreshToken, "refresh")
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := u.repo.GetUserByID(context.Background(), claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return u.generateTokenPair(user)
}

func (u *authUsecase) ValidateAccessToken(token string) (*modules.Claims, error) {
	return u.parseToken(token, "access")
}

func (u *authUsecase) generateTokenPair(user *modules.User) (*modules.TokenPair, error) {
	accessToken, err := u.createToken(user, "access", u.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := u.createToken(user, "refresh", u.refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &modules.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (u *authUsecase) createToken(user *modules.User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       string(user.Role),
		"token_type": tokenType,
		"iat":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.jwtSecret)
}

func (u *authUsecase) parseToken(tokenStr, expectedType string) (*modules.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if tt, _ := mapClaims["token_type"].(string); tt != expectedType {
		return nil, ErrInvalidToken
	}

	userID, _ := mapClaims["user_id"].(string)
	email, _ := mapClaims["email"].(string)
	role, _ := mapClaims["role"].(string)

	if userID == "" || email == "" || role == "" {
		return nil, ErrInvalidToken
	}

	return &modules.Claims{UserID: userID, Email: email, Role: modules.Role(role)}, nil
}

func isValidRole(r modules.Role) bool {
	switch r {
	case modules.RoleAdmin, modules.RoleDispatcher, modules.RoleCourier, modules.RoleClient:
		return true
	}
	return false
}

func isNotFoundErr(err error) bool {
	return err != nil && err.Error() == "user not found"
}
