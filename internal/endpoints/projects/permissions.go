package projects

import (
	"net/http"
	"strconv"
	"strings"

	"packster/internal"
	"packster/internal/auth"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

type setPermissionRequest struct {
	UserID      int  `json:"user_id"`
	CanDownload bool `json:"can_download"`
	CanUpload   bool `json:"can_upload"`
	CanDelete   bool `json:"can_delete"`
}

type permissionEntryDTO struct {
	UserID      int    `json:"user_id"`
	DisplayName string `json:"display_name"`
	Project     int    `json:"project"`
	CanDownload bool   `json:"can_download"`
	CanUpload   bool   `json:"can_upload"`
	CanDelete   bool   `json:"can_delete"`
	IsOwner     bool   `json:"is_owner"`
}

// HandleListPermissions returns the permission grants on a project, joined
// with each grantee's display name. Restricted to the project owner.
func (h *ProjectsHandler) HandleListPermissions(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.ProjectRepo.GetByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if project.Owner != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can view permissions"})
		return
	}

	entries, err := h.PermissionRepo.ListByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]permissionEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, permissionEntryDTO{
			UserID:      e.Account,
			DisplayName: e.DisplayName,
			Project:     e.Project,
			CanDownload: e.CanDownload,
			CanUpload:   e.CanUpload,
			CanDelete:   e.CanDelete,
			IsOwner:     e.Account == project.Owner,
		})
	}
	c.JSON(http.StatusOK, out)
}

// HandleSetPermission upserts a permission row granting the named user the
// requested download/upload/delete flags on the project. Restricted to the
// project owner; the owner's own row cannot be modified through this endpoint.
func (h *ProjectsHandler) HandleSetPermission(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.ProjectRepo.GetByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if project.Owner != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can manage permissions"})
		return
	}

	var req setPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if req.UserID == project.Owner {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify owner permissions"})
		return
	}

	exists, err := h.UserRepo.UserExistsByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	perm := types.Permission{
		Account:     req.UserID,
		Project:     project.ID,
		CanDownload: req.CanDownload,
		CanUpload:   req.CanUpload,
		CanDelete:   req.CanDelete,
	}
	if err := h.PermissionRepo.Set(perm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perm)
}

type userCandidateDTO struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
}

// HandleSearchUsers prefix-searches users on the caller's host by display name
// for the permissions UI. Restricted to the project owner; the caller is
// excluded from results.
func (h *ProjectsHandler) HandleSearchUsers(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.ProjectRepo.GetByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if project.Owner != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can search users"})
		return
	}

	host, ok := internal.Hosts[sess.Host]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown host"})
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, []userCandidateDTO{})
		return
	}

	users, err := h.UserRepo.SearchByName(host.Id, q, sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]userCandidateDTO, 0, len(users))
	for _, u := range users {
		out = append(out, userCandidateDTO{ID: u.ID, DisplayName: u.DisplayName})
	}
	c.JSON(http.StatusOK, out)
}

// HandleRevokePermission deletes the permission row for a user on a project.
// Restricted to the project owner; the owner's own row cannot be revoked.
func (h *ProjectsHandler) HandleRevokePermission(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	project, err := h.ProjectRepo.GetByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if project.Owner != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can manage permissions"})
		return
	}

	if userID == project.Owner {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot revoke owner permissions"})
		return
	}

	if err := h.PermissionRepo.Delete(userID, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
