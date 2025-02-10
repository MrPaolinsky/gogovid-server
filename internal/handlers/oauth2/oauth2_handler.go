package oauth2

import "github.com/gin-gonic/gin"

func Oauth2Handler(r *gin.Engine) {
	g := r.Group("/oauth2")

	// Google
	g.GET("/google", AuthGoogle)
	g.GET("/google/callback", AuthGoogleCallback)

	//Github
	g.GET("/github", AuthGithub)
	g.GET("/github/callback", AuthGithubCallback)
}
