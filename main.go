// The MIT License (MIT)
//
// Copyright (c) 2015 Ian Bishop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// sqljgo converts an sqlite input file to json
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
	var tables Tables = make(Tables)
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

		// ------------------------------------------------------------------
		// Credit to http://play.golang.org/p/jxza3pbqq9 for the following
		//
		// Make a slice for the values
		values := make([]interface{}, len(columns))

		// rows.Scan wants '[]interface{}' as an argument, so we must copy the
		// references into such a slice
		// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		// ------------------------------------------------------------------

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
