package database

import (
	"log"

	"github.com/google/uuid"
)

func (a *DatabaseAdapter) Save(data *DummyOrm) (uuid.UUID, error) {
	if err := a.db.Create(data).Error; err != nil {
		log.Println("Cant create data", err)
		return uuid.Nil, err
	}

	return data.UserId, nil
}

func (a *DatabaseAdapter) GetByUUID(uuid uuid.UUID) (DummyOrm, error) {
	var res DummyOrm
	if err := a.db.First(&res, "user_id = ?", uuid).Error; err != nil {
		log.Println("Cant get data", err)
		return res, err
	}

	return res, nil
}
