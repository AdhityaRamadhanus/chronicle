package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/AdhityaRamadhanus/chronicle/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/chronicle/server/render"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

func parseAuthorizationHeader(authHeader, scheme string) (cred string, err error) {
	splittedHeader := strings.Split(authHeader, " ")
	if len(splittedHeader) != 2 {
		return "", errors.New("Cannot parse authorization header")
	}
	parsedScheme := splittedHeader[0]
	if scheme != parsedScheme {
		return "", errors.New("Unexpected Scheme, expected " + scheme)
	}
	return splittedHeader[1], nil
}

func Authenticate(nextHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		authHeader, ok := req.Header["Authorization"]
		if !ok || len(authHeader) == 0 {
			render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
				"status": http.StatusUnauthorized,
				"error": map[string]interface{}{
					"code":    "ErrInvalidAuthorizationHeader",
					"message": "Authorization Header is not present",
				},
			})
			return
		}

		cred, err := parseAuthorizationHeader(authHeader[0], "Bearer")
		if err != nil {
			render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
				"status": http.StatusUnauthorized,
				"error": map[string]interface{}{
					"code":    "ErrInvalidAuthorizationHeader",
					"message": err.Error(),
				},
			})
			return
		}

		token, err := jwt.Parse(cred, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}
			return []byte(viper.GetString("jwt_secret")), nil
		})
		if err != nil {
			render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
				"status": http.StatusUnauthorized,
				"error": map[string]interface{}{
					"code":    "ErrInvalidAccessToken",
					"message": err.Error(),
				},
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		req = req.WithContext(context.WithValue(req.Context(), contextkey.ClientID, claims["client"].(string)))
		nextHandler(res, req)
	})
}
