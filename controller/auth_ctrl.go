package controller

import (
	// "log"
	"encoding/json"

	"rofl-bot/domain"
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
// @Success 200 {object} domain.AuthResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /alena-rofl/auth [post]
func (a *AuthCtrl) Auth(c *gin.Context) {
	var body struct {
		InitData string `json:"init_data"`
	}
	c.BindJSON(&body)

	data, err := a.service.VerifyTG(body.InitData)
	if err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	userJson := data["user"]
	var userData map[string]interface{}
	if err := json.Unmarshal([]byte(userJson), &userData); err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
	}

	username := userData["username"].(string)
	userID := userData["id"].(float64)
	accessToken, err := a.service.GenerateAccessToken(uint(userID), username)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	refreshToken, err := a.service.GenerateRefreshToken(uint(userID), username)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400,
	})
}

// @Summary Обновление токена
// @Description После истечения access токена использовать это. принимает refresh токен, отдает новый access.
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body object{refresh_token=string} true "Refresh Token"
// @Success 200 {object} domain.AuthResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/refresh [post]
func (a *AuthCtrl) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.BindJSON(&body)

	userID := c.GetUint("user_id")
	username := c.GetString("username")
	refresh := body.RefreshToken

	res, err := a.service.CheckRefreshToken(refresh)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}
	if !res {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized"})
		return
	}

	accessToken, err := a.service.GenerateAccessToken(userID, username)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: body.RefreshToken,
		ExpiresIn:    180,
	})

}

// @Summary Удаление токенов
// @Description Удаление всех refresh токенов пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user_id body object{user_id=uint} true "User ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/clear [delete]
func (a *AuthCtrl) ClearTokens(c *gin.Context) {
	var body struct {
		UserID uint `json:"user_id"`
	}
	c.BindJSON(&body)

	userID := body.UserID

	if err := a.service.ClearTokens(userID); err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(200, domain.SuccessResponse{Message: "Tokens cleared."})
}
