package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"idm/inner/common"
)

const (
	JwtKey   = "jwt"
	IdmAdmin = "IDM_ADMIN"
	IdmUser  = "IDM_USER"
)

type IdmClaims struct {
	RealmAccess RealmAccessClaims `json:"realm_access"`
	jwt.RegisteredClaims
}

type RealmAccessClaims struct {
	Roles []string `json:"roles"`
}

var AuthMiddleware = func(logger *common.Logger) fiber.Handler {
	config := Config{
		ContextKey:   JwtKey,
		ErrorHandler: createJwtErrorHandler(logger),
		JWKSetURLs:   []string{"http://localhost:9990/realms/idm/protocol/openid-connect/certs"},
		Claims:       &IdmClaims{},
	}
	return New(config)
}

func createJwtErrorHandler(logger *common.Logger) fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		//logger.Error("failed autentication", zap.Error(err))
		logger.ErrorCtx(ctx.Context(), "failed autentication", zap.Error(err))
		// Если токен не может быть прочитан, то возвращаем 401
		return common.ErrResponse(
			ctx,
			fiber.StatusUnauthorized,
			err.Error(),
		)
	}
}
