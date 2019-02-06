package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-pg/pg"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ryex/go-broadcaster/internal/config"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
)

const usageText = `This program modifies users in the database.
commands available are:
  - add - adds a user to the database
  - remove - removes a user from the database
  - modify - modifies a user in the database
Usage:
  go run *.go [-config path/to/config.json] [args] <command> [command args]
Command Arguments:
	- add <name> <roles>
		- name - the new user's username
		- roles - a comma seperated list of role names | NONE
	- remove <name>
		- name - name of the user to remove
	- modify <name> <action> <action args>
		- name - name of user to modify
		- action - the action to take, one of (chpasswd, addrole, removerole, chname)
			- chpasswd
				- No Arguments
			- addrole <rolename>
				- rolename - name of role to add
			- removerole <rolename>
				- rolename - name of role to remove
			- chname <name>
				- name - the new user name
	- list <method> <order> [limit] [offset]
		- method - string, one of ID | NAME
		- order - string, one of ASC | DESC
		- limit - optional, limit for the number of object returned
		- offset - offset to start listing at
Arguments:
`

func main() {
	// Setup command flag proce3ssing
	flag.Usage = usage

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")
	dbnamePtr := flag.String("dbname", "", "Optional alternate database name to connect to")
	outFileNamePtr := flag.String("output", "usermod.sql", "output file to record queries to")
	debugPtr := flag.Bool("debug", false, "output debug info level log messages?")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Error: Must provide username")
		usage()
	}

	// Load config file for database connection
	cfgPath = *cfgPtr
	dbname := *dbnamePtr
	outFileName := *outFileNamePtr
	debug := *debugPtr

	cfgPath, pathErr := filepath.Abs(cfgPath)
	if pathErr != nil {
		fmt.Println("could not get absolute path for config", pathErr)
	}

	fmt.Println("Loading config from: ", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Println("Error when loading configuration", err)
	}
	// setup logging
	logutils.SetupLogging("dbprototyping", debug, os.Stdout)

	if dbname != "" {
		cfg.DBDatabase = dbname
	}

	logutils.Log.Info(fmt.Sprintf("Using config: %+v", cfg))

	// connect to the database
	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBURL + ":" + cfg.DPPort,
		Database: cfg.DBDatabase,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	// setup query logging
	if outFileName == "" {
		outFileName = "usermod.sql"
	}
	outFilePath := filepath.Join(root, outFileName)
	outFilePath, pathErr = filepath.Abs(outFilePath)
	if pathErr != nil {
		fmt.Println("could not get absolute path for output file", pathErr)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	setupDatabaseQueryLogging(db, outFile)
	logutils.Log.Info("Writing output to %s", outFilePath)

	// process commands
	a := flag.Args()
	cmd := ""
	if len(a) > 0 {
		cmd = a[0]
	}
	args := a[1:]

	var cmderr error

	switch cmd {
	case "add":
		if checkLen(args, 2, cmd) {
			return
		}
		cmderr = addUser(db, args[0], args[1])
		if cmderr != nil {
			logutils.Log.Error("Error adding user: %s", cmderr)
			return
		}
	case "remove":
		if checkLen(args, 1, cmd) {
			return
		}
		cmderr = removeUser(db, args[0])
		if cmderr != nil {
			logutils.Log.Error("Error removing user: %s", cmderr)
			return
		}
	case "modify":
		if !(len(args) == 2 || len(args) == 3) {
			logutils.Log.Error("Command '%s' expects 2 or 3 arguments, got %d", cmd, len(args))
			return
		}
		cmderr = modifyUser(db, args[0], args[1:])
		if cmderr != nil {
			logutils.Log.Error("Error modifing user: %s", cmderr)
			return
		}
	case "list":
		if !(len(args) == 2 || len(args) == 4) {
			logutils.Log.Error("Command '%s' expects 2 to 4 arguments, got %d", cmd, len(args))
			return
		}
		var limit int
		if len(args) >= 3 {
			var lierr error
			limit, lierr = strconv.Atoi(args[2])
			if lierr != nil {
				logutils.Log.Error("Error parseing limit: %s", lierr)
			}
		} else {
			limit = 100
		}
		var offset int
		if len(args) == 4 {
			var offerr error
			offset, offerr = strconv.Atoi(args[3])
			if offerr != nil {
				logutils.Log.Error("Error parseing offset: %s", offerr)
			}
		} else {
			offset = 0
		}
		cmderr = listUsers(db, args[0], args[1], limit, offset)
		if cmderr != nil {
			logutils.Log.Error("Error listing users: %s", cmderr)
			return
		}
	default:
		logutils.Log.Error("Unsupported command: %q", cmd)
		return
	}

	outFile.Sync()
}

// addUser takes a name and either
// a comma seperated list of role names or 'NONE'
// then adds a user to the database
func addUser(db *pg.DB, name string, rolesStr string) (err error) {
	uq := models.UserQuery{
		DB: db,
	}

	u, err := uq.GetUserByName(name)
	fmt.Printf("Info: User: '%v' Error: '%s'\n", u, err)
	if err == nil {
		err = fmt.Errorf("User '%s' already exists", name)
		return
	}

	// clear the error from no user?
	err = nil

	var roleNames []string
	if rolesStr != "NONE" {
		roleNames = strings.Split(rolesStr, ",")
	}

	pass, err := getPassword()
	if err != nil {
		return
	}

	user, err := uq.CreateUser(name, pass, roleNames)
	if err != nil {
		return
	}

	fmt.Printf("Created user %+v\n", user)

	return
}

