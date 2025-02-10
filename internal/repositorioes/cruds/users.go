package cruds

import (
	"errors"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"log"
	"time"

	"gorm.io/gorm"
)

type UsersCrud struct {
	dbRepo *repositorioes.DBRepo
}

func NewUsersCrud(dbRepo *repositorioes.DBRepo) *UsersCrud {
	return &UsersCrud{dbRepo: dbRepo}
}

func (c *UsersCrud) CreateOrUpdateUser(userInfo models.CreateUser) (*models.User, error) {
	var user models.User
	result := c.dbRepo.Db.Where("provider_user_id = ?", userInfo.ProviderUserId).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new user
			user = models.User{
				Email:          userInfo.Email,
				Name:           userInfo.Name,
				ProviderUserId: userInfo.ProviderUserId,
				Subscription:   *models.DefaultSubscription(),
			}

			if err := c.dbRepo.Db.Create(&user).Error; err != nil {
				log.Println("Error creating new user")
				return nil, err
			}
		} else {
			log.Println("Error deciding if user already exists")
			return nil, result.Error
		}
	} else {
		// Update existing user
		if user.Name != userInfo.Name || user.Email != userInfo.Email {
			user.UpdatedAt = time.Now()
		}

		user.Name = userInfo.Name
		user.Email = userInfo.Email

		if err := c.dbRepo.Db.Save(&user).Error; err != nil {
			log.Println("Error updating user information")
			return nil, err
		}
	}

	return &user, nil
}
