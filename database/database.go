package database

import (
	"database/sql"

	"github.com/makew0rld/1rg-server/config"
	_ "github.com/mattn/go-sqlite3"
)

func Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", config.Config.DBPath+"?_txlock=immediate")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
	PRAGMA foreign_keys = true;
	PRAGMA busy_timeout = 5000;
	CREATE TABLE IF NOT EXISTS rolodex (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		pronouns TEXT NOT NULL,
		email TEXT NOT NULL,
		bio TEXT NOT NULL,
		birthday TEXT NOT NULL,
		website TEXT NOT NULL,
		bluesky TEXT NOT NULL,
		goodreads TEXT NOT NULL,
		fedi TEXT NOT NULL,
		github TEXT NOT NULL,
		instagram TEXT NOT NULL
	) STRICT;
		`)
	if err != nil {
		return nil, err
	}
	return db, nil
}
