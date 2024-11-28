package gui

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StartAdminGUI запускает интерфейс администратора
func StartAdminGUI(database *sql.DB, app fyne.App) {
	adminWindow := app.NewWindow("Администратор: Главная")
	adminWindow.Resize(fyne.NewSize(600, 400))

	// Кнопки функционала администратора
	addCarButton := widget.NewButton("Добавить автомобиль", func() {
		// Реализация добавления автомобиля
	})
	confirmCheckButton := widget.NewButton("Подтвердить чек", func() {
		// Реализация подтверждения чеков
	})
	deleteClientButton := widget.NewButton("Удалить пользователя", func() {
		// Реализация удаления клиента
	})
	analyzeButton := widget.NewButton("Анализ продаж", func() {
		// Реализация анализа продаж
	})

	// Размещение кнопок
	adminWindow.SetContent(container.NewVBox(
		widget.NewLabel("Добро пожаловать, Администратор!"),
		addCarButton,
		confirmCheckButton,
		deleteClientButton,
		analyzeButton,
	))

	adminWindow.Show()
}
