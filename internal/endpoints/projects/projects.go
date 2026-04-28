package projects

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"packster/internal"
	"packster/internal/auth"
	"packster/internal/requests"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

type projectDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Org        int    `json:"org"`
	WebURL     string `json:"web_url"`
	Repository int    `json:"repository"`
	Owner      int    `json:"owner"`
}

type importProjectRequest struct {
	SourceID  int    `json:"source_id"`
	SourceURL string `json:"source_url"`
	Org       int    `json:"org"`
}

// HandleListImported returns the projects imported into Packster that the
// caller has any permission row on, filtered to the host they are signed in
// against. Each row is enriched with metadata fetched from the upstream host.
func (h *ProjectsHandler) HandleListImported(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	host, ok := internal.Hosts[sess.Host]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown host: %s", sess.Host)})
		return
	}

	projects, err := h.ProjectRepo.ListAccessible(sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]projectDTO, 0, len(projects))
	for _, p := range projects {
		if p.Host != host.Id {
			continue
		}
		dto, err := h.fetchProjectMeta(sess, p)
		if err != nil {
			continue
		}
		out = append(out, dto)
	}

	c.JSON(http.StatusOK, out)
}

// HandleImport adopts an upstream repository as a Packster project. Verifies
// the caller belongs to the requested org, that the upstream project actually
// belongs to that org, and that the repository hasn't already been imported.
// The caller becomes the owner with full permissions.
func (h *ProjectsHandler) HandleImport(c *gin.Context) {
	sess, err := auth.ParseSession(c, h.Cfg.Secret)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	var req importProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SourceID == 0 || req.Org == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_id and org are required"})
		return
	}

	if !slices.Contains(sess.Orgs, req.Org) {
		c.JSON(http.StatusForbidden, gin.H{"error": "org not allowed for this user"})
		return
	}

	host, ok := internal.Hosts[sess.Host]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown host: %s", sess.Host)})
		return
	}

	exists, err := h.UserRepo.UserExistsByID(sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user no longer exists"})
		return
	}

	gp, err := requests.FetchGitlabProject(h.HTTP, sess.ProviderToken, host.Url, req.SourceID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if gp.Namespace.ID != req.Org {
		c.JSON(http.StatusForbidden, gin.H{"error": "project does not belong to org"})
		return
	}

	if existing, err := h.ProjectRepo.GetByHostRepository(host.Id, req.SourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "project already imported"})
		return
	}

	project, err := h.ProjectRepo.Import(sess.UserID, host.Id, req.SourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, projectDTO{
		ID:         project.ID,
		Name:       gp.Name,
		Org:        gp.Namespace.ID,
		WebURL:     gp.WebURL,
		Repository: project.Repository,
		Owner:      project.Owner,
	})
}

// HandleDeleteProject removes the project and every dependent row (products,
// versions, permissions, token_access) along with the on-disk version blobs.
// Restricted to the project owner.
func (h *ProjectsHandler) HandleDeleteProject(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can delete the project"})
		return
	}

	paths, err := h.ProjectRepo.Delete(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, p := range paths {
		removeBlob(p)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ProjectsHandler) fetchProjectMeta(sess *auth.Session, p types.Project) (projectDTO, error) {
	gp, err := requests.FetchGitlabProject(h.HTTP, sess.ProviderToken, sess.Host, p.Repository)
	if err != nil {
		return projectDTO{}, err
	}
	return projectDTO{
		ID:         p.ID,
		Name:       gp.Name,
		Org:        gp.Namespace.ID,
		WebURL:     gp.WebURL,
		Repository: p.Repository,
		Owner:      p.Owner,
	}, nil
}
