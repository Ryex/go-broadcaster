package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

// GET /api/user/id/:id
func (a *Api) GetUserById(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.UserQuery{
		DB: a.DB,
	}

	u, err := q.GetUserById(id)

	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"user": u,
		},
	})

}

// GET /api/user/name/:name
func (a *Api) GetUserByName(c echo.Context) error {
	name := c.Param("name")

	q := models.UserQuery{
		DB: a.DB,
	}

	u, err := q.GetUserByName(name)

	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"user": u,
		},
	})

}

// POST /api/user
func (a *Api) AddUser(c echo.Context) error {
	name := c.FormValue("username")
	pass := c.FormValue("password")
	roleStrs := strings.Split(c.FormValue("roles"), ",")

	if name == "" {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: errors.New("Missing Username"),
		})
	}

	rq := models.RoleQuery{
		DB: a.DB,
	}

	roles, err := rq.GetRoles(roleStrs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.UserQuery{
		DB: a.DB,
	}

	u, err := q.AddUser(name, pass, roles)

	if err != nil {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusCreated, Responce{
		Data: H{
			"created": u,
		},
	})

}

// DELETE /api/user/:id
func (a *Api) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.UserQuery{
		DB: a.DB,
	}

	err = q.DeleteUserById(id)

	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"deleted": id,
		},
	})
}

// GET /api/role/name/:name
func (a *Api) GetRoleByName(c echo.Context) error {
	name := c.Param("name")

	rq := models.RoleQuery{
		DB: a.DB,
	}

	r, err := rq.GetRoleByName(name)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"role": r,
		},
	})
}

// GET /api/role/id/:id
func (a *Api) GetRoleById(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	rq := models.RoleQuery{
		DB: a.DB,
	}

	r, err := rq.GetRoleById(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"role": r,
		},
	})
}

// POST /api/role
func (a *Api) AddRole(c echo.Context) error {
	name := c.FormValue("name")
	perms := strings.Split(c.FormValue("permissions"), ",")
	parentStrs := strings.Split(c.FormValue("parents"), ",")

	if name == "" {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: errors.New("Missing Name"),
		})
	}

	rq := models.RoleQuery{
		DB: a.DB,
	}

	var parents []models.Role
	var err error

	if len(parentStrs) > 0 {
		parents, err = rq.GetRoles(parentStrs)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Responce{
				Err: err,
			})
		}
	}

	r, err := rq.AddRole(name, perms, parents)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusCreated, Responce{
		Data: H{
			"created": r,
		},
	})

}

// DELETE /api/role/:id
func (a *Api) DeleteRole(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	rq := models.RoleQuery{
		DB: a.DB,
	}

	err = rq.DeleteRoleById(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"deleted": id,
		},
	})

}

// PUT /api/role/:id //can not create
func (a *Api) UpdateRole(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	rq := models.RoleQuery{
		DB: a.DB,
	}

	r, err := rq.GetRoleById(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	name := c.FormValue("name")
	perms := strings.Split(c.FormValue("permissions"), ",")
	parentStrs := strings.Split(c.FormValue("parents"), ",")

	if name == "" {
		name = r.IdStr
	}

	var parents []models.Role

	if len(parentStrs) > 0 {
		parents, err = rq.GetRoles(parentStrs)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Responce{
				Err: err,
			})
		}
	}

	r, err = rq.UpdateRole(r.Id, name, perms, parents)

	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"role": r,
		},
	})

}
