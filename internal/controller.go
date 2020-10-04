package internal

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang-http-service/api"
	"net/http"
)

type controller struct {
}

func NewController() api.ServerInterface {
	return &controller{}
}

func (c *controller) CreateUser(ctx echo.Context) error {
	u := api.User{}
	if err := ctx.Bind(&u); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	zap.S().Infow("received user request", "user", u)
	return ctx.NoContent(http.StatusCreated)
}

func (c *controller) GetUsers(ctx echo.Context, userName string) error {
	users := []api.User{
		{
			Id:   0,
			Name: userName,
		},
		{
			Id:   1,
			Name: userName,
		},
	}
	return ctx.JSON(http.StatusOK, users)
}
