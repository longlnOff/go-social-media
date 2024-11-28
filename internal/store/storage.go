package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound    = errors.New("record not found")
	QueryTimeoutDuration = 3 * time.Second
	ErrConflict          = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
)

type Storage struct {
	Posts interface {
		GetByID(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}

	Users interface {
		Create(context.Context, *User, *sql.Tx) error
		GetByID(context.Context, int64) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, invitationExpiration time.Duration) error
		Activate(context.Context, string) error
	}

	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}

	Followers interface {
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
