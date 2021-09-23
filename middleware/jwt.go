package middleware

import (
	"errors"
	"strings"
	"time"

	LibJwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func (m *GoMiddleware) SetPayload(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := SetPayload(c, m.jwtSecret); err != nil {
			var code int
			var message interface{}
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				message = he.Message
			}
			return echo.NewHTTPError(code, message)
		}

		return next(c)
	}
}

func SetPayload(c echo.Context, secret string) error {
	var header = c.Request().Header

	authorization := header.Get("Authorization")
	if authorization == "" {
		return echo.NewHTTPError(401, map[string]interface{}{"error": "Unauthorize"})
	}

	token, err := GetTokenFromBarear(authorization, secret)
	if err != nil {
		return echo.NewHTTPError(401, map[string]interface{}{"error": err.Error()})
	}

	c.Set("payload", Get("data", token))

	return nil
}

func GetTokenFromBarear(authorization string, secret string) (*LibJwt.Token, error) {
	if authorization != "" {
		bearerToken := strings.Split(authorization, " ")
		if len(bearerToken) == 2 {
			var token *LibJwt.Token
			var err error
			if token, err = ParseBearerToken(bearerToken[1], secret); err != nil {
				return nil, err
			}
			return token, nil
		}
		return nil, errors.New("Token Mismatch")
	}
	return nil, errors.New("Token Not found")
}

func ParseBearerToken(tokenString string, secret string) (*LibJwt.Token, error) {
	var tokenParse interface{}
	var err error
	tokenParse, err = LibJwt.Parse(tokenString, func(token *LibJwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*LibJwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method: " + token.Header["alg"].(string))
		}

		// verify exp
		registerCliams := Get("claims", token)
		if registerCliams["exp"].(float64) <= float64(time.Now().Unix()) {
			return nil, errors.New("Access Token is Expired")
		}

		return []byte(secret), nil
	})

	return tokenParse.(*LibJwt.Token), err
}

func Get(key string, tokenParse *LibJwt.Token) map[string]interface{} {
	if key == "data" || key == "claims" {
		cliams := tokenParse.Claims.(LibJwt.MapClaims)
		return cliams[key].(map[string]interface{})
	}
	return nil
}
