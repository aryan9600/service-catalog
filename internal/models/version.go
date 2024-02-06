package models

import (
	"strings"

	"gorm.io/gorm"
)

const VersionTableName = "versions"

// Version represents a version of a Service.
type Version struct {
	Model
	Version   string `json:"version"`
	ServiceID int    `json:"serviceID"`
	Changelog string `json:"changelog"`
}

type CreateVersionInput struct {
	Version   string `json:"version" binding:"required,max=50"`
	ServiceID int
	Changelog string `json:"changelog"`
	UserID    uint
}

func CreateVersion(input CreateVersionInput) (*Version, error) {
	version := &Version{
		Version:   input.Version,
		ServiceID: input.ServiceID,
		Changelog: input.Changelog,
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var service Service
		if err := tx.Model(&service).Where("id = ?", input.ServiceID).Where("user_id = ?", input.UserID).Find(&service).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				return ErrUniqueConstraintViolation
			}
			return err
		}
		if service.ID == 0 {
			return ErrRecordNotFound
		}
		if err := tx.Model(version).Create(version).Error; err != nil {
			return err
		}
		service.Versions = append(service.Versions, version.Version)
		if err := tx.Model(&service).Save(&service).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}
