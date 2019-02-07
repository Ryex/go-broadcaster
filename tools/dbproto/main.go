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
	"github.com/ryex/go-broadcaster/internal/models"
)

const usageText = `This program prototypes the database schema using the models package
recording the queries used. Commands available are:
  - proto - prototype the database
  - drop - drop the tables in the database prototype
  - list - list models used
Usage:
  go run *.go [-config path/to/config.json] [args] <command>
Arguments:
`

func main() {
	// Setup command flag proce3ssing
	flag.Usage = usage

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")
	dbnamePtr := flag.String("dbname", "", "Optional alternate database name to connect to")
	outFileNamePtr := flag.String("output", "schema.sql", "output file to record queries to")
	debugPtr := flag.Bool("debug", false, "output debug info level log messages?")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Error: Must provide at least one command")
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

	switch cmd {
	case "proto":
		fmt.Println("Prototyping Database Model Schema:")
		// create schema but error if tables exist
		err = models.CreateSchema(db, false)
		if err != nil {
			logutils.Log.Error("Error creating Schema: %s", err)
		}
	case "drop":
		fmt.Println("Droping Database Model Schema:")
		// drop tables if they exist but don't cascade
		err = models.DropSchema(db, true, false)
		if err != nil {
			logutils.Log.Error("Error droping Schema: %s", err)
		}
	case "list":
		fmt.Println("Models used:")
		names := models.ModelNames()
		for _, name := range names {
			fmt.Printf("- %s\n", name)
		}
	default:
		err = fmt.Errorf("Unsupported command: %q", cmd)
		if err != nil {
			return
		}
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
