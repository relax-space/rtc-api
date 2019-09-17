package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func UserClaimMiddleware(skipPaths ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {
			for _, t := range skipPaths {
				if strings.HasPrefix(c.Path(), t) {
					return next(c)
				}
			}
			token := c.Request().Header.Get("Authorization")
			tokenErr := echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			if token == "" {
				return tokenErr
			}

			userClaim, err := UserClaim{}.FromToken(token)
			if err != nil {
				return err
			}

			req := c.Request()
			c.SetRequest(req.WithContext(context.WithValue(req.Context(), userClaimContextName, userClaim)))

			return next(c)
		}
	}
}

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
