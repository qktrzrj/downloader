package main

import (
	"database/sql"
	"github.com/andlabs/ui"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	//url := "https://download.jetbrains.8686c.com/idea/ideaIC-2019.2.2.dmg"
	//db, _ = sql.Open("sqlite3", "./downloader.db")
	ui.Main(SetUI())
}
