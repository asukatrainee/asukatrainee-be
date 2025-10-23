package controller

import (
	"context"
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	userDto "github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

type (
	OAuthController interface {
		GoogleLogin(ctx *gin.Context)
		GoogleCallback(ctx *gin.Context)
	}

	oauthController struct {
		oauthService service.OAuthService
	}
)

func NewOAuthController(oauthService service.OAuthService) OAuthController {
	return &oauthController{oauthService: oauthService}
}

// Redirect user ke halaman login Google
func (c *oauthController) GoogleLogin(ctx *gin.Context) {
	r := ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "provider", "google"))
	gothic.BeginAuthHandler(ctx.Writer, r)
}

func (c *oauthController) GoogleCallback(ctx *gin.Context) {
	user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		res := utils.BuildResponseFailed("Failed to authenticate via Google", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	googleData := userDto.GoogleUserData{
		Email:  user.Email,
		Name:   user.Name,
		Avatar: user.AvatarURL,
	}

	tokenResp, err := c.oauthService.GoogleLogin(ctx, googleData)
	if err != nil {
		res := utils.BuildResponseFailed("Failed to process Google login", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("Successfully logged in with Google", tokenResp)
	ctx.JSON(http.StatusOK, res)
}
