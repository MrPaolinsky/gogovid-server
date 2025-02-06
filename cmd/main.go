package main

import (
	"go-streamer/internal/handlers"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/utils"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	s3Repo := repositorioes.NewS3Repo()

	r := gin.Default()

	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = true

	r.Use(cors.New(corsConf))
	r.Use(func(c *gin.Context) {
		c.Set(utils.S3_REPO_CTX_KEY, s3Repo)
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		repo := c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)
		repo.TestListObject()
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.GET("/video/:fileId", handlers.ServeVideo)

	r.POST("/video", handlers.UploadVideo)

	r.Run()
}
