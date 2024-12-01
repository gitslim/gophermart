package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceImpl реализует интерфейс service.UserService
type UserServiceImpl struct {
	userStorage storage.UserStorage
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(userStorage storage.UserStorage) service.UserService {
	return &UserServiceImpl{
		userStorage: userStorage,
	}
}

// Register регистрирует нового пользователя
func (s *UserServiceImpl) Register(ctx context.Context, login, password string) (*models.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.userStorage.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user := &models.User{
		Login:        login,
		PasswordHash: string(hashedPassword),
		Balance:      0,
		CreatedAt:    time.Now(),
	}

	if err := s.userStorage.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login аутентифицирует пользователя
func (s *UserServiceImpl) Login(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.userStorage.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *UserServiceImpl) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userStorage.GetUserByID(ctx, id)
}
