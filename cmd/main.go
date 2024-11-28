package main

import (
	"car-sales-system/internal/db"
	"car-sales-system/internal/gui"
	"log"
)

func main() {
	database, err := db.InitializeDatabase()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer database.Close()

	gui.StartMainGUI(database)

}
