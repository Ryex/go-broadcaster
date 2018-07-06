package models

import (
	"errors"
	"time"

	"github.com/go-pg/pg"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"golang.org/x/crypto/bcrypt"
)

type Role struct {
	Id          int64
	IdStr       string `sql:",unique"`
	parents     []Role
	permissions map[string]bool
}

type Permissions map[string]bool

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

func (r *Role) Assign(p string) error {
	if p != "" {
		r.permissions[p] = true
		return nil
	}
	return errors.New("empty permission")
}

func (r *Role) Revoke(p string) error {
	if p != "" {
		r.permissions[p] = false
		return nil
	}
	return errors.New("empty permission")
}

//checks this role or it's parents has a permission
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

//checks if THIS role (parents ignored) has the permission revoked
func (r *Role) Deny(p string) bool {
	if v, ok := r.permissions[p]; ok {
		return !v
	}
	return false
}

type RoleQuery struct {
	DB *pg.DB
}

func (rq *RoleQuery) GetRoleByName(name string) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id_str = ?", name).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (rq *RoleQuery) GetRoleById(id int64) (r *Role, err error) {
	r = new(Role)
	err = rq.DB.Model(r).Where("role.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (rq *RoleQuery) GetRoles(names []string) (roles []Role, err error) {
	roles = make([]Role, len(names))
	err = rq.DB.Model(roles).Where("role.id_str in (?)", pg.In(names)).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (rq *RoleQuery) AddRole(name string, perms []string, parents []Role) (r *Role, err error) {
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

func (rq *RoleQuery) UpdateRole(id int64, name string, perms []string, parents []Role) (r *Role, err error) {
	r = NewRole(name, parents...)
	r.Id = id
	for _, perm := range perms {
		r.Assign(perm)
	}
	err = rq.DB.Update(r)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (rq *RoleQuery) DeleteRoleById(id int64) (err error) {
	r := new(Role)
	_, err = rq.DB.Model(r).Where("role.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

type User struct {
	Id        int64
	Username  string `sql:",unique"`
	Password  string
	Roles     []Role
	CreatedAt time.Time `sql:"default:now()"`
}

// returns if ANY role has permission
func (u *User) HasPermit(p string) bool {
	for _, r := range u.Roles {
		if r.Permit(p) {
			return true
		}
	}
	return false
}

func (u *User) AllPermit(p string) bool {
	for _, r := range u.Roles {
		if !r.Permit(p) {
			return false
		}
	}
	return true
}

func (u *User) AnyDeny(p string) bool {
	for _, r := range u.Roles {
		if r.Deny(p) {
			return true
		}
	}
	return false
}

func (u *User) MatchHashPass(pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass))
	if err != nil {
		return false
	}
	return true
}

type UserQuery struct {
	DB *pg.DB
}

func (uq *UserQuery) GetUserById(id int64) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (uq *UserQuery) GetUserByName(name string) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.username = ?", name).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (uq *UserQuery) DeleteUserById(id int64) (err error) {
	u := new(User)
	_, err = uq.DB.Model(u).Where("user.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (uq *UserQuery) AddUser(name string, pass string, roles []Role) (u *User, err error) {
	if name == "" {
		err = errors.New("empty username")
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
