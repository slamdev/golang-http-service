package e2e

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"golang-http-service/api"
)

func TestE2E_Should_Verify_User_Flow(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// We add one user
	userToCreate := api.UserV1{
		Name: faker.Name(),
	}
	createUserRes, err := client.CreateUserWithResponse(ctx, userToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createUserRes.StatusCode())

	// We check the user is in the all users list
	getUsersRes, err := client.GetUsersWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getUsersRes.StatusCode())
	require.Greater(t, len(*getUsersRes.JSON200), 0)
	user, err := findUser(userToCreate.Name, *getUsersRes.JSON200)
	require.NoError(t, err)

	// We check the user can be fetched
	getUserRes, err := client.GetUserWithResponse(ctx, user.Id)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getUserRes.StatusCode())
	require.Equal(t, user.Id, getUserRes.JSON200.Id)
	require.Equal(t, user.Name, getUserRes.JSON200.Name)

	// We get 404 for user that doesn't exist
	rndInts, err := faker.RandomInt(0, 9999)
	require.NoError(t, err)
	getUserRes, err = client.GetUserWithResponse(ctx, int32(rndInts[0]))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, getUserRes.StatusCode())

	// We get 400 when user creation input is invalid
	userToCreate = api.UserV1{
		Name: "",
	}
	createUserRes, err = client.CreateUserWithResponse(ctx, userToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, createUserRes.StatusCode())
	require.NotNil(t, createUserRes.ApplicationproblemJSON400)
}

func findUser(name string, users []api.UserV1) (api.UserV1, error) {
	for i := range users {
		if users[i].Name == name {
			return users[i], nil
		}
	}
	return api.UserV1{}, errors.New("not found")
}
