package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"rofl-bot/config"

	// "log"
	"os"

	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (a *AuthService) VerifyTG(InitData string) (map[string]string, error) {
	var BOT_TOKEN = os.Getenv("BOT_TOKEN")

	if InitData == "" {
		return nil, errors.New("InitData is empty")
	}

	values, error := url.ParseQuery(InitData)
	if error != nil {
		return nil, error
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, errors.New("hash is empty")
	}
	values.Del("hash")

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := make([]string, 0, len(keys))
	for _, k := range keys {
		data = append(data, k+"="+values.Get(k))
	}

	data_check_string := strings.Join(data, "\n")
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(BOT_TOKEN))
	secret := secretKey.Sum(nil)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data_check_string))

	expected := hex.EncodeToString(mac.Sum(nil))
	recieved := hash

	if !hmac.Equal([]byte(expected), []byte(recieved)) {
		return nil, errors.New("hash is invalid")
	}

	res := make(map[string]string, len(values))
	for k, v := range values {
		res[k] = v[0]
	}

	return res, nil
}

func (a *AuthService) GenerateAccessToken(userID string) (string, error) {
	var jwtSecret = os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"user_id": userID,
		// "exp":     time.Now().Add(time.Hour * 24).Unix(),
		"exp": time.Now().Add(time.Minute * 3).Unix(), //для тестов
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) GenerateRefreshToken() (string, error) {
	token := uuid.New().String()
	//записать в бд
	return token, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtSecret = os.Getenv("JWT_SECRET")

		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])

		c.Next()
	}
}
