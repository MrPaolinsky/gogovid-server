package oauth2

import (
	"context"
	"go-streamer/internal/utils"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	oa2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOauthConfig struct {
	Config *oa2.Config
}

func NewGoogleOauthConfig() *GoogleOauthConfig {
	return &GoogleOauthConfig{
		Config: &oa2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func AuthGoogle(c *gin.Context) {
	cfg := c.MustGet(utils.GOOGLE_OAUTH_CONF_CTX_KEY).(*GoogleOauthConfig)
	url := cfg.Config.AuthCodeURL("TODO")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func AuthGoogleCallback(c *gin.Context) {
	cfg := c.MustGet(utils.GOOGLE_OAUTH_CONF_CTX_KEY).(*GoogleOauthConfig)

	if c.Query("state") != "TODO" {
		c.String(http.StatusBadRequest, "InvalidState")
		return
	}

	code := c.Query("code")
	token, err := cfg.Config.Exchange(context.TODO(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to exchange token")
		return
	}

	// Use the token to fetch user info
	client := cfg.Config.Client(context.TODO(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get user info")
		return
	}
	defer resp.Body.Close()

	log.Println("User info: ", resp)
	c.String(http.StatusOK, "Logged in with Google!")
}
