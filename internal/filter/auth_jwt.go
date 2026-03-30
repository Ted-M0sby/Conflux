package filter

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTFilter validates Bearer JWT when required.
type JWTFilter struct {
	Secret   []byte
	Required bool
	Skip     []string // path prefixes e.g. /admin
}

func (f JWTFilter) Name() string { return "jwt" }

func (f JWTFilter) Order() int { return 10 }

func (f JWTFilter) Handle(c *gin.Context, ctx *Context) bool {
	path := c.Request.URL.Path
	for _, p := range f.Skip {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	if !f.Required {
		// Optional: still parse if present for downstream
		if tok := bearerToken(c); tok != "" {
			if claims, err := parseJWT(tok, f.Secret); err == nil {
				ctx.Claims = claims
			}
		}
		return true
	}
	tok := bearerToken(c)
	if tok == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return false
	}
	claims, err := parseJWT(tok, f.Secret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return false
	}
	ctx.Claims = claims
	return true
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	parts := strings.SplitN(strings.TrimSpace(h), " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseJWT(token string, secret []byte) (jwt.MapClaims, error) {
	var claims jwt.MapClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
