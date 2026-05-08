package handler_test

import (
	"context"
	"net/http"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

// mockAuthService implements service.AuthServicer
type mockAuthService struct {
	loginToken string
	loginUser  *domain.User
	loginErr   error
	logoutErr  error
	tokenUser  *domain.User
	tokenErr   error
}

func (m *mockAuthService) Login(_ context.Context, _, _ string, _ *http.Request) (string, *domain.User, error) {
	return m.loginToken, m.loginUser, m.loginErr
}
func (m *mockAuthService) Logout(_ context.Context, _ string) error { return m.logoutErr }
func (m *mockAuthService) GetUserByToken(_ context.Context, _ string) (*domain.User, error) {
	return m.tokenUser, m.tokenErr
}

// mockUserService implements service.UserServicer
type mockUserService struct {
	createUser  *domain.User
	createErr   error
	updateUser  *domain.User
	updateErr   error
	statusErr   error
	getUser     *domain.User
	getErr      error
	listUsers   []*domain.User
	listTotal   int64
	listErr     error
	stats       *domain.UserStats
	statsErr    error
}

func (m *mockUserService) Create(_ context.Context, _ service.CreateUserInput) (*domain.User, error) {
	return m.createUser, m.createErr
}
func (m *mockUserService) Update(_ context.Context, _ int64, _ service.UpdateUserInput) (*domain.User, error) {
	return m.updateUser, m.updateErr
}
func (m *mockUserService) UpdateStatus(_ context.Context, _ int64, _ string) error {
	return m.statusErr
}
func (m *mockUserService) GetByID(_ context.Context, _ int64) (*domain.User, error) {
	return m.getUser, m.getErr
}
func (m *mockUserService) List(_ context.Context, _ domain.UserListParams) ([]*domain.User, int64, error) {
	return m.listUsers, m.listTotal, m.listErr
}
func (m *mockUserService) GetStats(_ context.Context) (*domain.UserStats, error) {
	return m.stats, m.statsErr
}

// mockInvitationService implements service.InvitationServicer
type mockInvitationService struct {
	sendErr    error
	validateToken *domain.InvitationToken
	validateUser  *domain.User
	validateErr   error
	acceptErr     error
}

func (m *mockInvitationService) SendInvitation(_ context.Context, _ int64) error {
	return m.sendErr
}
func (m *mockInvitationService) ValidateToken(_ context.Context, _ string) (*domain.InvitationToken, *domain.User, error) {
	return m.validateToken, m.validateUser, m.validateErr
}
func (m *mockInvitationService) AcceptInvitation(_ context.Context, _, _ string) error {
	return m.acceptErr
}

// mockPropertyService implements service.PropertyServicer
type mockPropertyService struct {
	createProp  *domain.Property
	createErr   error
	updateProp  *domain.Property
	updateErr   error
	deleteErr   error
	getProp     *domain.Property
	getErr      error
	listProps   []*domain.Property
	listTotal   int64
	listErr     error
	stats       *domain.PropertyStats
	statsErr    error
}

func (m *mockPropertyService) Create(_ context.Context, _ service.CreatePropertyInput) (*domain.Property, error) {
	return m.createProp, m.createErr
}
func (m *mockPropertyService) Update(_ context.Context, _ int64, _ service.UpdatePropertyInput) (*domain.Property, error) {
	return m.updateProp, m.updateErr
}
func (m *mockPropertyService) Delete(_ context.Context, _ int64) error {
	return m.deleteErr
}
func (m *mockPropertyService) GetByID(_ context.Context, _ int64) (*domain.Property, error) {
	return m.getProp, m.getErr
}
func (m *mockPropertyService) List(_ context.Context, _ domain.PropertyListParams) ([]*domain.Property, int64, error) {
	return m.listProps, m.listTotal, m.listErr
}
func (m *mockPropertyService) GetStats(_ context.Context, _ int64) (*domain.PropertyStats, error) {
	return m.stats, m.statsErr
}

func makeAdminUser() *domain.User {
	return &domain.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin",
		Role:  domain.RoleAdmin,
		Status: domain.StatusActive,
	}
}

func makeInvitationToken(userID int64) *domain.InvitationToken {
	exp := time.Now().Add(72 * time.Hour)
	return &domain.InvitationToken{
		ID:        1,
		UserID:    userID,
		TokenHash: "hash",
		ExpiresAt: exp,
	}
}
