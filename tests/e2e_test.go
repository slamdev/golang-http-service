package tests

import (
	"context"
	"errors"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"golang-http-service/api"
	"testing"
)

func TestE2E_Should_Verify_User_Flow(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("http://localhost:8080/api")
	assert.NoError(t, err)

	// We add one user
	userToCreate := api.UserV1{
		Name: faker.Name(),
	}
	createUserRes, err := client.CreateUserWithResponse(ctx, userToCreate)
	assert.NoError(t, err)
	assert.Equal(t, 201, createUserRes.StatusCode())

	// We check the user is in the all users list
	getUsersRes, err := client.GetUsersWithResponse(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 200, getUsersRes.StatusCode())
	assert.Greater(t, len(*getUsersRes.JSON200), 0)
	user, err := findUser(userToCreate.Name, *getUsersRes.JSON200)
	assert.NoError(t, err)

	// We check the user can be fetched
	getUserRes, err := client.GetUserWithResponse(ctx, user.Id)
	assert.NoError(t, err)
	assert.Equal(t, 200, getUserRes.StatusCode())
	assert.Equal(t, user.Id, getUserRes.JSON200.Id)
	assert.Equal(t, user.Name, getUserRes.JSON200.Name)

	// We get 404 for user that doesn't exist
	rndInts, err := faker.RandomInt(0, 9999)
	assert.NoError(t, err)
	getUserRes, err = client.GetUserWithResponse(ctx, int32(rndInts[0]))
	assert.NoError(t, err)
	assert.Equal(t, 404, getUserRes.StatusCode())

	// We get 400 when user creation input is invalid
	userToCreate = api.UserV1{
		Name: "",
	}
	createUserRes, err = client.CreateUserWithResponse(ctx, userToCreate)
	assert.NoError(t, err)
	assert.Equal(t, 400, createUserRes.StatusCode())
	assert.NotNil(t, createUserRes.JSON400)
}

func findUser(name string, users []api.UserV1) (api.UserV1, error) {
	for i := range users {
		if users[i].Name == name {
			return users[i], nil
		}
	}
	return api.UserV1{}, errors.New("not found")
}
