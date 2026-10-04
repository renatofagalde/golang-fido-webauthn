package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
	"github.com/stretchr/testify/assert"
)

type fakeUserService struct {
	user *domain.User
	err  error
}

func (f *fakeUserService) Create(ctx context.Context, input application.CreateUserInput) (*domain.User, error) {
	return f.user, f.err
}

func TestUserHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		body           string
		service        *fakeUserService
		expectedStatus int
	}{
		{name: "success", body: `{"username":"test","display_name":"test"}`,
			service: &fakeUserService{
				user: &domain.User{
					Hash:        "abc-123",
					Username:    "test",
					DisplayName: "test",
					IsActive:    true,
				},
			}, expectedStatus: http.StatusCreated,
		},
	}

	for i, itemTest := range testCases {
		t.Run(fmt.Sprint("%d#)%s", i, itemTest.name), func(t *testing.T) {
			handler := NewUserHandler(itemTest.service)

			r := gin.New()
			r.POST("/fido/users", handler.Create)

			request := httptest.NewRequest(http.MethodPost, "/fido/users", bytes.NewBufferString(itemTest.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, request)

			assert.Equal(t, itemTest.expectedStatus, recorder.Code, "Body %s", recorder.Body.String())

			if itemTest.expectedStatus == http.StatusCreated {
				assert.Contains(t, recorder.Body.String(), itemTest.service.user.Hash)
			}
		})
	}
}
