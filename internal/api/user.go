package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
)

// GET /api/users
func (a *Api) GetUsers(c echo.Context) error {
	q := models.UserQuery{
		DB: a.DB,
	}

	users, count, err := q.GetUsers(c.QueryParams())
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"users": users,
			"count": count,
		},
	})
}

// GET /api/user/id/:id
func (a *Api) GetUserByID(c echo.Context) error {
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

	u, err := q.GetUserByID(id)

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

	q := models.UserQuery{
		DB: a.DB,
	}

	u, err := q.CreateUser(name, pass, roleStrs)

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

	err = q.DeleteUserByID(id)

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

// PUT /api/user/id/:id
func (a *Api) UpdateUserByID(c echo.Context) error {
	name := c.FormValue("username")
	pass := c.FormValue("password")
	roleStrs := strings.Split(c.FormValue("roles"), ",")

	if name == "" {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: errors.New("Missing Username"),
		})
	}

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

	u, err := q.UpdateUserByID(id, name, pass, roleStrs)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"updated": u,
		},
	})

}

// PUT /api/user/name/:name
func (a *Api) UpdateUserByName(c echo.Context) error {
	name := c.FormValue("username")
	pass := c.FormValue("password")
	roleStrs := strings.Split(c.FormValue("roles"), ",")

	if name == "" {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: errors.New("Missing Username"),
		})
	}

	q := models.UserQuery{
		DB: a.DB,
	}

	u, err := q.UpdateUserByName(name, pass, roleStrs)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"updated": u,
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
func (a *Api) GetRoleByID(c echo.Context) error {
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

	r, err := rq.GetRoleByID(id)
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
		parents, err = rq.GetRolesByName(parentStrs)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Responce{
				Err: err,
			})
		}
	}

	r, err := rq.CreateRole(name, perms, parents)
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
func (a *Api) DeleteRoleByID(c echo.Context) error {
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

	err = rq.DeleteRoleByID(id)
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
func (a *Api) UpdateRoleByID(c echo.Context) error {

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

	name := c.FormValue("name")
	perms := strings.Split(c.FormValue("permissions"), ",")
	parentStrs := strings.Split(c.FormValue("parents"), ",")

	var parents []models.Role

	if len(parentStrs) > 0 {
		parents, err = rq.GetRolesByName(parentStrs)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Responce{
				Err: err,
			})
		}
	}

	r, err := rq.UpdateRoleByID(id, name, perms, parents)

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
