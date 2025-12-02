package service

import (
	"crypto/hmac"
	"crypto/sha256"

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

const BOT_TOKEN = "8476710845:AAGm2wxJRCuHS_3nmVqqbzadp88aQVU3nTQ"

// var jwtSecret = []byte(os.Getenv("JWT_SECRET")) //на проде поставить в env
var jwtSecret = "crazy0little0alena"

func VerifyTG(InitData string) (map[string]string, error) {
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

func GenerateAccessToken(userID string) (string, error) {
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

func GenerateRefreshToken() (string, error) {
	token := uuid.New().String()
	//записать в бд
	return token, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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

// @Summary Авторизация через Telegram
// @Description Проверяет initData и выдаёт access + refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param InitData body object{init_data=string} true "Init Data"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /alena-rofl/auth [post]
func Auth(c *gin.Context) {
	var body struct {
		InitData string `json:"init_data"`
	}
	c.BindJSON(&body)

	data, err := VerifyTG(body.InitData)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID := data["user_id"]
	accessToken, err := GenerateAccessToken(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    180,
	})
}

// @Summary Обновление токена
// @Description После истечения access токена использовать это. принимает refresh токен, отдает новый access. Пока не работает!!
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body object{refresh_token=string} true "Refresh Token"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alena-rofl/refresh [post]
func Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.BindJSON(&body)

	var userID string
	var expires time.Time

	//вытащить из бд

	if time.Now().After(expires) {
		c.JSON(401, gin.H{"error": "refresh token expired"})
		return
	}

	accessToken, err := GenerateAccessToken(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": body.RefreshToken,
		"expires_in":    180,
	})

}
