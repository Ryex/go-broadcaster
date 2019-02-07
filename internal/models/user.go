package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// User is the model that holds user information
type User struct {
	ID        int64
	Username  string `sql:",unique"`
	Password  string
	Roles     []Role    `pg:"many2many:user_to_roles"`
	CreatedAt time.Time `sql:"default:now()"`
}

// UserToRole is a ManyToMany join table for users and roles
type UserToRole struct {
	UserID int64
	User   *User
	RoleID int64
	Role   *Role
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
func (u *User) AddRole(r *Role) {
	for _, role := range u.Roles {
		if r.IDStr == role.IDStr {
			return
		}
	}
	u.Roles = append(u.Roles, *r)
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(r *Role) {

	delPos := -1
	for i, role := range u.Roles {
		if r.IDStr == role.IDStr {
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

// AfterInsert runs after the model is insterted and ensures all the
// role relations are in the database and synced
func (u *User) AfterInsert(db orm.DB) error {
	return u._ensureRoles(db)
}

// AfterInsert runs after the model is updated and ensures all the
// role relations are in the database and synced
func (u *User) AfterUpdate(db orm.DB) error {
	return u._ensureRoles(db)
}

// AfterDelete runs after the model is updated and ensures all the
// role relations in the database are deleted too
func (u *User) AfterDelete(db orm.DB) error {
	_, err := db.Model((*UserToRole)(nil)).
		Where("user_id = ?", u.ID).
		Delete()
	if err != nil {
		logutils.Log.Errorf(
			"error removing old role relations for deleted user '%d:%s': %s",
			err,
			u.ID,
			u.Username)
		return err
	}
	return nil
}

func (u *User) _getRoleIds() []int64 {
	ids := make([]int64, len(u.Roles))
	for i, role := range u.Roles {
		ids[i] = role.ID
	}
	return ids
}

func (u *User) _ensureRoles(db orm.DB) error {

	// pull relation roles form the database
	var userToRoleMaps []UserToRole
	err := db.Model(&userToRoleMaps).Where("user_id = ?", u.ID).Select()
	if err != nil {
		logutils.Log.Errorf("error pulling role relations: %s", err)
		return err
	}

	// pull the ids out of the map
	dbRoleIds := make([]interface{}, len(userToRoleMaps))
	for i, m := range userToRoleMaps {
		dbRoleIds[i] = m.RoleID
	}

	roleIds := make([]interface{}, len(u.Roles))
	for i, v := range u._getRoleIds() {
		roleIds[i] = v
	}

	// find the ids that need to be removed
	toDelete := make([]interface{}, 0)

	for _, id := range dbRoleIds {
		if !utils.GenericDEContaines(roleIds, id) {
			toDelete = append(toDelete, id)
		}
	}

	// find the ids that need to be added
	toInsert := make([]interface{}, 0)

	for _, id := range roleIds {
		if !utils.GenericDEContaines(dbRoleIds, id) {
			toInsert = append(toInsert, id)
		}
	}

	if err = u._deleteRoleRelations(db, toDelete...); err != nil {
		return err
	}

	if err = u._insertRoleRelations(db, toInsert...); err != nil {
		return err
	}

	return nil
}

func (u *User) _deleteRoleRelations(db orm.DB, toDelete ...interface{}) error {
	if len(toDelete) > 0 {
		_, err := db.Model((*UserToRole)(nil)).
			Where("user_id = ?", u.ID).
			WhereIn("role_id IN (?)", toDelete...).
			Delete()
		if err != nil {
			logutils.Log.Errorf("error removing old role relations: %s", err)
			return err
		}
	}
	return nil
}

func (u *User) _insertRoleRelations(db orm.DB, toInsert ...interface{}) error {
	if len(toInsert) > 0 {
		toInsertMaps := make([]UserToRole, 0)
		for _, id := range toInsert {
			toInsertMaps = append(toInsertMaps, UserToRole{
				UserID: u.ID,
				RoleID: id.(int64),
			})
		}
		_, err := db.Model(&toInsertMaps).Insert()
		if err != nil {
			logutils.Log.Errorf("error inserting new role relations: %s", err)
			return err
		}
	}
	return nil
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
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetUsersLimitByName returns users ordering by name
func (uq *UserQuery) GetUsersLimitByName(order string, limit int, offset int) (users []User, count int, err error) {

	if !utils.StringInSlice(strings.ToUpper(order), []string{"ASC", "DESC"}) {
		err = fmt.Errorf("order must be one of ASC | DESC")
		return
	}

	q := uq.DB.Model(&users)
	q.Order(fmt.Sprintf("user.username %s", order))
	q.Limit(limit)
	q.Offset(offset)
	count, err = q.SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetUsersLimitByID returns users ordering by ID
func (uq *UserQuery) GetUsersLimitByID(order string, limit int, offset int) (users []User, count int, err error) {

	if !utils.StringInSlice(strings.ToUpper(order), []string{"ASC", "DESC"}) {
		err = fmt.Errorf("order must be one of ASC | DESC")
		return
	}

	q := uq.DB.Model(&users)
	q.Order(fmt.Sprintf("user.id %s", order))
	q.Limit(limit)
	q.Offset(offset)
	count, err = q.SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetUserByID returns a user from the database by ID
func (uq *UserQuery) GetUserByID(id int64) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("id = ?", id).Relation("Roles").Select()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// GetUserByName returns a user from the database by name
func (uq *UserQuery) GetUserByName(name string) (u *User, err error) {
	u = new(User)
	err = uq.DB.Model(u).Where("username = ?", name).Relation("Roles").Select()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// DeleteUserByID removes a user from the database by ID
func (uq *UserQuery) DeleteUserByID(id int64) (err error) {
	u := new(User)
	_, err = uq.DB.Model(u).Where("id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error: %s", err)
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

	u = new(User)
	u.Username = name
	err = u.UpdatePassword(pass)
	if err != nil {
		logutils.Log.Error("password hashing error: %s", err)
		return
	}
	u.Roles = roles

	err = uq.DB.Insert(u)
	if err != nil {
		logutils.Log.Error("db query error: %s", err)
	}
	return
}

// Update uses a used model to update the corasponding user in the databases
func (uq *UserQuery) Update(user *User) (u *User, err error) {
	u = user
	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error: %s", err)
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

	roles, err := rq.GetRolesByName(roleStrs)
	if err != nil {
		return
	}

	// hash the password so we dont store it plaintext
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		logutils.Log.Error("password hashing error: %s", err)
		return
	}

	u.Update(name, string(hashpass), roles)

	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error: %s", err)
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
		logutils.Log.Error("password hashing error: %s", err)
		return
	}

	u.Update(name, string(hashpass), roles)

	uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.AddRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.AddRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.RemoveRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.RemoveRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.AddRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.AddRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.RemoveRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
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
	u.RemoveRole(r)
	err = uq.DB.Update(u)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}
