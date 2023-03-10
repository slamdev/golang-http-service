package control

import (
	"context"
	"fmt"
	"golang-http-service/pkg/business/entity"
	"math/rand"
)

type ConstraintViolationError struct {
	err        string
	constraint string
}

func (e *ConstraintViolationError) Error() string {
	return fmt.Sprintf("db query failed; %s", e.err)
}

type EmptyResultSetError struct {
	err string
}

func (e *EmptyResultSetError) Error() string {
	return fmt.Sprintf("db query failed; %s", e.err)
}

type userRepo struct {
	db map[int32]entity.User
}

type UserRepo interface {
	CreateUser(ctx context.Context, u entity.User) error
	FindUser(ctx context.Context, id int32) (entity.User, error)
	FindAllUsers(ctx context.Context) []entity.User
}

func NewUserRepo() UserRepo {
	return &userRepo{
		db: make(map[int32]entity.User),
	}
}

func (r *userRepo) CreateUser(_ context.Context, u entity.User) error {
	for i := range r.db {
		if r.db[i].Name == u.Name {
			return &ConstraintViolationError{
				err:        fmt.Sprintf("user with name %s already exists", u.Name),
				constraint: "UNIQUE_KEY",
			}
		}
	}
	id := int32(rand.Intn(999))
	u.Id = id
	r.db[id] = u
	return nil
}

func (r *userRepo) FindUser(_ context.Context, id int32) (entity.User, error) {
	if u, ok := r.db[id]; ok {
		return u, nil
	}
	return entity.User{}, &EmptyResultSetError{
		err: fmt.Sprintf("user with id %d is not found", id),
	}
}

func (r *userRepo) FindAllUsers(_ context.Context) []entity.User {
	users := make([]entity.User, 0, len(r.db))
	for _, u := range r.db {
		users = append(users, u)
	}
	return users
}
