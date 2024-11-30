package gui

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StartMainGUI запускает главное окно выбора роли
func StartMainGUI(database *sql.DB) {
	application := app.New()
	mainWindow := application.NewWindow("Car Sales System")
	mainWindow.Resize(fyne.NewSize(400, 200))

	// Кнопки для выбора роли
	clientButton := widget.NewButton("Войти как Клиент", func() {
		openClientLogin(database, application)
		mainWindow.Close()
	})
	adminButton := widget.NewButton("Войти как Администратор", func() {
		openAdminLogin(database, application)
		mainWindow.Close()
	})

	// Контейнер с кнопками
	mainWindow.SetContent(container.NewVBox(
		widget.NewLabel("Выберите роль:"),
		clientButton,
		adminButton,
	))

	mainWindow.ShowAndRun()
}
