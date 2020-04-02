package main_test

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Init initializes the database backed by the file located at path.
func sqlite3db(path string) (*sqlx.DB, error) {
	log.Printf("Sqlite3 database path is %s", path)

	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to the sqlite3 database: %w", err)
	}

	return db, nil
}

func genSqlData(db *sqlx.DB, data []map[string]interface{}) error {
	db.MustExec("DROP TABLE IF EXISTS test;")

	const schema = `
CREATE TABLE IF NOT EXISTS test (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    lastname TEXT NOT NULL,
    age INTEGER NOT NULL,
    salary INTEGER NOT NULL
);
`
	db.MustExec(schema)

	tx := db.MustBegin()
	for i := range data {
		_, err := tx.NamedExec("INSERT INTO test (id, name, lastname, age, salary) VALUES (:id, :name, :lastname, :age, :salary)", data[i])
		if err != nil {
			return err
		}

		if i > 0 && i%1000 == 0 {
			log.Println("sqlite3 progess:", i)
		}
	}
	tx.Commit()

	log.Println("sqlite3 completed successfully...", len(data))

	return nil
}

func sqlQuery(db *sqlx.DB, qry string) ([]map[string]interface{}, error) {
	rows, err := db.Queryx(qry)
	if err != nil {
		return nil, err
	}

	res := []map[string]interface{}{}
	for rows.Next() {
		m := map[string]interface{}{}
		err := rows.MapScan(m)
		if err != nil {
			return nil, err
		}

		res = append(res, m)
	}

	return res, nil
}
