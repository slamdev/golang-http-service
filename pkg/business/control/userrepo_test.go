package control

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang-http-service/pkg/business/entity"
)

func TestUserRepo_Should_Create_User(t *testing.T) {
	r := NewUserRepo()
	ctx := context.Background()
	name := "some"

	err := r.CreateUser(ctx, entity.User{Name: name})
	assert.NoError(t, err)

	us := r.FindAllUsers(ctx)
	assert.Len(t, us, 1)
	assert.Equal(t, us[0].Name, name)
	assert.NotZero(t, us[0].Id)
}

func TestUserRepo_Should_Not_Create_User_With_Same_Name(t *testing.T) {
	r := NewUserRepo()
	ctx := context.Background()
	name := "some"
	_ = r.CreateUser(ctx, entity.User{Name: name})

	err := r.CreateUser(ctx, entity.User{Name: name})
	var expected *ValidationError
	assert.ErrorAs(t, err, &expected)
}

func TestUserRepo_Should_Find_User(t *testing.T) {
	r := NewUserRepo()
	ctx := context.Background()
	name := "some"
	_ = r.CreateUser(ctx, entity.User{Name: name})

	us := r.FindAllUsers(ctx)
	assert.Len(t, us, 1)

	u, err := r.FindUser(ctx, us[0].Id)
	assert.NoError(t, err)
	assert.Equal(t, u.Name, name)
	assert.NotZero(t, u.Id)
}

func TestUserRepo_Should_Fail_To_Find_Absent_User(t *testing.T) {
	r := NewUserRepo()
	ctx := context.Background()

	_, err := r.FindUser(ctx, 1)
	var expected *MissingEntityError
	assert.ErrorAs(t, err, &expected)
}
