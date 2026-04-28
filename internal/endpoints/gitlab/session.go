package gitlab

import (
	"net/http"
	"os"

	"packster/internal/auth"

	"github.com/gin-gonic/gin"
)

// HandleSession verifies the bearer JWT and that the user still exists in the
// database. If the user has been deleted but a stale token is still in use,
// any data they leave behind is purged before the unauthorized response.
func (h *GitlabHandler) HandleSession(c *gin.Context) {
	sess, err := h.parseSession(c)
	if err != nil {
		auth.Unauthorized(c, err)
		return
	}

	exists, err := h.Repo.UserExistsByID(sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		if paths, perr := h.Repo.PurgeUserData(sess.UserID); perr == nil {
			for _, p := range paths {
				if p != "" {
					_ = os.Remove(p)
				}
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user no longer exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *GitlabHandler) parseSession(c *gin.Context) (*auth.Session, error) {
	return auth.ParseSession(c, h.Cfg.Secret)
}
