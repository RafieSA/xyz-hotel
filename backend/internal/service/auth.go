package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
)

// Claims embedded in JWT.
type Claims struct {
	UserID    int64  `json:"uid"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	TokenType string `json:"typ"` // "access" or "refresh"
	jwt.RegisteredClaims
}

var validate = validator.New()

// AuthService handles register/login/refresh with bcrypt + JWT.
type AuthService struct {
	UserRepo  *repo.UserRepo
	JWTSecret []byte
}

func NewAuthService(userRepo *repo.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{UserRepo: userRepo, JWTSecret: []byte(jwtSecret)}
}

// HashPassword hashes with bcrypt.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword compares bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *AuthService) generateTokens(u *model.User) (access, refresh string, err error) {
	now := time.Now()
	accessClaims := Claims{
		UserID:    u.ID,
		Role:      u.Role,
		Email:     u.Email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", u.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	refreshClaims := Claims{
		UserID:    u.ID,
		Role:      u.Role,
		Email:     u.Email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", u.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	access, err = at.SignedString(s.JWTSecret)
	if err != nil {
		return "", "", err
	}
	refresh, err = rt.SignedString(s.JWTSecret)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// Register creates a new customer user. Role is forced to customer to prevent privilege escalation (BFLA).
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if name == "" || email == "" || password == "" {
		return nil, errors.New("Name, email and password are required. Please fill in all fields")
	}
	if len(password) < 8 {
		return nil, errors.New("Password needs at least 8 characters")
	}
	if err := validate.Var(email, "required,email"); err != nil {
		return nil, errors.New("Enter a valid email address like name@example.com")
	}
	_, err := s.UserRepo.FindByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("This email is already registered. Please sign in instead")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		if err != nil && !strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("check email: %w", err)
		}
	}
	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Name:     name,
		Email:    email,
		Password: hashed,
		Role:     model.RoleCustomer,
	}
	if err := s.UserRepo.Create(ctx, u); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("This email is already registered. Please sign in instead")
		}
		return nil, err
	}
	slog.Info("user registered", "user_id", u.ID, "email", u.Email)
	return u, nil
}

// Login verifies credentials and issues JWT pair.
func (s *AuthService) Login(ctx context.Context, email, password string) (access, refresh string, user *model.User, err error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			return "", "", nil, errors.New("Email or password is incorrect. Please check and try again")
		}
		return "", "", nil, err
	}
	if !CheckPassword(u.Password, password) {
		slog.Warn("login failed: bad password", "email", email)
		return "", "", nil, errors.New("Email or password is incorrect. Please check and try again")
	}
	access, refresh, err = s.generateTokens(u)
	if err != nil {
		return "", "", nil, err
	}
	slog.Info("user login", "user_id", u.ID, "email", u.Email)
	return access, refresh, u, nil
}

// Refresh validates a refresh token and issues a new pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("Your session expired. Please sign in again")
	}
	if claims.TokenType != "refresh" {
		return "", "", errors.New("This token cannot refresh your session. Please sign in again")
	}
	u, err := s.UserRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", errors.New("We could not find your account. Please sign in again")
	}
	return s.generateTokens(u)
}

// IssueTokensForTest is exported for tests to generate tokens without DB lookup.
func (s *AuthService) IssueTokensForTest(u *model.User) (string, string, error) {
	return s.generateTokens(u)
}
