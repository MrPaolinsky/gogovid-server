package repositorioes

import (
	"go-streamer/internal/models"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBRepo struct {
	Db *gorm.DB
}

func NewDBRepo(engine string) *DBRepo {
	var db *gorm.DB
	var err error

	if engine == "postgresql" {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN: os.Getenv("DATABASE_URI"),
		}), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})

		if err != nil {
			log.Panic("Error creating postgresql connection")
		}
	} else if engine == "mariadb" {
		db, err = gorm.Open(
			mysql.New(mysql.Config{
				DSN: os.Getenv("DATABASE_URI"),
			}),
			&gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
			},
		)
		if err != nil {
			log.Panic("Error creating mariadb connection")
		}
	} else {
		log.Panic("Please provide a valid DB engine, mariadb and postgresql are supported values.")
	}

	repo := &DBRepo{
		Db: db,
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		repo.migrateSchemas()
	}

	return repo
}

func (r *DBRepo) migrateSchemas() {
	err := r.Db.AutoMigrate(
		&models.Subscription{},
		&models.User{},
		&models.Video{},
		&models.ApiToken{},
		&models.TokenHistory{},
	)

	if err != nil {
		log.Println("Failed to auto migrate schemas")
	}
}
