package models

import (
	"errors"
	"net/url"
	"time"

	"github.com/go-pg/pg"
	//"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"golang.org/x/crypto/bcrypt"
)

// Role is a struct for holding user rolle Permissions.
type Role struct {
	Id          int64
	IdStr       string `sql:",unique"`
	parents     []Role
	permissions map[string]bool
}

// Permissions is a simple type of strings mapped to bools.
type Permissions map[string]bool

// NewRole consturcts a new Role.
// Usage: role := NewRole("rolename", ParentRole1, ParentRole2, ...)
func NewRole(id string, parents ...Role) *Role {
	role := &Role{
		IdStr:       id,
		permissions: make(Permissions),
	}
	for _, parent := range parents {
		// can't get a role parented to itself
		if parent.IdStr == id {
			continue
		}
		role.parents = append(role.parents, parent)
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
		r.permissions[p] = true
		return nil
	}
	return errors.New("empty permission")
}

// Revoke removes a permission form a role.
func (r *Role) Revoke(p string) error {
	if p != "" {
		r.permissions[p] = false
		return nil
	}
	return errors.New("empty permission")
}

// Permit checks if this role or it's parents has a permission
// if a permission is granted in a parents but revoked in the child
// returns false
func (r *Role) Permit(p string) bool {
	//check is this role has the permission
	if v, ok := r.permissions[p]; ok {
		return v
	}
	//check if any of the parent roles has the permission
	for _, parent := range r.parents {
		// sanity check to prevent recusion should the wworst happen
		if parent.IdStr == r.IdStr {
			continue
		}
		if parent.Permit(p) {
			return true
		}
	}
	return false
}

// Deny checks if THIS role (parents ignored) has the permission revoked
func (r *Role) Deny(p string) bool {
	if v, ok := r.permissions[p]; ok {
		return !v
	}
	return false
}

