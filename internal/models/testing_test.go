package models

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	v, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		os.Exit(1)
	}
	var err error
	db, err = sqlx.Connect("postgres", v)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	defer db.Close()

	os.Exit(m.Run())
}
