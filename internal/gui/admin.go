package gui

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var AnameValidationRegex = regexp.MustCompile(`^[а-яА-Яa-zA-Z]+$`)
var AphoneValidationRegex = regexp.MustCompile(`^[0-9]+$`)

// StartAdminGUI запускает интерфейс администратора
func StartAdminGUI(database *sql.DB, app fyne.App) {
	adminWindow := app.NewWindow("Администратор: Главная")
	adminWindow.Resize(fyne.NewSize(600, 400))

	// Кнопки функционала администратора
	addCarButton := widget.NewButton("Добавить автомобиль", func() {
		// Реализация добавления автомобиля
		addCarWindow := app.NewWindow("Добавить автомобиль")
		addCarWindow.Resize(fyne.NewSize(400, 400))

		brandEntry := CreateValidatedEntry("Марка", addCarWindow, `^[^\d]+$`, "Марка не должна содержать цифры")
		modelEntry := widget.NewEntry()
		modelEntry.SetPlaceHolder("Модель")
		yearEntry := CreateValidatedEntry("Год выпуска", addCarWindow, `^\d+$`, "Год выпуска должен содержать только цифры")
		colorEntry := CreateValidatedEntry("Цвет", addCarWindow, `^[^\d]+$`, "Цвет не должен содержать цифры")
		priceEntry := CreateValidatedEntry("Цена", addCarWindow, `^\d+$`, "Цена должна содержать только цифры")

		saveButton := widget.NewButton("Сохранить", func() {
			brand := brandEntry.Text
			model := modelEntry.Text
			year := yearEntry.Text
			color := colorEntry.Text
			price := priceEntry.Text

			if brand == "" || model == "" || year == "" || color == "" || price == "" {
				dialog.ShowError(fmt.Errorf("все поля должны быть заполнены"), addCarWindow)
				return
			}

			_, err := database.Exec(
				"INSERT INTO Cars (Brand, Model, YearOfRelease, Color, Price) VALUES (?, ?, ?, ?, ?)",
				brand, model, year, color, price,
			)
			if err != nil {
				dialog.ShowError(fmt.Errorf("ошибка добавления автомобиля: %v", err), addCarWindow)
				return
			}

			dialog.ShowInformation("Успех", "Автомобиль успешно добавлен", addCarWindow)

		})

		cancelButton := widget.NewButton("Отмена", func() {
			addCarWindow.Close()
		})

		addCarWindow.SetContent(container.NewVBox(
			widget.NewLabel("Введите данные для нового автомобиля:"),
			brandEntry,
			modelEntry,
			yearEntry,
			colorEntry,
			priceEntry,
			container.NewHBox(saveButton, cancelButton),
		))

		addCarWindow.Show()
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

func openAdminLogin(database *sql.DB, app fyne.App) {
	loginWindow := app.NewWindow("Админ: вход")
	loginWindow.Resize(fyne.NewSize(300, 300))

	// Объявляем переменные для ввода
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Логин")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	loginButton := widget.NewButton("Войти", func() {
		login := loginEntry.Text
		password := passwordEntry.Text

		var id int
		err := database.QueryRow("SELECT ID_Admin FROM Administrator WHERE Login = ? AND Password = ?", login, password).Scan(&id)
		if err != nil {
			dialog.ShowError(fmt.Errorf("неверный логин или пароль"), loginWindow)
			return
		}

		dialog.ShowInformation("Успешный вход", "Добро пожаловать!", loginWindow)
		StartAdminGUI(database, app) // Запуск GUI админа
		loginWindow.Close()
	})

	// Размещение элементов в окне
	loginWindow.SetContent(container.NewVBox(
		widget.NewLabel("Введите логин и пароль администратора:"),
		loginEntry,
		passwordEntry,
		loginButton,
	))

	loginWindow.Show()
}

func CreateValidatedEntry(placeHolder string, parentWindow fyne.Window, pattern string, errorMessage string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeHolder)

	regex := regexp.MustCompile(pattern)
	entry.OnChanged = func(input string) {
		if !regex.MatchString(input) && len(input) > 0 {
			// Удаляем последний символ
			entry.SetText(input[:len(input)-1])

			// Устанавливаем курсор в конец текста после удаления
			entry.CursorColumn = len(entry.Text)

			dialog.ShowError(
				errors.New(errorMessage),
				parentWindow,
			)
		}
	}

	return entry
}
