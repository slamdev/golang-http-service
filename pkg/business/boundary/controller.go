package boundary

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang-http-service/api"
	"golang-http-service/pkg/business/control"
	"golang-http-service/pkg/business/entity"
	"net/http"
)

type controller struct {
	userRepo control.UserRepo
}

func NewController(userRepo control.UserRepo) api.StrictServerInterface {
	return &controller{
		userRepo: userRepo,
	}
}

func (c *controller) CreateUser(ctx context.Context, request api.CreateUserRequestObject) (api.CreateUserResponseObject, error) {
	u := entity.User{
		Name: request.Body.Name,
	}
	if err := c.userRepo.CreateUser(ctx, u); err != nil {
		if _, ok := err.(*control.ConstraintViolationError); ok {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return nil, fmt.Errorf("failed to create users; %w", err)
	}
	return api.CreateUser201Response{}, nil
}

func (c *controller) GetUser(ctx context.Context, request api.GetUserRequestObject) (api.GetUserResponseObject, error) {
	if u, err := c.userRepo.FindUser(ctx, request.Userid); err != nil {
		if _, ok := err.(*control.ConstraintViolationError); ok {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return nil, fmt.Errorf("failed to create users; %w", err)
	} else {
		return api.GetUser200JSONResponse{
			Id:   u.Id,
			Name: u.Name,
		}, nil
	}
}

func (c *controller) GetUsers(ctx context.Context, _ api.GetUsersRequestObject) (api.GetUsersResponseObject, error) {
	users := c.userRepo.FindAllUsers(ctx)
	res := make(api.GetUsers200JSONResponse, len(users))
	for i, u := range users {
		res[i] = api.UserV1{
			Id:   u.Id,
			Name: u.Name,
		}
	}
	return res, nil
}
