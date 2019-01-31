package models

import (
	"errors"

	"github.com/go-pg/pg"
	//"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
)

// Role is a struct for holding user rolle Permissions.
type Role struct {
	Id       int64
	IdStr    string `sql:",unique"`
	ParentId int64
	Parent   *Role
	Perms    map[string]bool
}

// Permissions is a simple type of strings mapped to bools.
type Permissions map[string]bool

// NewRole consturcts a new Role.
// Usage: role := NewRole("rolename", ParentRole1, ParentRole2, ...)
func NewRole(id string, parent *Role) *Role {
	var pid int64
	if parent != nil {
		pid = parent.Id
	}
	role := &Role{
		IdStr:    id,
		Perms:    make(Permissions),
		ParentId: pid,
		Parent:   parent,
	}
	return role
}

// Name returns the Role's name.
func (r *Role) Name() string {
	return r.IdStr
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
	if v, ok := r.Perms[p]; ok {
		return v
	}
	//check if any of the parent roles has the permission
	if r.Parent.Permit(p) {
		return true
	}
	return false
}

// Deny checks if THIS role (parents ignored) has the permission revoked
func (r *Role) Deny(p string) bool {
	if v, ok := r.Perms[p]; ok {
		return !v
	}
	return false
}

// Properly set the Parent Roles
func (r *Role) SetParent(prole *Role) {
	r.ParentId = prole.Id
	r.Parent = prole
}

// Update updates all the information for a role
func (r *Role) Update(name string, perms []string, parent *Role) {
	r.IdStr = name
	r.Perms = make(Permissions)
	r.Parent = parent
	for _, perm := range perms {
		r.Assign(perm)
	}
}

// RoleQuery handles Role model queries on the database
type RoleQuery struct {
	DB *pg.DB
}

// GetRoleByName returns a role form the database by name
func (rq *RoleQuery) GetRoleByName(name string) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id_str = ?", name).Relation("Parent").Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetRoleById returns a Role from the database by Id
func (rq *RoleQuery) GetRoleById(id int64) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id = ?", id).Relation("Parent").Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetRoles returns users from the Database
// support pagination
func (rq *RoleQuery) GetRoles(queryValues urlvalues.Values) (roles []Role, count int, err error) {
	//var pagervalues urlvalues.Values
	//err = urlvalues.Decode(queryValues, pagervalues)
	q := rq.DB.Model(&roles)
	count, err = q.Apply(urlvalues.Pagination(queryValues)).Relation("Parent").SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetRolesByName returns a number of roles from the database by their names,
func (rq *RoleQuery) GetRolesByName(names []string) (roles []Role, err error) {
	roles = make([]Role, len(names))
	if len(roles) > 0 {
		err = rq.DB.Model(roles).Where("role.id_str in (?)", pg.In(names)).Relation("Parent").Select()
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
		logutils.Log.Error("db query error", err)
	}
	return
}

// Update uses the model to update the corasponding roel in the databases
func (rq *RoleQuery) Update(role *Role) (r *Role, err error) {
	r = role
	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UpdateRoleById updates a role's information in the database by it's Id
func (rq *RoleQuery) UpdateRoleById(id int64, name string, perms []string, parent *Role) (r *Role, err error) {
	r, err = rq.GetRoleById(id)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}

	if name == "" {
		name = r.IdStr
	}

	r.Update(name, perms, parent)

	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// DeleteRoleById removes a role from the database by it's Id.
func (rq *RoleQuery) DeleteRoleById(id int64) (err error) {
	r := new(Role)
	_, err = rq.DB.Model(r).Where("role.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}
