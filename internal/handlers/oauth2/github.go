package oauth2

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/repositorioes/cruds"
	"go-streamer/internal/utils"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	oa2 "golang.org/x/oauth2"
)

type GitHubUserResponse struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Company           string `json:"company"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
	Hireable          *bool  `json:"hireable"`
	Bio               string `json:"bio"`
	TwitterUsername   string `json:"twitter_username"`
	PublicRepos       int    `json:"public_repos"`
	PublicGists       int    `json:"public_gists"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type GithubOauthConfig struct {
	Config *oa2.Config
}

func NewGithubOauthConfig() *GithubOauthConfig {
	return &GithubOauthConfig{
		Config: &oa2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
			Scopes:       []string{"user:email"},
			Endpoint: oa2.Endpoint{
				AuthURL:  os.Getenv("GITHUB_AUTH_URL"),
				TokenURL: os.Getenv("GITHUB_TOKEN_URL"),
			},
		},
	}
}

func AuthGithub(c *gin.Context) {
	cfg := c.MustGet(utils.GITHUB_OAUTH_CONF_CTX_KEY).(*GithubOauthConfig)
	state, err := generateOAuthState()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create auth intent with github")
		return
	}

	// Encrypt state before storing cookie
	encryptedState, err := utils.EncryptCookieValue(state)
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "Failed to secure connection")
		return
	}

	setSecureStateCookie(c, stateCookieName, encryptedState, int(stateExpiration.Seconds()))
	url := cfg.Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func AuthGithubCallback(c *gin.Context) {
	cfg := c.MustGet(utils.GITHUB_OAUTH_CONF_CTX_KEY).(*GithubOauthConfig)
	resp, err := handleCallback(*cfg, c)

	if err != nil {
		c.String(http.StatusUnauthorized, "Where are you going?")
		return
	}

	ghUser, err := parseGitHubResponse(resp)

	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create user")
		return
	}

	userInfo := mapGitHubUserToCreateUser(ghUser)
	dbRepo := c.MustGet(utils.DB_REPO_CTX_KEY).(*repositorioes.DBRepo)
	newUser, err := cruds.NewUsersCrud(dbRepo).CreateOrUpdateUser(*userInfo)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create user")
		return
	}

	jwtToken, err := generateJWTToken(strconv.Itoa(int(newUser.ID)))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create auth session")
		return
	}

	c.JSON(http.StatusOK, models.SuccessAuthIntentResponse[*models.User]{
		Token:           jwtToken,
		UserInformation: newUser,
	})
}

func handleCallback(cfg GithubOauthConfig, c *gin.Context) (*http.Response, error) {
	state := c.Query("state")
	if state == "" {
		c.String(http.StatusBadRequest, "Missing state parameter")
		return nil, nil
	}

	// Get and decrypt stored state
	encryptedStoredState, err := c.Cookie(stateCookieName)
	if err != nil {
		return nil, err
	}

	storedState, err := utils.DecryptCookieValue(encryptedStoredState)
	if err != nil {
		return nil, err
	}

	setSecureStateCookie(c, stateCookieName, "", -1)

	if !hmac.Equal([]byte(storedState), []byte(state)) {
		return nil, errors.New("Callback state do not match cookie state")
	}

	code := c.Query("code")
	token, err := cfg.Config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	// Use the token to fetch user info
	client := cfg.Config.Client(context.TODO(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func parseGitHubResponse(resp *http.Response) (*GitHubUserResponse, error) {
	// Read the response body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Unmarshal the JSON response into the GitHubUserResponse struct
	var githubUser GitHubUserResponse
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub response: %v", err)
	}

	return &githubUser, nil
}

func mapGitHubUserToCreateUser(githubUser *GitHubUserResponse) *models.CreateUser {
	return &models.CreateUser{
		Name:           githubUser.Name,
		Email:          githubUser.Email,
		ProviderUserId: fmt.Sprintf("%d", githubUser.ID), // Convert ID to string
	}
}
