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
	"github.com/mel-app/backend/src"
)

// usage prints the usage string for the app.
func usage() {
	fmt.Printf(
`%s <command> [<args> ...]

bless <user> - mark the given user as a manager
curse <user> - mark the given user as a client (undo a bless)
password <user> <pass> - reset the password for the given user
transfer <project> <user> - transfer the project from the current manager to the given user
users - list the names of all users
projects [<user>] - list project names and ids, for all users if none is given
delete [user <user>] [project <project>] - delete the given user or project
serve [<port>] - run the server on localhost, defaulting to port 8080
init - initialise a new database

The environmental variable DATABASE_URL is passed through to the sql module to open the database.
`, os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	dbname := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbname)
	if err != nil {
		fmt.Printf("Error opening DB: %q\n", err)
		return
	}
	defer db.Close()

	// Run the command.
	if os.Args[1] == "bless" && len(os.Args) == 3 {
		bless(os.Args[2], db)
	} else if os.Args[1] == "curse" && len(os.Args) == 3 {
		curse(os.Args[2], db)
	} else if os.Args[1] == "password" && len(os.Args) == 4 {
		password(os.Args[2], os.Args[3], db)
	} else if os.Args[1] == "transfer" && len(os.Args) == 4 {
		transfer(os.Args[2], os.Args[3], db)
	} else if os.Args[1] == "users" && len(os.Args) == 2 {
		all_users(db)
	} else if os.Args[1] == "projects" && len(os.Args) == 3 {
		user_projects(os.Args[2], db)
	} else if os.Args[1] == "projects" && len(os.Args) == 2 {
		all_projects(db)
	} else if os.Args[1] == "delete" && os.Args[2] == "user" && len(os.Args) == 4 {
		delete_user(os.Args[3], db)
	} else if os.Args[1] == "delete" && os.Args[2] == "project" && len(os.Args) == 4 {
		delete_project(os.Args[3], db)
	} else if os.Args[1] == "init" && len(os.Args) == 2 {
		initDB(db)
	} else if os.Args[1] == "serve" {
		port := "8080"
		if len(os.Args) == 3 {
			port = os.Args[2]
		} else if len(os.Args) != 2 {
			usage()
			return
		}
		backend.Run(port, db)
	} else {
		usage()
	}
}

// bless marks a user as a manager
func bless(user string, db *sql.DB) {
	err := backend.NewDB(db).SetIsManager(user, true)
	if err != nil {
		fmt.Printf("Error blessing user: %q\n", err)
	}
}

// curse marks a user as a client
func curse(user string, db *sql.DB) {
	err := backend.NewDB(db).SetIsManager(user, false)
	if err != nil {
		fmt.Printf("Error cursing user: %q\n", err)
	}
}

// delete_user removes the given user
func delete_user(user string, db *sql.DB) {
	err := backend.NewDB(db).DeleteUser(user)
	if err != nil {
		fmt.Printf("Error deleting user: %q\n", err)
	}
}

// delete_project removes the given project
func delete_project(spid string, db *sql.DB) {
	pid, err := strconv.Atoi(spid)
	if err != nil || pid < 0 {
		fmt.Printf("Invalid pid %s\n", spid)
	}
	_, err = db.Exec("DELETE FROM views WHERE pid=$1", pid)
	if err != nil {
		fmt.Printf("Error deleting project: %q\n", err)
	}
	_, err = db.Exec("DELETE FROM owns WHERE pid=$1", pid)
	if err != nil {
		fmt.Printf("Error deleting project: %q\n", err)
	}
	_, err = db.Exec("DELETE FROM projects WHERE id=$1", pid)
	if err != nil {
		fmt.Printf("Error deleting project: %q\n", err)
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
	_, err = db.Exec("UPDATE owns SET name=$1 WHERE pid=$2", user, uint(pid))
	if err != nil {
		fmt.Printf("Failed to update the owner: %q\n", err)
	}
}

// all_users lists all of the users in the database
func all_users(db *sql.DB) {
	rows, err := db.Query("SELECT name FROM users")
	if err != nil {
		fmt.Printf("Error getting rows: %q\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		name := ""
		err = rows.Scan(&name)
		if err != nil {
			fmt.Printf("Error getting value: %q\n", err)
			return
		}
		fmt.Printf("%q\n", name)
	}
	if rows.Err() != nil {
		fmt.Printf("Error getting more rows: %q\n", rows.Err())
	}
}

// user_projects lists the user's projects with the corresponding PID
func user_projects(user string, db *sql.DB) {
	rows, err := db.Query("SELECT pid FROM owns WHERE name=$1", user)
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
		err = db.QueryRow("SELECT name FROM projects WHERE id=$1", id).
			Scan(&name)
		if err != nil {
			fmt.Printf("Failed to get the name for project %d: %q\n", id, err)
			return
		}
		fmt.Printf("%d: %q\n", id, name)
	}
	if rows.Err() != nil {
		fmt.Printf("Error getting more rows: %q\n", rows.Err())
	}
}

// all_projects lists all of the projects in the database
func all_projects(db *sql.DB) {
	rows, err := db.Query("SELECT id, name FROM projects")
	if err != nil {
		fmt.Printf("Error getting rows: %q\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		id := -1
		name := ""
		err = rows.Scan(&id, &name)
		if err != nil {
			fmt.Printf("Error getting value: %q\n", err)
			return
		}
		fmt.Printf("%d: %q\n", id, name)
	}
	if rows.Err() != nil {
		fmt.Printf("Error getting more rows: %q\n", rows.Err())
	}
}

// initDB initialises the database with the expected tables
func initDB(db *sql.DB) {
	backend.NewDB(db).Init()
}

// vim: sw=4 ts=4 noexpandtab
