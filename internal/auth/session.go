package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Session struct {
	UserID        int
	ProviderToken string
	Host          string
	Orgs          []int
}

func ParseSession(c *gin.Context, secret string) (*Session, error) {
	raw := extractBearer(c.GetHeader("Authorization"))
	if raw == "" {
		return nil, fmt.Errorf("missing token")
	}

	tok, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	subStr, _ := claims["sub"].(string)
	id, err := strconv.Atoi(subStr)
	if err != nil {
		return nil, fmt.Errorf("invalid subject")
	}

	providerToken, _ := claims["token"].(string)

	var hostURL string
	if hostMap, ok := claims["host"].(map[string]any); ok {
		hostURL, _ = hostMap["url"].(string)
	}

	var orgs []int
	if rawOrgs, ok := claims["orgs"].([]any); ok {
		for _, v := range rawOrgs {
			if f, ok := v.(float64); ok {
				orgs = append(orgs, int(f))
			}
		}
	}

	return &Session{
		UserID:        id,
		ProviderToken: providerToken,
		Host:          hostURL,
		Orgs:          orgs,
	}, nil
}

func Unauthorized(c *gin.Context, err error) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
