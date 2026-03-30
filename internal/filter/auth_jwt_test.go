package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTFilter_Required(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	f := JWTFilter{Secret: secret, Required: true, Skip: []string{"/admin"}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api", nil)
	c.Request = req
	if f.Handle(c, FromGin(c)) {
		t.Fatal("expected false without token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "u1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	s, err := tok.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	reqOK, _ := http.NewRequest(http.MethodGet, "/api", nil)
	reqOK.Header.Set("Authorization", "Bearer "+s)
	c2.Request = reqOK
	if !f.Handle(c2, FromGin(c2)) {
		t.Fatal("expected allow with valid token")
	}
}
