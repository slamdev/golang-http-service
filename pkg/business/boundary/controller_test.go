package boundary

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang-http-service/api"
	controlmock "golang-http-service/pkg/business/control/mock"
	"golang-http-service/pkg/business/entity"
	"testing"
)

func TestController_Should_Create_User(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := controlmock.NewMockUserRepo(ctrl)
	c := NewController(repo)
	ctx := context.Background()
	name := "some"
	repo.EXPECT().CreateUser(ctx, entity.User{Name: name})

	_, err := c.CreateUser(ctx, api.CreateUserRequestObject{Body: &api.UserV1{Name: name}})

	assert.NoError(t, err)
}
