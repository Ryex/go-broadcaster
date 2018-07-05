package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Role struct {
	Id          string
	parents     map[string]Role
	permissions map[string]bool
}

type Permissions map[string]bool
type Parents map[string]Role

func NewRole(id string, parents ...Role) *Role {
	role := &Role{
		Id:          id,
		parents:     make(Parents),
		permissions: make(Permissions),
	}
	for _, parent := range parents {
		role.parents[parent.Id] = parent
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
	for k, parent := range r.parents {
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

type User struct {
	Username string
	Password string
	Roles    []Role
}

// returns if ANY role has permission
func (u *User) HasPermit(p string) bool {
	for _, r := range u.Roles {
		if r.Permit(p) {
			return true
		}
	}
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
