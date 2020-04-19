package middleware

import (
	"fmt"
	"github.com/dovbysh/go-skeleton/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const AuthorizationHeader = "Authorization"
const BearerHeader = "Bearer "
const UserKey = "UserKey"

type SkipperFunc func(*gin.Context) bool
type Skippers []SkipperFunc
type UserGetter func(AuthKey string) (*models.User, error)

func AllowPathPrefixSkipper(prefixes ...string) SkipperFunc {
	return func(c *gin.Context) bool {
		path := c.Request.URL.Path
		pathLen := len(path)

		for _, p := range prefixes {
			if pl := len(p); pathLen >= pl && path[:pl] == p {
				return true
			}
		}
		return false
	}
}

func SkipHandler(c *gin.Context, skippers ...SkipperFunc) bool {
	for _, skipper := range skippers {
		if skipper(c) {
			return true
		}
	}
	return false
}

func UserAuth(skippers Skippers, geter UserGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if SkipHandler(c, skippers...) {
			c.Next()
			return
		}

		a := c.GetHeader(AuthorizationHeader)
		if a == "" || !strings.HasPrefix(a, BearerHeader) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fmt.Errorf("no auth headers"))
			return
		}
		authKey := a[len(BearerHeader):]

		user, err := geter(authKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		c.Set(UserKey, user)
		c.Next()
	}
}

func GetUser(c *gin.Context) (*models.User, error) {
	u, exists := c.Get(UserKey)
	if !exists {
		return nil, fmt.Errorf("no user")
	}

	user, ok := u.(*models.User)
	if !ok {
		return nil, fmt.Errorf("not user")
	}

	return user, nil
}
