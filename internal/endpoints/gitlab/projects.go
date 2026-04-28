package gitlab

import (
	"net/http"
	"slices"
	"strconv"

	"packster/internal/auth"
	"packster/internal/requests"

	"github.com/gin-gonic/gin"
)

type candidateProject struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	FullPath       string `json:"full_path"`
	WebURL         string `json:"web_url"`
	Org            int    `json:"org"`
	Visibility     string `json:"visibility,omitempty"`
	LastActivityAt string `json:"last_activity_at,omitempty"`
}

// HandleListCandidates returns projects on the configured GitLab host that the
// authenticated user can import for the given org. Filters by minimum access
// level (defaults to 50 = Owner). The org must be one the user belongs to.
func (h *GitlabHandler) HandleListCandidates(c *gin.Context) {
	sess, err := h.parseSession(c)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	org, err := strconv.Atoi(c.Query("org"))
	if err != nil || org == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org is required"})
		return
	}

	if !slices.Contains(sess.Orgs, org) {
		c.JSON(http.StatusForbidden, gin.H{"error": "org not allowed for this user"})
		return
	}

	minAccess, _ := strconv.Atoi(c.Query("min_access_level"))
	if minAccess <= 0 {
		minAccess = 50
	}

	client := &http.Client{}
	projects, err := requests.FetchGitlabGroupProjects(client, sess.ProviderToken, sess.Host, org, minAccess)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	out := make([]candidateProject, 0, len(projects))
	for _, p := range projects {
		out = append(out, candidateProject{
			ID:             p.ID,
			Name:           p.Name,
			FullPath:       p.PathWithNamespace,
			WebURL:         p.WebURL,
			Org:            org,
			Visibility:     p.Visibility,
			LastActivityAt: p.LastActivityAt,
		})
	}

	c.JSON(http.StatusOK, out)
}
