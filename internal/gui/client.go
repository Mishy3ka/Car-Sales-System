package gui

import (
	"database/sql"
	"fmt"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var nameValidationRegex = regexp.MustCompile(`^[а-яА-Яa-zA-Z]+$`)
var phoneValidationRegex = regexp.MustCompile(`^[0-9]+$`)
var currentClientID int

func createValidatedEntry(placeHolder string, parentWindow fyne.Window) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeHolder)

	entry.OnChanged = func(input string) {
		if !nameValidationRegex.MatchString(input) && input != "" {
			entry.SetText(input[:len(input)-1]) // Удаляем последний символ
			dialog.ShowError(
				fmt.Errorf("поле '%s' может содержать только буквы", placeHolder),
				parentWindow,
			)
		}
	}

	return entry
}

func createPhoneValidatedEntry(placeHolder string, parentWindow fyne.Window) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeHolder)

	entry.OnChanged = func(input string) {
		if !phoneValidationRegex.MatchString(input) && input != "" {
			entry.SetText(input[:len(input)-1]) // Удаляем последний символ
			dialog.ShowError(
				fmt.Errorf("поле '%s' может содержать только цифры", placeHolder),
				parentWindow,
			)
		}
	}

	return entry
}

func StartClientGUI(database *sql.DB, app fyne.App) {
	clientWindow := app.NewWindow("Клиент: Главная")
	clientWindow.Resize(fyne.NewSize(600, 400))

	// Кнопки функционала клиента
	browseCarsButton := widget.NewButton("Просмотр автомобилей", func() {
		dialog.ShowInformation("Важная информация", "Для того чтобы купить автомобиль просто нажмите на него", clientWindow)
		rows, err := database.Query("SELECT ID_Car, Brand, Model, YearOfRelease, Price FROM Cars")
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при загрузке списка автомобилей: %v", err), clientWindow)
			return
		}
		defer rows.Close()

		var cars []string
		var carIDs []int
		for rows.Next() {
			var id int
			var brand, model string
			var year int
			var price float64
			if err := rows.Scan(&id, &brand, &model, &year, &price); err == nil {
				carDetails := fmt.Sprintf("%s %s - %d, Цена: %.2f", brand, model, year, price)
				cars = append(cars, carDetails)
				carIDs = append(carIDs, id)
			}
		}

		if len(cars) == 0 {
			dialog.ShowInformation("Список пуст", "Автомобили отсутствуют.", clientWindow)
			return
		}

		// Создаем виджеты для каждого автомобиля
		var carWidgets []fyne.CanvasObject
		for i, car := range cars {
			index := i // Создаём копию переменной, чтобы избежать проблем с замыканием
			carButton := widget.NewButton(fmt.Sprintf("Купить: %s", car), func() {
				carID := carIDs[index]
				var price float64
				err := database.QueryRow("SELECT Price FROM Cars WHERE ID_Car = ?", carID).Scan(&price)
				if err != nil {
					dialog.ShowError(fmt.Errorf("ошибка при получении цены: %v", err), clientWindow)
					return
				}

				// Вставка данных в таблицу Checks
				_, err = database.Exec(
					"INSERT INTO Checks (ID_Client, ID_Car, ID_Admin, Price) VALUES (?, ?, NULL, ?)",
					currentClientID, carID, price,
				)
				if err != nil {
					dialog.ShowError(fmt.Errorf("ошибка при добавлении чека: %v", err), clientWindow)
					return
				}

				// Сообщение об успешной покупке
				dialog.ShowInformation("Успешная покупка", "Автомобиль успешно куплен!", clientWindow)
			})
			carWidgets = append(carWidgets, carButton)
		}

		// Создаем контейнер для списка автомобилей
		carList := container.NewVBox(carWidgets...)

		// Открываем всплывающее окно с прокручиваемым списком автомобилей
		popup := app.NewWindow("Список автомобилей")
		popup.SetContent(container.NewVScroll(carList))
		popup.Resize(fyne.NewSize(400, 300))
		popup.Show()
	})

	purchaseHistoryButton := widget.NewButton("История покупок", func() {
		rows, err := database.Query(`
			SELECT c.Brand, c.Model, c.YearOfRelease, chk.Price
			FROM Checks chk
			LEFT JOIN Cars c ON chk.ID_Car = c.ID_Car
			WHERE chk.ID_Client = ?
			  AND (c.IsArchived = FALSE OR c.IsArchived = TRUE OR c.ID_Car IS NULL)
		`, currentClientID)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Ошибка",
				Content: fmt.Sprintf("Ошибка получения данных: %v", err),
			})
			return
		}
		defer rows.Close()

		var purchases []string
		for rows.Next() {
			var brand, model string
			var year sql.NullInt32
			var price float64
			if err := rows.Scan(&brand, &model, &year, &price); err == nil {
				if year.Valid {
					purchase := fmt.Sprintf("%s %s (%d), Цена: %.2f", brand, model, year.Int32, price)
					purchases = append(purchases, purchase)
				} else {
					purchase := fmt.Sprintf("%s %s (удалено из базы), Цена: %.2f", brand, model, price)
					purchases = append(purchases, purchase)
				}
			}
		}

		if len(purchases) == 0 {
			purchases = append(purchases, "История покупок пуста.")
		}

		purchaseList := widget.NewList(
			func() int { return len(purchases) },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(purchases[i])
			},
		)

		popup := app.NewWindow("История покупок")
		popup.SetContent(container.NewMax(purchaseList))
		popup.Resize(fyne.NewSize(400, 300))
		popup.Show()
	})

	// Размещение кнопок
	clientWindow.SetContent(container.NewVBox(
		widget.NewLabel("Добро пожаловать, Клиент!"),
		browseCarsButton,
		purchaseHistoryButton,
	))

	clientWindow.Show()
}

