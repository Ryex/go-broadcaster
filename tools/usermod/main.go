package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-pg/pg"

	"github.com/ryex/go-broadcaster/internal/config"
	"github.com/ryex/go-broadcaster/internal/logutils"
)

const usageText = `This program modifies users in the database.
commands available are:
  - add - adds a user to the database
  - remove - removes a user from the database
  - modify - modifies a user in the database
Usage:
  go run *.go [-config path/to/config.json] [args] <command> [command args]
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

	switch cmd {
	case "add":
		// TODO
	case "remove":
		// TODO
	case "modify":
		// TODO
	default:
		logutils.Log.Error("Unsupported command: %q", cmd)
		return
	}

	outFile.Sync()
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
