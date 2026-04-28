package projects

import (
	"net/http"

	"packster/internal/auth"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

type accessKind int

const (
	accessRead accessKind = iota
	accessUpload
	accessDelete
)

func (h *ProjectsHandler) authorize(c *gin.Context, projectID int, kind accessKind) (sess *auth.Session, project *types.Project, perm *types.Permission, ok bool) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return nil, nil, nil, false
	}

	project, err = h.ProjectRepo.GetByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, nil, nil, false
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return nil, nil, nil, false
	}

	perm, err = h.PermissionRepo.Get(sess.UserID, project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, nil, nil, false
	}
	if perm == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to this project"})
		return nil, nil, nil, false
	}

	allowed := false
	switch kind {
	case accessRead:
		allowed = perm.CanDownload || perm.CanUpload || perm.CanDelete
	case accessUpload:
		allowed = perm.CanUpload
	case accessDelete:
		allowed = perm.CanDelete
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return nil, nil, nil, false
	}

	return sess, project, perm, true
}
