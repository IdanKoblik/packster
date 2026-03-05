package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"artifactor/pkg/requests"
	"artifactor/pkg/users"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	userExistsFn func(username string) (bool, error)
	createUserFn func(request *requests.RegisterRequest) error
}

func (m *mockRepo) UserExists(username string) (bool, error) {
	return m.userExistsFn(username)
}

func (m *mockRepo) CreateUser(request *requests.RegisterRequest) error {
	return m.createUserFn(request)
}

func setupRouter(handler *AuthHandler, signingKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", func(c *gin.Context) {
		handler.HandleRegister(c, signingKey)
	})
	return r
}

func makeRegisterRequest(t *testing.T, r *gin.Engine, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}
	req, _ := http.NewRequest(http.MethodPost, "/register", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validRequest() requests.RegisterRequest {
	return requests.RegisterRequest{
		Username: "alice",
		Name:     "Alice",
		Mail:     "alice@example.com",
		Password: "secret",
		Permissions: users.UserPermissions{
			Upload: true,
			Delete: false,
		},
	}
}

func TestHandleRegister_MissingBody(t *testing.T) {
	handler := &AuthHandler{Repo: &mockRepo{}}
	r := setupRouter(handler, "key")

	req, _ := http.NewRequest(http.MethodPost, "/register", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing request body")
}

func TestHandleRegister_MissingUsername(t *testing.T) {
	handler := &AuthHandler{Repo: &mockRepo{}}
	r := setupRouter(handler, "key")

	body := validRequest()
	body.Username = ""
	w := makeRegisterRequest(t, r, body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing username field")
}

func TestHandleRegister_MissingPassword(t *testing.T) {
	handler := &AuthHandler{Repo: &mockRepo{}}
	r := setupRouter(handler, "key")

	body := validRequest()
	body.Password = ""
	w := makeRegisterRequest(t, r, body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing password field")
}

func TestHandleRegister_MissingName(t *testing.T) {
	handler := &AuthHandler{Repo: &mockRepo{}}
	r := setupRouter(handler, "key")

	body := validRequest()
	body.Name = ""
	w := makeRegisterRequest(t, r, body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing name field")
}

func TestHandleRegister_UserExistsError(t *testing.T) {
	repo := &mockRepo{
		userExistsFn: func(username string) (bool, error) {
			return false, errors.New("redis unavailable")
		},
	}
	handler := &AuthHandler{Repo: repo}
	r := setupRouter(handler, "key")

	w := makeRegisterRequest(t, r, validRequest())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleRegister_UserAlreadyExists(t *testing.T) {
	repo := &mockRepo{
		userExistsFn: func(username string) (bool, error) {
			return true, nil
		},
	}
	handler := &AuthHandler{Repo: repo}
	r := setupRouter(handler, "key")

	w := makeRegisterRequest(t, r, validRequest())

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already exists")
}

func TestHandleRegister_CreateUserError(t *testing.T) {
	repo := &mockRepo{
		userExistsFn: func(username string) (bool, error) {
			return false, nil
		},
		createUserFn: func(request *requests.RegisterRequest) error {
			return errors.New("db error")
		},
	}
	handler := &AuthHandler{Repo: repo}
	r := setupRouter(handler, "key")

	w := makeRegisterRequest(t, r, validRequest())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleRegister_Success(t *testing.T) {
	repo := &mockRepo{
		userExistsFn: func(username string) (bool, error) {
			return false, nil
		},
		createUserFn: func(request *requests.RegisterRequest) error {
			return nil
		},
	}
	handler := &AuthHandler{Repo: repo}
	r := setupRouter(handler, "test-signing-key")

	w := makeRegisterRequest(t, r, validRequest())

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotEmpty(t, w.Body.String())
}