// removeUser takes a name and removes the user from the database
func removeUser(db *pg.DB, name string) (err error) {
	uq := models.UserQuery{
		DB: db,
	}

	u, err := uq.GetUserByName(name)
	if u == nil {
		err = fmt.Errorf("User '%s' does not exist", name)
		return
	}
	// clear the error from no user?
	err = nil

	err = uq.DeleteUserById(u.Id)
	if err != nil {
		return
	}

	fmt.Printf("Deleated user %+v\n", u)

	return

}

func modifyUser(db *pg.DB, name string, a []string) (err error) {
	uq := models.UserQuery{
		DB: db,
	}

	u, err := uq.GetUserByName(name)
	if u == nil {
		err = fmt.Errorf("User '%s' does not exist", name)
		return
	}
	// clear the error from no user?
	err = nil

	cmd := ""
	if len(a) > 0 {
		cmd = a[0]
	}
	args := a[1:]

	var cmderr error

	switch cmd {
	case "chpasswd":
		if checkLen(args, 0, cmd) {
			return
		}
		var pass string
		pass, cmderr = getPassword()
		if cmderr != nil {
			return cmderr
		}
		u.UpdatePassword(pass)
		_, cmderr = uq.Update(u)
		if cmderr != nil {
			return cmderr
		}
		fmt.Printf("Updated password for user '%s'\n", u.Username)
	case "addrole":
		if checkLen(args, 1, cmd) {
			return
		}
		cmderr = addRole(&uq, u, args[0])
		if cmderr != nil {
			return cmderr
		}
	case "removerole":
		if checkLen(args, 1, cmd) {
			return
		}
		cmderr = removeRole(&uq, u, args[0])
		if cmderr != nil {
			return cmderr
		}
	case "chname":
		if checkLen(args, 1, cmd) {
			return
		}
		cmderr = changeName(&uq, u, args[0])
		if cmderr != nil {
			return cmderr
		}
	default:
		logutils.Log.Error("Unsupported command: %q", cmd)
		return
	}

	return

}

func addRole(uq *models.UserQuery, u *models.User, roleName string) (err error) {

	rq := models.RoleQuery{
		DB: uq.DB,
	}

	r, dberr := rq.GetRoleByName(roleName)
	if dberr != nil {
		return dberr
	}

	u.AddRole(r)

	_, err = uq.Update(u)
	if err != nil {
		return
	}

	fmt.Printf("Added role '%s' to user '%s'", roleName, u.Username)
	return
}

func removeRole(uq *models.UserQuery, u *models.User, roleName string) (err error) {

	rq := models.RoleQuery{
		DB: uq.DB,
	}

	r, dberr := rq.GetRoleByName(roleName)
	if dberr != nil {
		return dberr
	}

	u.RemoveRole(r)

	_, err = uq.Update(u)
	if err != nil {
		return
	}

	fmt.Printf("Removed role '%s' to user '%s'", roleName, u.Username)
	return

}

func changeName(uq *models.UserQuery, u *models.User, name string) (err error) {
	oldName := u.Username
	u.Username = name
	_, err = uq.Update(u)
	if err != nil {
		return
	}

	fmt.Printf("Changed username of user Id:%d form '%s' to '%s'", u.Id, oldName, name)
	return
}

func listUsers(db *pg.DB, method string, order string, limit int, offset int) (err error) {
	uq := models.UserQuery{
		DB: db,
	}

	var users []models.User
	var count int
	switch strings.ToUpper(method) {
	case "ID":
		users, count, err = uq.GetUsersLimitById(order, limit, offset)
	case "NAME":
		users, count, err = uq.GetUsersLimitByName(order, limit, offset)
	default:
		err = fmt.Errorf("method must be one of ID | NAME")
		return
	}

	fmt.Printf("Count: %d \n", count)
	for _, user := range users {
		fmt.Printf("%d: %s \n", user.Id, user.Username)
	}

	return
}

func getPassword() (pass string, err error) {

	fmt.Print("Enter Password:    ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return
	}
	fmt.Println("\nPassword typed: " + string(bytePassword))

	fmt.Print("Re-Enter Password: ")
	bytePassword2, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return
	}
	fmt.Println("\nPassword typed: " + string(bytePassword2))

	if string(bytePassword) != string(bytePassword2) {
		err = fmt.Errorf("Passwords don't match")
		return
	}
	pass = string(bytePassword)
	// TODO
	return
}

func checkLen(args []string, l int, cmd string) bool {
	if len(args) != l {
		logutils.Log.Error("Command '%s' expects %d arguments got %d", cmd, l, len(args))
		return true
	}
	return false
}

func usage() {
	fmt.Print(usageText)
	flag.PrintDefaults()
	os.Exit(2)
}

type dbLogger struct {
	out io.Writer
}

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	query, err := q.FormattedQuery()
	if err != nil {
		panic(err)
	}
	out := d.out
	if out == nil {
		out = os.Stdout
	}
	fmt.Fprintf(out, "%s;\n", query)
}

func setupDatabaseQueryLogging(db *pg.DB, out io.Writer) {
	logger := new(dbLogger)
	logger.out = out
	db.AddQueryHook(logger)
}
