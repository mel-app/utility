/*
MEL app utility.

This provides a executable for performing various simple actions which would
need extra authentication.

Author:		Alastair Hughes
Contact:	<hobbitalastair at yandex dot com>
*/

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mel-app/backend/src"
)

// usage prints the usage string for the app.
func usage() {
	fmt.Printf(
`%s [bless <user>] | [password <user> <pass>] | [transfer <project> <user>] | [list <user>] | [serve [<port>]]

bless - mark the given user as a manager
password - reset the password for the given user
transfer - transfer the project from the current manager to the given user
list - list the project ids for the given user
serve - run the server on localhost:8080

The environmental variables DATABASE_TYPE and DATABASE_URL are passed through to the sql module to open the database
`, os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	// For the database, default to sqlite and a local database if no
	// DATABASE_TYPE or DATABASE_URL are set.
	dbtype := os.Getenv("DATABASE_TYPE")
	dbname := os.Getenv("DATABASE_URL")
	if dbtype == "" && dbname != "" {
		dbtype = "postgres"
	} else if dbtype == "" && dbname == "" {
		dbtype = "sqlite3"
		dbname = "test.db"
	} else if dbtype != "" && dbname == "" {
		fmt.Printf("DATABASE_TYPE is set, but not DATABASE_URL!")
		usage()
		return
	}

	// Open the database. Note that we don't actually need this for serve.
	db, err := sql.Open(dbtype, dbname)
	if err != nil {
		fmt.Printf("Error opening DB: %q\n", err)
		return
	}
	defer db.Close()

	// Run the command.
	if os.Args[1] == "bless" && len(os.Args) == 3 {
		bless(os.Args[2], db)
	} else if os.Args[1] == "password" && len(os.Args) == 4 {
		password(os.Args[2], os.Args[3], db)
	} else if os.Args[1] == "transfer" && len(os.Args) == 4 {
		transfer(os.Args[2], os.Args[3], db)
	} else if os.Args[1] == "list" && len(os.Args) == 3 {
		list(os.Args[2], db)
	} else if os.Args[1] == "serve" {
		port := "8080"
		if len(os.Args) == 3 {
			port = os.Args[2]
		} else if len(os.Args) != 2 {
			usage()
			return
		}
		fmt.Printf("Starting server on :%s...\n", port)
		backend.Run(port, dbtype, dbname)
	} else {
		usage()
	}
}

// bless marks a user as a manager
func bless(user string, db *sql.DB) {
	_, err := db.Exec("UPDATE users SET is_manager=? WHERE id=?", true, user)
	if err != nil {
		fmt.Printf("Error blessing user: %q\n", err)
	}
}

// password resets the given user's password
func password(user, password string, db *sql.DB) {
	backend.SetPassword(user, password, db)
}

// transfer sets a project's owner to "user"
func transfer(spid string, user string, db *sql.DB) {
	pid, err := strconv.Atoi(spid)
	if err != nil || pid < 0 {
		fmt.Printf("Invalid pid %s\n", spid)
	}
	_, err = db.Exec("UPDATE owns SET name=? WHERE pid=?", user, uint(pid))
	if err != nil {
		fmt.Printf("Failed to update the owner: %q\n", err)
	}
}

// list the user's projects with the corresponding PID
func list(user string, db *sql.DB) {
	rows, err := db.Query("SELECT pid FROM owns WHERE name=?", user)
	if err != nil {
		fmt.Printf("Error getting rows: %q\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		id := -1
		err = rows.Scan(&id)
		if err != nil {
			fmt.Printf("Error getting value: %q\n", err)
			return
		}
		name := ""
		err = db.QueryRow("SELECT name FROM projects WHERE id=?", id).
			Scan(&name)
		if err != nil {
			fmt.Printf("Failed to get the name for project %d: %q\n", id, err)
			return
		}
		fmt.Printf("%d: %s\n", id, name)
	}
	if rows.Err() != nil {
		fmt.Printf("Error getting more rows: %q\n", rows.Err())
	}
}

// vim: sw=4 ts=4 noexpandtab
