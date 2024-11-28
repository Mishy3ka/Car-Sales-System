package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // Импорт SQLite
)

// InitializeDatabase открывает соединение и создаёт таблицы, если их нет
func InitializeDatabase() (*sql.DB, error) {
	// Подключение к базе данных
	db, err := sql.Open("sqlite3", "./carssale.db")
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Создание таблиц
	createTablesSQL := `
 CREATE TABLE IF NOT EXISTS Client (
  ID_Client INTEGER PRIMARY KEY AUTOINCREMENT,
  Name VARCHAR(50),
  LastName VARCHAR(50),
  Phone VARCHAR(15),
  Login VARCHAR(50) UNIQUE,
  Password VARCHAR(255)
 );

 CREATE TABLE IF NOT EXISTS Cars (
  ID_Car INTEGER PRIMARY KEY AUTOINCREMENT,
  Brand VARCHAR(50),
  Model VARCHAR(50),
  YearOfRelease INTEGER,
  Color VARCHAR(30),
  Price DECIMAL(10, 2)
 );

 CREATE TABLE IF NOT EXISTS Administrator (
  ID_Admin INTEGER PRIMARY KEY AUTOINCREMENT,
  Name VARCHAR(50),
  LastName VARCHAR(50),
  Login VARCHAR(50) UNIQUE,
  Password VARCHAR(255),
  Phone VARCHAR(15)
 );

 CREATE TABLE IF NOT EXISTS Checks (
  ID_Check INTEGER PRIMARY KEY AUTOINCREMENT,
  ID_Client INTEGER NOT NULL,
  ID_Car INTEGER NOT NULL,
  ID_Admin INTEGER NOT NULL,
  Price DECIMAL(10, 2),
  FOREIGN KEY (ID_Client) REFERENCES Client(ID_Client),
  FOREIGN KEY (ID_Car) REFERENCES Cars(ID_Car),
  FOREIGN KEY (ID_Admin) REFERENCES Administrator(ID_Admin)
 );
 `

	_, err = db.Exec(createTablesSQL)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблиц: %w", err)
	}

	log.Println("База данных успешно инициализирована.")
	return db, nil
}
