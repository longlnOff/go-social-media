package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	QueryTimeoutDuration = 3 * time.Second
	ErrConflict = errors.New("resource already exists")
)

type Storage struct {
	Posts interface {
		GetByID(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
	}

	Users interface {
		Create(context.Context, *User) error
		GetByID(context.Context, int64) (*User, error)
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
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}
