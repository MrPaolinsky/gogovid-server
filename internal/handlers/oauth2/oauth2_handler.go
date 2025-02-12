package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const stateCookieName = "oauth-state"
const stateExpiration = 1 * time.Minute

func Oauth2Handler(r *gin.Engine) {
	g := r.Group("/oauth2")

	// Google
	g.GET("/google", AuthGoogle)
	g.GET("/google/callback", AuthGoogleCallback)

	//Github
	g.GET("/github", AuthGithub)
	g.GET("/github/callback", AuthGithubCallback)
}

func generateJWTToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		log.Println("Error generating new JWT token")
		return "", nil
	}

	return tokenString, nil
}

func setSecureStateCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetCookie(name, value, maxAge, "", os.Getenv("API_DOMAIN"), true, true)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
}

func generateOAuthState() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
