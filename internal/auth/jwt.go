package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	tokenLifespan int
	signingKey    []byte
)

// SetTokenGenerationConfig reads the token generation configuration from env vars.
// This is NEEDS to be called before using the below JWT methods.
func SetTokenGenerationConfig() error {
	tokenLifespanStr := os.Getenv("TOKEN_HOUR_LIFESPAN")
	var err error
	if tokenLifespanStr == "" {
		err = fmt.Errorf("unable to read env var TOKEN_HOUR_LIFESPAN")
		return err
	}
	tokenLifespan, err = strconv.Atoi(tokenLifespanStr)
	if err != nil {
		err = fmt.Errorf("invalid value for env var TOKEN_HOUR_LIFESPAN: %s; must be an integer", tokenLifespanStr)
		return err
	}

	signingKey = []byte(os.Getenv("JWT_SIGNING_KEY"))
	if len(signingKey) == 0 {
		err = fmt.Errorf("unable to read env var JWT_SIGNING_KEY")
		return err
	}
	return nil
}

// GenerateToken generates a JWT which encodes the provided user ID and
// expires after the configured number of hours.
func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(tokenLifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey)

}

// CheckTokenValidity checks if the token is valid, returning an error if it isn't.
func CheckTokenValidity(token string) error {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ExtractUserIDFromToken extracts the user ID from the provided JWT string.
func ExtractUserIDFromToken(token string) (uint, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unable to extract user id: unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return 0, fmt.Errorf("unable to extract user id: unable to parse JWT: %s", err.Error())
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if ok && parsedToken.Valid {
		val, ok := claims["user_id"]
		if !ok {
			return 0, fmt.Errorf("invalid token: unable to find user id in token claims")
		}
		userID, ok := val.(float64)
		if !ok {
			return 0, fmt.Errorf("invalid token: unexpected user id type present in token")
		}
		return uint(userID), nil
	}
	return 0, nil
}
