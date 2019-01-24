package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pg/pg"

	"github.com/ryex/go-broadcaster/internal/config"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
)

const usageText = `This program modifies roles in the database.
commands available are:
  - add - adds a role to the database
  - remove - removes a role from the database
  - modify - modifies a role in the database
  - list - list roles in the database
  - describe - display information on a particular role
Usage:
  go run *.go [-config path/to/config.json] [args] <command> [command args]
Command Arguments:
  - add <name> <perms> <parents>
    - name - role name to add
    - perms - comma seperated list of permissions
    - parents - comma seperated list of parent role names | NONE
  - remove <name>
    - name - role name to remove
  - modify <name> <action> <action args>
    - name - role name to modify
    - action - the action to take, one of (allow, deny, addparent, removeparent)
      - assign <perm> - the role will grant that permission
        - perm - the name of the permision to grant
      - remove <perm>
        - perm - the name of the permission to remove
      - revoke <perm> - the role will deny that permission
        - perm - the permission to deny
      - addparent <role> - add the role as a parent of the current role
        - role - the role to ad as parent
      - removeparent <role> - remove the role as a parent of the current role
        - role - the rolle to remove as parent
  - list
  - describe <name>
    - name - role to describe
Arguments:
`

func main() {
	// Setup command flag proce3ssing
	flag.Usage = usage

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")
	dbnamePtr := flag.String("dbname", "", "Optional alternate database name to connect to")
	outFileNamePtr := flag.String("output", "rolemod.sql", "output file to record queries to")
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
		outFileName = "schema.sql"
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

	rq := models.RoleQuery{
		DB: db,
	}

	switch cmd {
	case "add":
		fmt.Println("test add")
		if len(args) != 3 {
			logutils.Log.Error("Command 'add' expects 3 arguments got %d", len(args))
			return
		}
		name := args[0]
		permsStr := args[1]
		parentsStr := args[2]
		fmt.Printf("Adding Role '%s'\n", name)
		perms := strings.Split(permsStr, ",")
		var parentNames []string
		if parentsStr != "NONE" {
			parentNames = strings.Split(parentsStr, ",")
		}
		parents, dberr := rq.GetRolesByName(parentNames)
		if dberr != nil {
			logutils.Log.Error("Error fetching parent roles: %s", dberr)
			return
		}
		role, dberr := rq.CreateRole(name, perms, parents)
		if dberr != nil {
			logutils.Log.Error("Error creating role", dberr)
			return
		}
		fmt.Printf("Created role %+v\n", role)
	case "remove":
		if len(args) != 1 {
			logutils.Log.Error("Command 'remove' expects 1 argument got %d", len(args))
			return
		}
		name := args[0]
		role, dberr := rq.GetRoleByName(name)
		if dberr != nil {
			logutils.Log.Error(fmt.Sprintf("Role '%s' does not exist", name), dberr)
			return
		}
		dberr = rq.DeleteRoleByID(role.ID)
		if dberr != nil {
			logutils.Log.Error("Error deleting role", dberr)
			return
		}
		fmt.Printf("Removed role '%s'\n", name)
	case "modify":
		if len(args) != 3 {
			logutils.Log.Error("Command 'modify' expects 3 arguments got %d", len(args))
			return
		}
		name := args[0]
		role, dberr := rq.GetRoleByName(name)
		if dberr != nil {
			logutils.Log.Error(fmt.Sprintf("Role '%s' does not exist", name), dberr)
			return
		}
		mcmd := args[1]
		err = modifyRole(mcmd, role, &rq, args[2:])
		if err != nil {
			logutils.Log.Error("Could not modify role: %s", err)
			return
		}
		fmt.Printf("Modified role '%s' to %+v\n", name, role)
		return
	case "list":
		if len(args) != 0 {
			logutils.Log.Error("Command 'list' expects 0 arguments got %d", len(args))
			return
		}
		//var urlvalues url.Values
		roles, count, dberr := rq.GetRoles(nil)
		if dberr != nil {
			logutils.Log.Error("error fetching roles", dberr)
			return
		}
		fmt.Printf("Roles: (Total: %d)\n", count)
		for i, role := range roles {
			fmt.Printf("  %d: %s\n", i, role.IDStr)
		}
	case "describe":
		if len(args) != 1 {
			logutils.Log.Error("Command 'describe' expects 1 argument got %d", len(args))
			return
		}
		name := args[0]
		role, dberr := rq.GetRoleByName(name)
		if dberr != nil {
			logutils.Log.Error(fmt.Sprintf("Role '%s' does not exist", name), dberr)
			return
		}
		describeRole(role)
	default:
		logutils.Log.Error("Unsupported command: %q", cmd)
		return
	}

	outFile.Sync()

}

func modifyRole(cmd string, role *models.Role, rq *models.RoleQuery, args []string) (err error) {

	if len(args) != 1 {
		err = fmt.Errorf("action %s expects 1 argument got %d", cmd, len(args))
		return err
	}
	switch cmd {
	case "assign":
		perm := args[0]
		err = role.Assign(perm)
		if err != nil {
			err = fmt.Errorf("error assiging permission '%s': %s", perm, err)
			return err
		}
		_, dberr := rq.Update(role)
		if dberr != nil {
			logutils.Log.Error("Error Updating Role: %s", dberr)
			return dberr
		}
	case "remove":
		perm := args[0]
		err = role.Remove(perm)
		if err != nil {
			err = fmt.Errorf("error removing permission '%s': %s ", perm, err)
			return err
		}
		_, dberr := rq.Update(role)
		if dberr != nil {
			logutils.Log.Error("Error Updating Role: %s", dberr)
			return dberr
		}
	case "revoke":
		perm := args[0]
		err = role.Revoke(perm)
		if err != nil {
			err = fmt.Errorf("error revoking permission'%s': %s ", perm, err)
			return err
		}
		_, dberr := rq.Update(role)
		if dberr != nil {
			logutils.Log.Error("Error Updating Role: %s", dberr)
			return dberr
		}
	case "addparent":
		name := args[0]
		prole, dberr := rq.GetRoleByName(name)
		if dberr != nil {
			logutils.Log.Error(fmt.Sprintf("Role '%s' does not exist", name), dberr)
			return
		}
		role.AddParent(*prole)
		_, dberr = rq.Update(role)
		if dberr != nil {
			logutils.Log.Error("Error Updating Role: %s", dberr)
			return dberr
		}
	case "removeparent":
		name := args[0]
		prole, dberr := rq.GetRoleByName(name)
		if dberr != nil {
			logutils.Log.Error(fmt.Sprintf("Role '%s' does not exist", name), dberr)
			return
		}
		err = role.RemoveParent(*prole)
		if err != nil {
			err = fmt.Errorf("error removing parent '%s': %s ", name, err)
			return err
		}
		_, dberr = rq.Update(role)
		if dberr != nil {
			logutils.Log.Error("Error Updating Role: %s", dberr)
			return dberr
		}
	default:
		err = fmt.Errorf("Unsupported command: %q", cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func describeRole(role *models.Role) {
	fmt.Printf("Role %s:\n", role.IDStr)
	fmt.Println("Permissions:")
	for k, v := range role.Perms {
		status := "granted"
		if !v {
			status = "denied"
		}
		fmt.Printf("- %s -> %s\n", k, status)
	}
	fmt.Println("Parents:")
	for _, parent := range role.Parents {
		fmt.Printf("- %s\n", parent.IDStr)
	}
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
