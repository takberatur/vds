package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type UserRepository interface {
	BaseRepository
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	Create(ctx context.Context, user *model.User, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error
}

type userRepository struct {
	*baseRepository
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT 
			u.id, u.email, u.password_hash, u.full_name, u.avatar_url, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
			r.id, r.name, r.permissions
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`

	var user model.User
	user.Role = &model.Role{}

	err := r.db.QueryRow(subCtx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.AvatarURL,
		&user.RoleID, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&user.Role.ID, &user.Role.Name, &user.Role.Permissions,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT 
			u.id, u.email, u.password_hash, u.full_name, u.avatar_url, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
			r.id, r.name, r.permissions
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

	var user model.User
	user.Role = &model.Role{}

	err := r.db.QueryRow(subCtx, query, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.AvatarURL,
		&user.RoleID, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&user.Role.ID, &user.Role.Name, &user.Role.Permissions,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User, passwordHash string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (email, full_name, password_hash, avatar_url, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, (SELECT id FROM roles WHERE name = 'customer')), $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()

	var pwdHash *string
	if passwordHash != "" {
		pwdHash = &passwordHash
	}

	err := r.db.QueryRow(subCtx, query,
		user.Email,
		user.FullName,
		pwdHash, // Pass pointer (can be nil)
		user.AvatarURL,
		user.RoleID, // Can be nil, handled by COALESCE
		true,        // is_active
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := r.db.Exec(subCtx, query, time.Now(), userID)
	return err
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `UPDATE users SET avatar_url = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(subCtx, query, avatarURL, time.Now(), userID)
	return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(subCtx, query, passwordHash, time.Now(), userID)
	return err
}

func (r *userRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `UPDATE users SET full_name = $1, email = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(subCtx, query, req.FullName, req.Email, time.Now(), userID)
	return err
}
