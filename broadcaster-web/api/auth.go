package api

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/shared/models"
)

// POST /auth
func (a *Api) Login(c echo.Context) error {
	name := c.FormValue("username")
	pass := c.FormValue("password")

	uq := models.UserQuery{
		DB: a.DB,
	}
	u, err := uq.GetUserByName(name)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if u.MatchHashPass(pass) {
		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["name"] = u.Username
		claims["roles"] = u.GetRoleNames()
		claims["exp"] = time.Now().Add(a.AuthTimeout).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(a.Cfg.AuthSecret))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{
			"token": t,
		})
	}

	return echo.ErrUnauthorized
}

// GET /api/authvalid
func (a *Api) AuthValid(c echo.Context) error {
	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"message": "Auth Token is valid.",
		},
	})
}
