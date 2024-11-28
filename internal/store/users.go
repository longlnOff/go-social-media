package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash
	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User, tx *sql.Tx) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.QueryRowContext(ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, username, email, created_at, password FROM users
		WHERE id = $1
	`
	user := User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User,
	token string, invitationExpiration time.Duration) error {
	// transaction wraper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := s.Create(ctx, user, tx); err != nil {
			return err
		}
		// create the user invite
		if err := s.createUserInvitation(ctx, tx, token, invitationExpiration, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string,
	expiration time.Duration, userID int64) error {
	query := `
		INSERT INTO user_invitations (token, expiry, user_id)
		VALUES ($1, $2, $3)
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, token, time.Now().Add(expiration), userID)
	if err != nil {
		return err
	}
	return nil
}


func (s *UserStore) Activate(ctx context.Context, token string) error {
	// we have token
	// search for token in DB
	// if not found, return error
	// if found, update user status to active
	// clean the invitation
	// return 

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// search user by token
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrRecordNotFound
		}

		// update user
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		// clean the invitations
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `
		DELETE FROM user_invitations
		WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users
		SET username = $1,
		email = $2,
		is_active = $3
		WHERE id = $4
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui ON ui.user_id = u.id
	WHERE ui.token = $1 AND ui.expiry > $2
	`
	user := &User{}
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}