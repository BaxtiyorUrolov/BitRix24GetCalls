package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func OpenDatabase(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Bazaga bog'lanishni tekshiramiz
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database is not reachable: %v", err)
	}

	fmt.Println("Database connected successfully!")
	return db, nil
}
