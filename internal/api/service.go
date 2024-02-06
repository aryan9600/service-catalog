package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/gin-gonic/gin"
)

// ListServicesOutput represents the output returned when fetching a list of Services.
type ListServicesOutput struct {
	Data []models.Service `json:"data"`
}

// GetServiceWithVersionsOutput represents the output returned when fetching a Service along with
// its related Version objects.
type GetServiceWithVersionsOutput struct {
	Data ServiceWithVersions `json:"data"`
}

// ServiceWithVersions represents a Service with its related Versions.
type ServiceWithVersions struct {
	models.Service
	Versions []models.Version `json:"versions"`
}

// ServiceOutput represents the output returned when fetching/creating/updating
// a single Service.
type ServiceOutput struct {
	Data models.Service `json:"data"`
}

// CreateVersionOutput represnets the output returned after creating a Version object.
type CreateVersionOutput struct {
	Data models.Version `json:"data"`
}

// ListServices godoc
// @Summary     List all services for the authenticated user.
// @Produce     json
// @Param       Authorization header string true "Bearer token"
// @Param       limit query int false "Limit results"
// @Param       offset query int false "Query offset"
// @Param       sortKey query string false "Key to sort records by"
// @Param       descending query bool false "Sort records in descending order"
// @Param       name query string false "Search records by name"
// @Success     200  {object}  ListServicesOutput
// @Router      /services [get]
//
// ListServices returns a list of services for the authenticated user based on the following query parameters:
func ListServices(c *gin.Context) {
	uID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}
	userID, ok := uID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}

	input := models.ListServicesInput{}
	input.UserID = userID

	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid query parameters: %s", err.Error())})
		return
	}

	services, err := models.ListServices(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to list services: %s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, ListServicesOutput{
		Data: services,
	})
}

// GetService   godoc
// @Summary     Get requested service.
// @Description If the 'versions' query param is absent/false, versions is omitted from the response.
// @Produce     json
// @Param       Authorization header string true "Bearer token"
// @Param       versions query bool false "Return related versions"
// @Success     200  {object}  ServiceOutput
// @Success     200  {object}  GetServiceWithVersionsOutput
// @Router      /service/{id} [get]
//
// GetService returns the requested Service based on the 'id' query parameter.
// If the 'versions' query parameter is present and set to 'true', the Version
// objects for that Service are also present in the response.
func GetService(c *gin.Context) {
	svcIdStr := c.Param("id")
	svcId, err := strconv.Atoi(svcIdStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("invalid service id: %s", svcIdStr)})
		return
	}

	uID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}
	userID, ok := uID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}

	withVersions := c.Query("versions")
	if withVersions == "true" {
		getServiceWithVersions(c, uint(svcId), userID)
	} else {
		getService(c, uint(svcId), userID)
	}
}

func getService(c *gin.Context, svcID, userID uint) {
	svc, err := models.GetService(svcID, userID)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("unable to fetch service: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to fetch service: %s", err.Error())})
		}
		return
	}

	c.JSON(http.StatusOK, ServiceOutput{
		Data: *svc,
	})
}

func getServiceWithVersions(c *gin.Context, svcID, userID uint) {
	output, err := models.GetServiceWithVersions(svcID, userID)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("unable to fetch service: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to fetch service: %s", err.Error())})
		}
		return
	}

	svc := models.Service{
		Model: models.Model{
			ID:        output[0].ServiceID,
			CreatedAt: output[0].ServiceCreatedAt,
			UpdatedAt: output[0].ServiceUpdatedAt,
		},
		Name:        output[0].Name,
		Description: output[0].Description,
		UserID:      int(userID),
	}
	var versions []models.Version
	for _, val := range output {
		versions = append(versions, models.Version{
			Model: models.Model{
				ID:        val.VersionID,
				CreatedAt: val.VersionCreatedAt,
				UpdatedAt: val.ServiceUpdatedAt,
			},
			Version:   val.Version,
			Changelog: val.Changelog,
			ServiceID: int(svcID),
		})
	}

	c.JSON(http.StatusOK, GetServiceWithVersionsOutput{
		Data: ServiceWithVersions{
			Service:  svc,
			Versions: versions,
		},
	})
}

// CreateService godoc
// @Summary Create a service.
// @Accept  json
// @Produce json
// @Param   Authorization header string true "Bearer token"
// @Param   service body   models.CreateServiceInput true  "Service JSON"
// @Success 201  {object}  ServiceOutput
// @Router  /service [post]
//
// CreateService creates a Service for the authenticated user.
// The request body must contain a name and an optional description.
func CreateService(c *gin.Context) {
	uID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}
	userID, ok := uID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}

	var input models.CreateServiceInput
	input.UserID = userID
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid service input: %s", err.Error())})
		return
	}

	service, err := models.CreateService(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to create service: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, ServiceOutput{
		Data: *service,
	})
}

// UpdateService godoc
// @Summary Update a service
// @Accept  json
// @Produce json
// @Param   Authorization header string true "Bearer token"
// @Param   version body   models.UpdateServiceInput true  "Service update JSON"
// @Success 201  {object}  ServiceOutput
// @Router  /service/{id} [patch]
//
// UpdateService updates the Service according to the provided input.
func UpdateService(c *gin.Context) {
	svcIdStr := c.Param("id")
	svcId, err := strconv.Atoi(svcIdStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("invalid service id: %s", svcIdStr)})
		return
	}

	uID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}
	userID, ok := uID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}

	var input models.UpdateServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid service update input: %s", err.Error())})
		return
	}
	if input.Name == "" && input.Description == "" {
		c.JSON(http.StatusNoContent, gin.H{"message": fmt.Sprintf("empty service update input")})
		return
	}

	svc, err := models.UpdateService(input, uint(svcId), userID)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("unable to update service: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to update service: %s", err.Error())})
		}
		return
	}
	c.JSON(http.StatusOK, ServiceOutput{
		Data: *svc,
	})
}

// CreateVersion godoc
// @Summary Create a version for a service
// @Accept  json
// @Produce json
// @Param   Authorization header string true "Bearer token"
// @Param   version body   models.CreateVersionInput true  "Version JSON"
// @Success 201  {object}  CreateVersionOutput
// @Router  /service/{id}/version [post]
//
// CreateVersion creates a new Version for the provided Service.
// The Service must exist in the database. The request body must contain
// the service id and a unique version string.
func CreateVersion(c *gin.Context) {
	svcIdStr := c.Param("id")
	svcId, err := strconv.Atoi(svcIdStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("invalid service id: %s", svcIdStr)})
		return
	}

	uID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}
	userID, ok := uID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch user details"})
		return
	}

	var input models.CreateVersionInput
	input.UserID = userID
	input.ServiceID = svcId

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid version input: %s", err.Error())})
		return
	}

	version, err := models.CreateVersion(input)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) || errors.Is(err, models.ErrUniqueConstraintViolation) {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("unable to create version: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to create version %s", err.Error())})
		}
		return
	}

	c.JSON(http.StatusCreated, CreateVersionOutput{
		Data: *version,
	})
}
