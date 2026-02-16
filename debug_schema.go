package main

import (
	"database/sql"
	"fmt"
	"github.com/nathfavour/autocommiter.go/internal/index"
	_ "modernc.org/sqlite"
)

func main() {
	path, _ := index.GetDBPath()
	db, _ := sql.Open("sqlite", path)
	defer db.Close()

	rows, _ := db.Query("SELECT name FROM pragma_table_info('repo_cache')")
	defer rows.Close()
	fmt.Println("Columns in repo_cache:")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf(" - %s\n", name)
	}
}
