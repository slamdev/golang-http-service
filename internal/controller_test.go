package internal

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestController_CreateUser_201(t *testing.T) {
	// prepare request
	userJSON := `{"name":"Jon Snow","id":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// prepare echo context
	e := echo.New()
	ctx := e.NewContext(req, rec)
	// prepare controller
	c := NewController()

	err := c.CreateUser(ctx)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Empty(t, rec.Body.String())
	}
}

func TestController_CreateUser_400(t *testing.T) {
	// prepare request
	userJSON := `{"name":"Jon Snow","id":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// prepare echo context
	e := echo.New()
	ctx := e.NewContext(req, rec)
	// prepare controller
	c := NewController()

	err := c.CreateUser(ctx)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.NotEmpty(t, rec.Body.String())
	}
}

func TestController_GetUsers_200(t *testing.T) {
	// prepare request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	// prepare echo context
	e := echo.New()
	ctx := e.NewContext(req, rec)
	// prepare controller
	c := NewController()

	err := c.GetUsers(ctx, "aaa")

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Body.String())
	}
}