// Update updates all the information for a role
func (r *Role) Update(name string, perms []string, parents []Role) {
	r.IdStr = name
	r.permissions = make(Permissions)
	r.parents = parents
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
	err = rq.DB.Model(r).Where("role.id_str = ?", name).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetRoleByID returns a Role from the database by ID
func (rq *RoleQuery) GetRoleByID(id int64) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetRoles returns a number of roles from the database by their names,
func (rq *RoleQuery) GetRoles(names []string) (roles []Role, err error) {
	roles = make([]Role, len(names))
	err = rq.DB.Model(roles).Where("role.id_str in (?)", pg.In(names)).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// CreateRole adds a role to the database.
func (rq *RoleQuery) CreateRole(name string, perms []string, parents []Role) (r *Role, err error) {
	r = NewRole(name, parents...)
	for _, perm := range perms {
		r.Assign(perm)
	}
	err = rq.DB.Insert(r)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UpdateRoleByID updates a role's information in the database by it's ID
func (rq *RoleQuery) UpdateRoleByID(id int64, name string, perms []string, parents []Role) (r *Role, err error) {
	r, err = rq.GetRoleByID(id)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}

	if name == "" {
		name = r.IdStr
	}

	r.Update(name, perms, parents)

	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// DeleteRoleByID removes a role from the database by it's ID.
func (rq *RoleQuery) DeleteRoleByID(id int64) (err error) {
	r := new(Role)
	_, err = rq.DB.Model(r).Where("role.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// User is the model that holds user information
type User struct {
	ID        int64
	Username  string `sql:",unique"`
	Password  string
	Roles     []Role
	CreatedAt time.Time `sql:"default:now()"`
}

// HasPermit returns if ANY role of the user has the given permission
func (u *User) HasPermit(p string) bool {
	for _, r := range u.Roles {
		if r.Permit(p) {
			return true
		}
	}
	return false
}

// AllPermit returns is ALL roles of the user have a given permission
func (u *User) AllPermit(p string) bool {
	for _, r := range u.Roles {
		if !r.Permit(p) {
			return false
		}
	}
	return true
}

// AnyDeny returns is ANY roles of the user deny a permission
func (u *User) AnyDeny(p string) bool {
	for _, r := range u.Roles {
		if r.Deny(p) {
			return true
		}
	}
	return false
}

// MatchHashPass matches a bcrypt hashed password to the
// stored bcrypt hashed password
func (u *User) MatchHashPass(pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass))
	if err != nil {
		return false
	}
	return true
}

// AddRole adds a role to the user
func (u *User) AddRole(r Role) {
	for _, role := range u.Roles {
		if r.IdStr == role.IdStr {
			return
		}
	}
	u.Roles = append(u.Roles, r)
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(r Role) {

	delPos := -1
	for i, role := range u.Roles {
		if r.IdStr == role.IdStr {
			delPos = i
			break
		}
	}
	// if we found a match
	if delPos >= 0 {
		// delete
		u.Roles[delPos] = u.Roles[len(u.Roles)-1]
		u.Roles = u.Roles[:len(u.Roles)-1]
	}

}

// Update updats the user's information
func (u *User) Update(username string, password string, roles []Role) {
	u.Username = username
	u.Password = password
	u.Roles = roles
}

// GetRoleNames returns a list of all the role names the user has
func (u *User) GetRoleNames() (names []string) {
	names = make([]string, len(u.Roles))
	for i, r := range u.Roles {
		names[i] = r.Name()
	}
	return
}

// UserQuery handles model quieries for the User model
type UserQuery struct {
	DB *pg.DB
}

// GetUsers returns users from the Database
// support pagination
func (uq *UserQuery) GetUsers(queryValues url.Values) (users []User, count int, err error) {
	var pagervalues urlvalues.Values
	err = urlvalues.Decode(queryValues, pagervalues)
	q := uq.DB.Model(&users)
	count, err = q.Apply(urlvalues.Pagination(pagervalues)).Column(
		"id", "username", "roles", "created_at").SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetUserByID returns a user from the database by ID
func (uq *UserQuery) GetUserByID(id int64) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetUserByName returns a user from the database by name
func (uq *UserQuery) GetUserByName(name string) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.username = ?", name).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// DeleteUserByID removes a user from the database by ID
func (uq *UserQuery) DeleteUserByID(id int64) (err error) {
	u := new(User)
	_, err = uq.DB.Model(u).Where("user.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// CreateUser add a user to the database
func (uq *UserQuery) CreateUser(name string, pass string, roleStrs []string) (u *User, err error) {
	if name == "" {
		err = errors.New("empty username")
		return
	}

	rq := RoleQuery{
		DB: uq.DB,
	}

	roles, err := rq.GetRoles(roleStrs)
	if err != nil {
		return
	}

	// hash the password so we dont store it plaintext
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		logutils.Log.Error("password hashing error", err)
		return
	}

	u = new(User)
	u.Username = name
	u.Password = string(hashpass)
	u.Roles = roles

	err = uq.DB.Insert(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UpdateUserByID updates a user's information by ID
func (uq *UserQuery) UpdateUserByID(id int64, name string, pass string, roleStrs []string) (u *User, err error) {
	u, err = uq.GetUserByID(id)
	if err != nil {
		return
	}

	rq := RoleQuery{
		DB: uq.DB,
	}

	roles, err := rq.GetRoles(roleStrs)
	if err != nil {
		return
	}

	// hash the password so we dont store it plaintext
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		logutils.Log.Error("password hashing error", err)
		return
	}

	u.Update(name, string(hashpass), roles)

	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}

	return
}

// UpdateUserByName updates a user's information by name
func (uq *UserQuery) UpdateUserByName(name string, pass string, roleStrs []string) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}

	rq := RoleQuery{
		DB: uq.DB,
	}

	roles, err := rq.GetRoles(roleStrs)
	if err != nil {
		return
	}

	// hash the password so we dont store it plaintext
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		logutils.Log.Error("password hashing error", err)
		return
	}

	u.Update(name, string(hashpass), roles)

	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByIDAddRoleByName adds a role in the databases found by name
// to a user in the database found by ID
func (uq *UserQuery) UserByIDAddRoleByName(id int64, rName string) (u *User, err error) {
	u, err = uq.GetUserByID(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByName(rName)
	if err != nil {
		return
	}
	u.AddRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByNameAddRoleByName adds a role in the databases found by name
// to a user in the database found by name
func (uq *UserQuery) UserByNameAddRoleByName(name string, rName string) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByName(rName)
	if err != nil {
		return
	}
	u.AddRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByIDRemoveRoleByName removes a role in the databases found by name
// to a user in the database found by ID
func (uq *UserQuery) UserByIDRemoveRoleByName(id int64, rName string) (u *User, err error) {
	u, err = uq.GetUserByID(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByName(rName)
	if err != nil {
		return
	}
	u.RemoveRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByNameRemoveRoleByName removes a role in the databases found by name
// to a user in the database found by name
func (uq *UserQuery) UserByNameRemoveRoleByName(name string, rName string) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByName(rName)
	if err != nil {
		return
	}
	u.RemoveRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByIDAddRoleByID adds a role in the databases found by ID
// to a user in the database found by ID
func (uq *UserQuery) UserByIDAddRoleByID(id int64, rid int64) (u *User, err error) {
	u, err = uq.GetUserByID(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByID(rid)
	if err != nil {
		return
	}
	u.AddRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByNameAddRoleByID adds a role in the databases found by ID
// to a user in the database found by name
func (uq *UserQuery) UserByNameAddRoleByID(name string, rid int64) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByID(rid)
	if err != nil {
		return
	}
	u.AddRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByIDRemoveRoleByID removes a role in the databases found by ID
// to a user in the database found by ID
func (uq *UserQuery) UserByIDRemoveRoleByID(id int64, rid int64) (u *User, err error) {
	u, err = uq.GetUserByID(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByID(rid)
	if err != nil {
		return
	}
	u.RemoveRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// UserByNameRemoveRoleByID removes a role in the databases found by ID
// to a user in the database found by name
func (uq *UserQuery) UserByNameRemoveRoleByID(name string, rid int64) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleByID(rid)
	if err != nil {
		return
	}
	u.RemoveRole(*r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}
