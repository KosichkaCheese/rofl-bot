package controller

import (
	"time"

	"rofl-bot/service"

	"github.com/gin-gonic/gin"
)

type AuthCtrl struct {
	service *service.AuthService
}

func NewAuthCtrl(service *service.AuthService) *AuthCtrl {
	return &AuthCtrl{service: service}
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
func (a *AuthCtrl) Auth(c *gin.Context) {
	var body struct {
		InitData string `json:"init_data"`
	}
	c.BindJSON(&body)

	data, err := a.service.VerifyTG(body.InitData)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID := data["user_id"]
	accessToken, err := a.service.GenerateAccessToken(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := a.service.GenerateRefreshToken()
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
func (a *AuthCtrl) Refresh(c *gin.Context) {
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

	accessToken, err := a.service.GenerateAccessToken(userID)
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
