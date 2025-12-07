package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"rofl-bot/config"
	"rofl-bot/domain"
	"rofl-bot/repository"

	// "log"
	// "os"

	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg  *config.Config
	repo *repository.AuthRep
}

func NewAuthService(cfg *config.Config, repo *repository.AuthRep) *AuthService {
	return &AuthService{cfg: cfg, repo: repo}
}

func (a *AuthService) VerifyTG(InitData string) (map[string]string, error) {
	var BOT_TOKEN = a.cfg.BOT_TOKEN

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

func (a *AuthService) GenerateAccessToken(userID uint, username string) (string, error) {
	var jwtSecret = a.cfg.JWT_SECRET

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) CheckRefreshToken(refresh string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(a.cfg.JWT_SECRET))
	mac.Write([]byte(refresh))
	hash := hex.EncodeToString(mac.Sum(nil))

	token, err := a.repo.GetToken(hash)
	if err != nil {
		return false, err
	}

	if token.Revoked {
		err := a.repo.DeleteToken(hash)
		if err != nil {
			return false, err
		}
		return false, errors.New("token is revoked")
	}

	if token.ExpiresIn.Before(time.Now()) {
		err := a.repo.DeleteToken(hash)
		if err != nil {
			return false, err
		}
		return false, errors.New("token is expired")
	}

	return true, nil
}

func (a *AuthService) GenerateRefreshToken(userID uint, username string) (string, error) {
	token := uuid.New().String()

	hash := hmac.New(sha256.New, []byte(a.cfg.JWT_SECRET))
	hash.Write([]byte(token))

	tokenObj := domain.Token{
		Refresh:   hex.EncodeToString(hash.Sum(nil)),
		UserId:    userID,
		Revoked:   false,
		ExpiresIn: time.Now().Add(time.Hour * 24 * 30)}

	_, err := a.repo.GetTokenByUserId(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := a.repo.CreateToken(&tokenObj, username); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		if err := a.repo.UpdateToken(userID, &tokenObj); err != nil {
			return "", err
		}
	}

	return token, nil
}

func (a *AuthService) ClearTokens(userID uint) error {
	return a.repo.DeleteTokensByUserId(userID)
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtSecret := []byte(cfg.JWT_SECRET)

		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(401, domain.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, domain.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		expTime, err := claims.GetExpirationTime()
		if err != nil {
			c.JSON(401, domain.ErrorResponse{Error: "Invalid token expiration time"})
			c.Abort()
			return
		}

		if time.Now().After(expTime.Time) {
			c.JSON(401, domain.ErrorResponse{Error: "Access token expired"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("username", claims["username"])

		c.Next()
	}
}
