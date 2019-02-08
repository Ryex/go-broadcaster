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
	"github.com/ryex/go-broadcaster/internal/migrations"
)

const usageText = `This program runs command on the db. Supported commands are:
  - up - runs all available migrations.
  - up [target] - runs available migrations up to the target one.
  - down - reverts last migration.
  - create <description> - creates a new migration from a template, auto detects version
  - reset - reverts all migrations.
  - version - prints current db version.
  - set_version [version] - sets db version without running migrations.
Usage:
  go run *.go [-config path/to/config.json] <command> [cmdargs]
Arguments:
`

func main() {
	flag.Usage = usage

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")
	dbnamePtr := flag.String("dbname", "", "Optional alternate database name to connect to")
	outFileNamePtr := flag.String("output", "migrate.sql", "output file to record queries to")
	debugPtr := flag.Bool("debug", false, "output debug info level log messages?")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Error: Must provide at least one command")
		usage()
	}

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

	logutils.SetupLogging("migrations", debug, os.Stdout)

	if dbname != "" {
		cfg.DBDatabase = dbname
	}

	logutils.Log.Info(fmt.Sprintf("Using config: %+v", cfg))

	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBURL + ":" + cfg.DPPort,
		Database: cfg.DBDatabase,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	// setup query logging
	if outFileName == "" {
		outFileName = "migrate.sql"
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

	fmt.Printf("Registered Versions: %v\n", migrations.RegisteredVersions())

	oldVersion, newVersion, err := migrations.Run(db, flag.Args()...)
	if err != nil {
		exitf(err.Error())
	}
	if newVersion != oldVersion {
		fmt.Printf("migrated from version %d to %d\n", oldVersion, newVersion)
	} else {
		fmt.Printf("version is %d\n", oldVersion)
	}
}

func usage() {
	fmt.Print(usageText)
	flag.PrintDefaults()
	os.Exit(2)
}

func errorf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

func exitf(s string, args ...interface{}) {
	errorf(s, args...)
	os.Exit(1)
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
