package product

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleModify_UnknownAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/badaction", nil)
	c.Params = gin.Params{{Key: "action", Value: "badaction"}}

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unknown action")
}

func TestHandleModify_DeleteVersion_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteVersion", bytes.NewBufferString("invalid json"))
	c.Params = gin.Params{{Key: "action", Value: "deleteVersion"}}

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleModify_DeleteVersion_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.DeleteVersionRequest{Product: "myproduct", Version: "1.0.0"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteVersion", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "deleteVersion"}}
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("DeleteVersion", "myproduct", "1.0.0", "mytoken", false).Return(errors.New("delete error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "delete error")
	repo.AssertExpectations(t)
}

func TestHandleModify_DeleteVersion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.DeleteVersionRequest{Product: "myproduct", Version: "1.0.0"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteVersion", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "deleteVersion"}}
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("DeleteVersion", "myproduct", "1.0.0", "mytoken", false).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestHandleModify_DeleteToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteToken", bytes.NewBufferString("invalid json"))
	c.Params = gin.Params{{Key: "action", Value: "deleteToken"}}

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleModify_DeleteToken_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.DeleteProductTokenRequest{Product: "myproduct", Token: "target-token"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteToken", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "deleteToken"}}
	c.Set("token", "mytoken")
	c.Set("admin", false)

	repo := &mockProductRepo{}
	repo.On("DeleteToken", "myproduct", "mytoken", "target-token", false).Return(errors.New("delete token error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "delete token error")
	repo.AssertExpectations(t)
}

func TestHandleModify_DeleteToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.DeleteProductTokenRequest{Product: "myproduct", Token: "target-token"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products/modify/deleteToken", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "deleteToken"}}
	c.Set("token", "mytoken")
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("DeleteToken", "myproduct", "mytoken", "target-token", true).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestHandleModify_AddToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/products/modify/addToken", bytes.NewBufferString("invalid json"))
	c.Params = gin.Params{{Key: "action", Value: "addToken"}}

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleModify_AddToken_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	perms := types.TokenPermissions{Download: true}
	body, _ := json.Marshal(types.AddProductTokenRequest{Product: "myproduct", Token: "new-token", Permissions: perms})
	c.Request = httptest.NewRequest(http.MethodPut, "/products/modify/addToken", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "addToken"}}
	c.Set("token", "mytoken")
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("AddToken", "myproduct", "mytoken", "new-token", perms, true).Return(errors.New("add token error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "add token error")
	repo.AssertExpectations(t)
}

func TestHandleModify_AddToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	perms := types.TokenPermissions{Download: true}
	body, _ := json.Marshal(types.AddProductTokenRequest{Product: "myproduct", Token: "new-token", Permissions: perms})
	c.Request = httptest.NewRequest(http.MethodPut, "/products/modify/addToken", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "action", Value: "addToken"}}
	c.Set("token", "mytoken")
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("AddToken", "myproduct", "mytoken", "new-token", perms, true).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleModify(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	repo.AssertExpectations(t)
}
