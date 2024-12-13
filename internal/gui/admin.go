package gui

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	addCarButton := widget.NewButton("Добавить автомобиль", func() { // Функция доабвления автомобиля
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

			yearInt, err := strconv.Atoi(year)
			if err != nil || yearInt > 2024 || yearInt < 1970 {
				dialog.ShowError(fmt.Errorf("ошибка: Год выпуска должен быть в диапазоне от 1970 до 2024"), addCarWindow)
				return
			}

			if brand == "" || model == "" || year == "" || color == "" || price == "" {
				dialog.ShowError(fmt.Errorf("все поля должны быть заполнены"), addCarWindow)
				return
			}

			_, err = database.Exec(
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

	deleteClientButton := widget.NewButton("Удалить пользователя", func() {
		openDeleteClientWindow(database, app)
	})

	deleteCarButton := widget.NewButton("Удалить автомобиль", func() { //Удаление автомобиля из базы данных

		rows, err := database.Query(`SELECT ID_Car, Brand, Model FROM Cars WHERE IsArchived = FALSE`)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка получения списка автомобилей: %v", err), adminWindow)
			return
		}
		defer rows.Close()

		var carList []string
		var carIDs []int
		for rows.Next() {
			var id int
			var brand, model string
			if err := rows.Scan(&id, &brand, &model); err == nil {
				carList = append(carList, fmt.Sprintf("%s %s", brand, model))
				carIDs = append(carIDs, id)
			}
		}

		if len(carList) == 0 {
			dialog.ShowInformation("Информация", "нет доступных автомобилей для удаления.", adminWindow)
			return
		}

		carSelect := widget.NewSelect(carList, func(selected string) {
			for i, car := range carList {
				if car == selected {
					_, err := database.Exec(`UPDATE Cars SET IsArchived = TRUE WHERE ID_Car = ?`, carIDs[i])
					if err != nil {
						dialog.ShowError(fmt.Errorf("ошибка удаления автомобиля: %v", err), adminWindow)
					} else {
						dialog.ShowInformation("успех", "автомобиль успешно удален.", adminWindow)
					}
					return
				}
			}
		})

		popup := app.NewWindow("Удаление автомобиля")
		popup.SetContent(container.NewVBox(
			widget.NewLabel("Выберите автомобиль для удаления:"),
			carSelect,
			widget.NewButton("Закрыть", func() { popup.Close() }),
		))
		popup.Resize(fyne.NewSize(400, 200))
		popup.Show()
	})

	analyzeButton := widget.NewButton("Анализ продаж", func() { //Функция для анализа продаж
		// SQL-запрос для анализа продаж
		rows, err := database.Query(`
			SELECT 
				Cars.Brand, 
				Cars.Model, 
				SUM(Checks.Price) AS TotalRevenue, 
				COUNT(Checks.ID_Check) AS TotalSales
			FROM Cars
			JOIN Checks ON Cars.ID_Car = Checks.ID_Car
			WHERE Cars.IsArchived = FALSE
			GROUP BY Cars.ID_Car
			ORDER BY TotalSales DESC
			LIMIT 3;
		`)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка анализа продаж: %v", err), adminWindow)
			return
		}
		defer rows.Close()

		// Формируем список результатов
		var results []string
		for rows.Next() {
			var brand, model string
			var totalRevenue float64
			var totalSales int
			if err := rows.Scan(&brand, &model, &totalRevenue, &totalSales); err == nil {
				results = append(results, fmt.Sprintf("%s %s: продаж: %d ,  доход: %.2f Р", brand, model, totalSales, totalRevenue))
			}
		}

		// Проверяем, есть ли результаты
		if len(results) == 0 {
			dialog.ShowInformation("Результаты анализа", "Продаж пока нет", adminWindow)
			return
		}

		// Отображаем результаты в новом окне
		resultsWindow := app.NewWindow("Результаты анализа")
		resultsWindow.Resize(fyne.NewSize(400, 300))
		resultsWindow.SetContent(container.NewVBox(
			widget.NewLabel("Топ-3 самых продаваемых автомобиля:"),
			widget.NewLabel(strings.Join(results, "\n")),
			widget.NewButton("Закрыть", func() {
				resultsWindow.Close()
			}),
		))
		resultsWindow.Show()
	})

	// Размещение кнопок
	adminWindow.SetContent(container.NewVBox(
		widget.NewLabel("Добро пожаловать, Администратор!"),
		addCarButton,
		deleteCarButton,
		deleteClientButton,
		analyzeButton,
	))

	adminWindow.Show()
}

func openAdminLogin(database *sql.DB, app fyne.App) { //функция входа в систему для админа
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

func CreateValidatedEntry(placeHolder string, parentWindow fyne.Window, pattern string, errorMessage string) *widget.Entry { //Функцция проверки вводимых символов
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

func openDeleteClientWindow(database *sql.DB, app fyne.App) { //Функция удаления пользователя
	deleteClientWindow := app.NewWindow("Удалить пользователя")
	deleteClientWindow.Resize(fyne.NewSize(400, 300))

	// Получение списка пользователей из базы данных
	rows, err := database.Query("SELECT ID_Client, Name, LastName FROM Client")
	if err != nil {
		dialog.ShowError(fmt.Errorf("ошибка получения пользователей: %v", err), deleteClientWindow)
		return
	}
	defer rows.Close()

	var clients []string
	clientMap := make(map[string]int)

	for rows.Next() {
		var id int
		var name, lastName string
		if err := rows.Scan(&id, &name, &lastName); err == nil {
			clientLabel := fmt.Sprintf("%s %s (ID: %d)", name, lastName, id)
			clients = append(clients, clientLabel)
			clientMap[clientLabel] = id
		}
	}

	if len(clients) == 0 {
		dialog.ShowInformation("Информация", "Нет пользователей для удаления", deleteClientWindow)
		return
	}

	clientSelect := widget.NewSelect(clients, func(selected string) {})
	clientSelect.PlaceHolder = "Выберите пользователя"

	deleteButton := widget.NewButton("Удалить", func() {
		selectedClient := clientSelect.Selected
		if selectedClient == "" {
			dialog.ShowError(fmt.Errorf("пользователь не выбран"), deleteClientWindow)
			return
		}

		clientID := clientMap[selectedClient]

		// Удаление пользователя и его чеков из базы данных
		_, err := database.Exec("DELETE FROM Checks WHERE ID_Client = ?", clientID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка удаления чеков пользователя: %v", err), deleteClientWindow)
			return
		}

		_, err = database.Exec("DELETE FROM Client WHERE ID_Client = ?", clientID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка удаления пользователя: %v", err), deleteClientWindow)
			return
		}

		dialog.ShowInformation("Успех", "Пользователь успешно удален", deleteClientWindow)
	})

	cancelButton := widget.NewButton("Отмена", func() {
		deleteClientWindow.Close()
	})

	deleteClientWindow.SetContent(container.NewVBox(
		widget.NewLabel("Выберите пользователя, которого хотите удалить:"),
		clientSelect,
		container.NewHBox(deleteButton, cancelButton),
	))

	deleteClientWindow.Show()
}
