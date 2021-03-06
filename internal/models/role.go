package models

import (
	"errors"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
)

// Permissions is a simple type of strings mapped to bools.
type Permissions map[string]bool

// Permit returns if the perm is granted
func (p Permissions) Permit(perm string) bool {
	if v, ok := p[perm]; ok {
		return v
	}
	return false
}

// Deny returns if the perm is denyied
func (p Permissions) Deny(perm string) bool {
	if v, ok := p[perm]; ok {
		return !v
	}
	return false
}

// Role is a struct for holding user rolle Permissions.
type Role struct {
	ID       int64
	IDStr    string `sql:",unique"`
	ParentID int64
	Parent   *Role
	Perms    Permissions
}

// NewRole consturcts a new Role.
// Usage: role := NewRole("rolename", ParentRole1, ParentRole2, ...)
func NewRole(id string, parent *Role) *Role {
	var pid int64
	if parent != nil {
		pid = parent.ID
	}
	role := &Role{
		IDStr:    id,
		Perms:    make(Permissions),
		ParentID: pid,
		Parent:   parent,
	}
	return role
}

// Name returns the Role's name.
func (r *Role) Name() string {
	return r.IDStr
}

// Assign grants a permission to the role.
func (r *Role) Assign(p string) error {
	if p != "" {
		r.Perms[p] = true
		return nil
	}
	return errors.New("empty permission")
}

// Remove removes a permission from a role.
func (r *Role) Remove(p string) error {
	if p != "" {
		_, ok := r.Perms[p]
		if ok {
			delete(r.Perms, p)
		} else {
			return errors.New("permission not assigned")
		}
		return nil
	}
	return errors.New("empty permission")
}

// Revoke revokes permission from a role.
func (r *Role) Revoke(p string) error {
	if p != "" {
		r.Perms[p] = false
		return nil
	}
	return errors.New("empty permission")
}

// Permit checks if this role or it's parents has a permission
// if a permission is granted in a parents but revoked in the child
// returns false
func (r *Role) Permit(p string) bool {
	//check is this role has the permission
	if r.Perms.Permit(p) {
		return true
	}
	//check if any of the parent roles has the permission
	if r.Parent != nil {
		if r.Parent.Permit(p) {
			return true
		}
	}
	return false
}

// Deny checks if THIS role (parents ignored) has the permission revoked
func (r *Role) Deny(p string) bool {
	return r.Perms.Deny(p)
}

func (r *Role) SetParent(prole *Role) {
	r.ParentID = prole.ID
	r.Parent = prole
}

// Update updates all the information for a role
func (r *Role) Update(name string, perms []string, parent *Role) {
	r.IDStr = name
	r.Perms = make(Permissions)
	r.Parent = parent
	if parent != nil {
		r.ParentID = parent.ID
	} else {
		r.ParentID = 0
	}

	for _, perm := range perms {
		r.Assign(perm)
	}
}

func (r *Role) AfterQuery(db orm.DB) error {
	if r.ParentID != 0 && r.Parent == nil {
		var role Role
		err := db.Model(&role).Where("role.id = ?", r.ParentID).Select()
		if err != nil {
			logutils.Log.Error("db query error %s", err)
			return err
		}
		r.Parent = &role
	}
	return nil
}

// RoleQuery handles Role model queries on the database
type RoleQuery struct {
	DB *pg.DB
}

// GetRoleByName returns a role form the database by name
func (rq *RoleQuery) GetRoleByName(name string) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id_str = ?", name).Select()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetRoleByID returns a Role from the database by ID
func (rq *RoleQuery) GetRoleByID(id int64) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetRoles returns users from the Database
// support pagination
func (rq *RoleQuery) GetRoles(queryValues urlvalues.Values) (roles []Role, count int, err error) {
	//var pagervalues urlvalues.Values
	//err = urlvalues.Decode(queryValues, pagervalues)
	q := rq.DB.Model(&roles)
	count, err = q.Apply(urlvalues.Pagination(queryValues)).SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetRolesByName returns a number of roles from the database by their names,
func (rq *RoleQuery) GetRolesByName(names []string) (roles []Role, err error) {
	roles = make([]Role, len(names))
	if len(roles) > 0 {
		err = rq.DB.Model(&roles).Where("role.id_str in (?)", pg.In(names)).Select()
		if err != nil {
			logutils.Log.Error("db query error: %s", err)
		}
	}
	return
}

// CreateRole adds a role to the database.
func (rq *RoleQuery) CreateRole(name string, perms []string, parent *Role) (r *Role, err error) {
	r = NewRole(name, parent)
	for _, perm := range perms {
		r.Assign(perm)
	}
	err = rq.DB.Insert(r)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// Update uses the model to update the corasponding roel in the databases
func (rq *RoleQuery) Update(role *Role) (r *Role, err error) {
	r = role
	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// UpdateRoleByID updates a role's information in the database by it's ID
func (rq *RoleQuery) UpdateRoleByID(id int64, name string, perms []string, parent *Role) (r *Role, err error) {
	r, err = rq.GetRoleByID(id)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}

	if name == "" {
		name = r.IDStr
	}

	r.Update(name, perms, parent)

	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// DeleteRoleByID removes a role from the database by it's ID.
func (rq *RoleQuery) DeleteRoleByID(id int64) (err error) {
	r := new(Role)
	_, err = rq.DB.Model(r).Where("role.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}
