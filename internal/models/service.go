package models

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm/clause"
)

const ServiceTableName = "services"

// Service represents a single service in the catalog.
type Service struct {
	Model
	Name        string `json:"name"`
	Description string `json:"description"`
	// Versions contains the different versions of this service.
	// It helps us fetch the versions without a JOIN query.
	Versions pq.StringArray `json:"versions" gorm:"type:varchar(50)[]"`
	UserID   int            `json:"userID"`
}

// ListServicesInput represnts the different input parameters that can be
// included in the database query. The form struct tags allows for convinient
// query parameter validation.
type ListServicesInput struct {
	Limit      int `form:"limit"`
	Offset     int `form:"offset"`
	UserID     uint
	SortKey    string `form:"sortKey"`
	Descending bool   `form:"descending"`
	Name       string `form:"name"`
}

// ListServices returns a list of Service objects based on the different input parameters.
func ListServices(input ListServicesInput) ([]Service, error) {
	var services []Service
	db := DB.Table(ServiceTableName)

	if input.UserID != 0 {
		db = db.Where("user_id = ?", input.UserID)
	}
	if input.Name != "" {
		match := "%" + input.Name + "%"
		db = db.Where("name LIKE ? ", match)
	}
	if input.Limit != 0 {
		db = db.Limit(input.Limit).Offset(input.Offset)
	}

	if input.SortKey != "" {
		if ok := isValidSortKey(input.SortKey); !ok {
			return nil, fmt.Errorf("invalid sort key: %s", input.SortKey)
		}
		orderClause := input.SortKey
		if input.Descending {
			orderClause += " DESC"
		}
		db = db.Order(orderClause)
	}

	if err := db.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetServiceWithVersionsTxOutput represents the various columns returned by the
// database query made in GetServiceWithVersions().
type GetServiceWithVersionsTxOutput struct {
	ServiceID        uint
	ServiceCreatedAt time.Time
	ServiceUpdatedAt time.Time
	Name             string
	Description      string
	VersionID        uint
	VersionCreatedAt time.Time
	VersionUpdatedAt time.Time
	Version          string
	Changelog        string
}

// getServiceWithVersionsTxFields returns the SQL compatible stringfied names
// of the fields GetServiceWithVersionsTxOutput.
func getServiceWithVersionsTxFields() []string {
	return []string{
		"services.id as service_id",
		"services.created_at as service_created_at",
		"services.updated_at as service_updated_at",
		"versions.id as version_id",
		"versions.created_at as version_created_at",
		"versions.updated_at as version_updated_at",
		"services.name",
		"services.description",
		"versions.version",
		"versions.changelog",
	}
}

// GetServiceWithVersions returns the requested Service for the provided ID along
// of the Version objects belonging to this Service.
func GetServiceWithVersions(svcID uint, userID uint) ([]GetServiceWithVersionsTxOutput, error) {
	output := make([]GetServiceWithVersionsTxOutput, 0)
	db := DB.Table(ServiceTableName)

	fields := getServiceWithVersionsTxFields()
	db = db.Select(fields)
	db = db.Joins(fmt.Sprintf("RIGHT JOIN %s ON services.id=versions.service_id", VersionTableName))
	db = db.Where("services.id = ?", svcID)
	if userID != 0 {
		db = db.Where("services.user_id = ?", userID)
	}

	err := db.Find(&output).Error
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, ErrRecordNotFound
	}

	return output, nil
}

// GetService returns the Service for the provided ID.
func GetService(svcID uint, userID uint) (*Service, error) {
	db := DB.Table(ServiceTableName)
	db = db.Where("id = ?", svcID)
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}

	var service Service
	err := db.Find(&service).Error
	if err != nil {
		return nil, err
	}
	if service.ID == 0 {
		return nil, ErrRecordNotFound
	}

	return &service, nil
}

// CreateServiceInput represents the input required to create a Service.
type CreateServiceInput struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description"`
	UserID      uint
}

// CreateService creates a new Service.
func CreateService(input CreateServiceInput) (*Service, error) {
	db := DB.Table(ServiceTableName)
	service := Service{
		Name:        input.Name,
		Description: input.Description,
		UserID:      int(input.UserID),
	}
	if err := db.Create(&service).Error; err != nil {
		return nil, err
	}

	return &service, nil
}

// CreateServiceInput represents the input required to create a Service.
type UpdateServiceInput struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description"`
}

func UpdateService(input UpdateServiceInput, id uint, userID uint) (*Service, error) {
	var updated Service
	db := DB.Model(&updated)
	db = db.Where("id = ?", id)
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}

	svc := Service{Name: input.Name, Description: input.Description}
	if err := db.Clauses(clause.Returning{}).Updates(svc).Error; err != nil {
		return nil, err
	}
	if updated.ID == 0 {
		return nil, ErrRecordNotFound
	}
	return &updated, nil
}

func isValidSortKey(sortKey string) bool {
	switch sortKey {
	case "name", "created_at", "updated_at":
		return true
	default:
		return false
	}
}
