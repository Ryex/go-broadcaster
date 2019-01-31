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




// User is the model that holds user information
type User struct {
	Id        int64
	Username  string `sql:",unique"`
	Password  string
	Roles     []Role `pg:"many2many:user_to_roles"`
	CreatedAt time.Time `sql:"default:now()"`
}

// ManyToMany join table for users and roles
type UserToRole struct {
	UserId int
	RoleId int
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

// UpdatePassword takes a unhashed password and updates the user model with
// a new hashed version
func (u *User) UpdatePassword(pass string) (err error) {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		logutils.Log.Error("password hashing error", err)
		return
	}

	u.Password = string(hashpass)

	return
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
		"id", "username", "roles", "created_at").Relation("Roles").SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetUserById returns a user from the database by Id
func (uq *UserQuery) GetUserById(id int64) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.id = ?", id).Relation("Roles").Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// GetUserByName returns a user from the database by name
func (uq *UserQuery) GetUserByName(name string) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("user.username = ?", name).Relation("Roles").Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

// DeleteUserById removes a user from the database by Id
func (uq *UserQuery) DeleteUserById(id int64) (err error) {
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

	roles, err := rq.GetRolesByName(roleStrs)
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

// Update uses a used model to update the corasponding user in the databases
func (uq *UserQuery) Update(user *User) (u *User, err error) {
	u = user
	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}

	return
}

// UpdateUserById updates a user's information by Id
func (uq *UserQuery) UpdateUserById(id int64, name string, pass string, roleStrs []string) (u *User, err error) {
	u, err = uq.GetUserById(id)
	if err != nil {
		return
	}

	rq := RoleQuery{
		DB: uq.DB,
	}

	roles, err := rq.GetRolesByName(roleStrs)
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

	roles, err := rq.GetRolesByName(roleStrs)
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

// UserByIdAddRoleByName adds a role in the databases found by name
// to a user in the database found by Id
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

// UserByIdRemoveRoleByName removes a role in the databases found by name
// to a user in the database found by Id
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

// UserByIdAddRoleById adds a role in the databases found by Id
// to a user in the database found by Id
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

// UserByNameAddRoleById adds a role in the databases found by Id
// to a user in the database found by name
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

// UserByIdRemoveRoleById removes a role in the databases found by Id
// to a user in the database found by Id
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

// UserByNameRemoveRoleById removes a role in the databases found by Id
// to a user in the database found by name
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
