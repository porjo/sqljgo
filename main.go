package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Tables map[string]Table

type Table []Row
type Row map[string]interface{}

func main() {

	var tables Tables = make(Tables)

	dbfile := flag.String("dbfile", "", "Sqlite database filename")
	flag.Parse()

	if *dbfile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", *dbfile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables[name] = make(Table, 0)
	}
	rows.Close()

	for name := range tables {
		rows, err := db.Query("SELECT * FROM " + name)
		if err != nil {
			log.Fatal(err)
		}

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			panic(err.Error())
		}

		// Make a slice for the values
		values := make([]interface{}, len(columns))

		// rows.Scan wants '[]interface{}' as an argument, so we must copy the
		// references into such a slice
		// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Fetch rows
		for rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error())
			}

			row := make(Row)
			// Print data
			for i, value := range values {
				if v, ok := value.([]byte); ok {
					value = string(v)
				}

				row[columns[i]] = value
			}
			tables[name] = append(tables[name], row)
		}
		rows.Close()
	}

	j, err := json.MarshalIndent(tables, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", j)
}
