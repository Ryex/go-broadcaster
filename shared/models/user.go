package models

import (
	"errors"
	"time"
	"net/url"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
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

func (r *Role) Name() string {
	return r.IdStr
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

func (r *Role) Update(name string, perms []string, parents []Role) {
	r.IdStr = name
	r.permissions = make(Permissions)
	r.parents = parents
	for _, perm := range perms {
		r.Assign(perm)
	}
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

func (rq *RoleQuery) UpdateRoleById(id int64, name string, perms []string, parents []Role) (r *Role, err error) {
	r, err = rq.GetRoleById(id)
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

func (u *User) AddRole(r Role) {
	for _, role := range u.Roles {
		if r.IdStr == role.IdStr {
			return
		}
	}
	u.Roles = append(u.Roles, r)
}

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

func (u *User) Update(username string, password string, roles []Role) {
	u.Username = username
	u.Password = password
	u.Roles = roles
}

func (u *User) GetRoleNames() (names []string) {
	names = make([]string, len(u.Roles))
	for i, r := range u.Roles {
		names[i] = r.Name()
	}
	return
}

type UserQuery struct {
	DB *pg.DB
}

func (uq *UserQuery) GetUsers(queryValues url.Values) (users []User, count int, err error) {
	q := uq.DB.Model(&users)
	count, err = q.Apply(orm.Pagination(queryValues)).Column(
			"id", "username", "roles", "created_at").SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
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

func (uq *UserQuery) AddUser(name string, pass string, roleStrs []string) (u *User, err error) {
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

func (uq *UserQuery) UpdateUserById(id int64, name string, pass string, roleStrs []string) (u *User, err error) {
	u, err = uq.GetUserById(id)
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

func (uq *UserQuery) UserByIdAddRoleByName(id int64, rName string) (u *User, err error) {
	u, err = uq.GetUserById(id)
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

func (uq *UserQuery) UserByIdRemoveRoleByName(id int64, rName string) (u *User, err error) {
	u, err = uq.GetUserById(id)
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

func (uq *UserQuery) UserByIdAddRoleById(id int64, rid int64) (u *User, err error) {
	u, err = uq.GetUserById(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleById(rid)
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

func (uq *UserQuery) UserByNameAddRoleById(name string, rid int64) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleById(rid)
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

func (uq *UserQuery) UserByIdRemoveRoleById(id int64, rid int64) (u *User, err error) {
	u, err = uq.GetUserById(id)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleById(rid)
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

func (uq *UserQuery) UserByNameRemoveRoleById(name string, rid int64) (u *User, err error) {
	u, err = uq.GetUserByName(name)
	if err != nil {
		return
	}
	rq := RoleQuery{
		DB: uq.DB,
	}
	r, err := rq.GetRoleById(rid)
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