func openClientLogin(database *sql.DB, app fyne.App) {
	loginWindow := app.NewWindow("Клиент: Вход")
	loginWindow.Resize(fyne.NewSize(400, 300))

	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Логин")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	loginButton := widget.NewButton("Войти", func() {
		login := loginEntry.Text
		password := passwordEntry.Text

		var id int
		err := database.QueryRow("SELECT ID_Client FROM Client WHERE Login = ? AND Password = ?", login, password).Scan(&id)
		if err != nil {
			dialog.ShowError(fmt.Errorf("неверный логин или пароль"), loginWindow)
			return
		}

		currentClientID = id

		dialog.ShowInformation("Успешный вход", "Добро пожаловать!", loginWindow)
		StartClientGUI(database, app) // Запуск GUI клиента
		loginWindow.Close()
	})

	registerButton := widget.NewButton("Зарегистрироваться", func() {
		openClientRegister(database, app)
		loginWindow.Close()
	})

	loginWindow.SetContent(container.NewVBox(
		widget.NewLabel("Введите данные для входа:"),
		loginEntry,
		passwordEntry,
		loginButton,
		registerButton,
	))

	loginWindow.Show()
}

// openClientRegister открывает окно для регистрации клиента
func openClientRegister(database *sql.DB, app fyne.App) {
	registerWindow := app.NewWindow("Клиент: Регистрация")
	registerWindow.Resize(fyne.NewSize(400, 400))

	nameEntry := createValidatedEntry("Имя", registerWindow)
	lastNameEntry := createValidatedEntry("Фамилия", registerWindow)
	phoneEntry := createPhoneValidatedEntry("Телефон", registerWindow) // Используем проверку телефона
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Логин")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	registerButton := widget.NewButton("Зарегистрироваться", func() {
		name := nameEntry.Text
		lastName := lastNameEntry.Text
		phone := phoneEntry.Text
		login := loginEntry.Text
		password := passwordEntry.Text

		if name == "" || lastName == "" || phone == "" || login == "" || password == "" {
			dialog.ShowError(fmt.Errorf("все поля должны быть заполнены"), registerWindow)
			return
		}

		_, err := database.Exec(
			"INSERT INTO Client (Name, LastName, Phone, Login, Password) VALUES (?, ?, ?, ?, ?)",
			name, lastName, phone, login, password,
		)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при регистрации"), registerWindow)
			return
		}

		dialog.ShowInformation("Регистрация успешна", "Теперь вы можете войти", registerWindow)
		registerWindow.Close()
	})

	registerWindow.SetContent(container.NewVBox(
		widget.NewLabel("Заполните данные для регистрации:"),
		nameEntry,
		lastNameEntry,
		phoneEntry,
		loginEntry,
		passwordEntry,
		registerButton,
	))

	registerWindow.Show()
}
