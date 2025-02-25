package cruds

import (
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
)

type DRMCrud struct {
	dbRepo *repositorioes.DBRepo
}

func NewDRMCrud(dbRepo *repositorioes.DBRepo) *DRMCrud {
	return &DRMCrud{dbRepo: dbRepo}
}

func (c *DRMCrud) StoreKey(key *models.DRMKey) error {
	return c.dbRepo.Db.Create(key).Error
}

func (c *DRMCrud) GetKey(keyId string) (*models.DRMKey, error) {
	var key models.DRMKey
	err := c.dbRepo.Db.Where("key_id = ?", keyId).First(&key).Error
	return &key, err
}

func (c *DRMCrud) StoreManyKeys(keys []*models.DRMKey) error {
	return c.dbRepo.Db.Create(keys).Error
}
