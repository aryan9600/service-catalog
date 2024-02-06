package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserAuthInput represents the user authentication credentials.
type UserAuthInput struct {
	Username string `json:"username" binding:"required,max=20"`
	Password string `json:"password" binding:"required,max=20"`
}

// RegisterOutput represents the output returned after a user is registered successfully.
type RegisterOutput struct {
	Data models.User `json:"data"`
}

// LoginOutput represents the output returned after a user logs in successfully.
type LoginOutput struct {
	AccessToken string `json:"accessToken"`
}

// Register godoc
// @Summary Register a user
// @Accept  json
// @Produce json
// @Param   creds body     UserAuthInput  true  "Auth creds JSON"
// @Success 201  {object}  RegisterOutput
// @Router  /auth/register [post]
//
// Register registers a new user.
func Register(c *gin.Context) {
	var input UserAuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid registration input: %s", err.Error())})
		return
	}

	user, err := models.CreateUser(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, models.ErrUniqueConstraintViolation) {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("unable to create user: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to create user: %s", err.Error())})
		}
		return
	}

	c.JSON(http.StatusCreated, RegisterOutput{
		Data: *user,
	})
}

// Login    godoc
// @Summary Login a user
// @Accept  json
// @Produce json
// @Param   creds body     UserAuthInput  true  "Auth creds JSON"
// @Success 200  {object}  LoginOutput
// @Router  /auth/login [post]
//
// Login returns an access token for the user, if found.
func Login(c *gin.Context) {
	var input UserAuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid login input: %s", err.Error())})
		return
	}

	user, err := models.GetUserByUsername(input.Username)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("unable to fetch user: %s", err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to fetch user: %s", err.Error())})
		}
		return
	}

	if !verifyPassword(user.Password, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid password"})
		return
	}
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to create JWT: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, LoginOutput{
		AccessToken: token,
	})
}

func verifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
