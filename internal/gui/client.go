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
		rows, _ := database.Query("SELECT Brand, Model, YearOfRelease, Price FROM Cars")

		defer rows.Close()

		var cars []string
		for rows.Next() {
			var brand, model string
			var year int
			var price float64
			if err := rows.Scan(&brand, &model, &year, &price); err == nil {
				carDetails := fmt.Sprintf("%s %s - %d, Цена: %.2f", brand, model, year, price)
				cars = append(cars, carDetails)
			}
		}

		carList := widget.NewList(
			func() int { return len(cars) },
			func() fyne.CanvasObject { return widget.NewLabel("Купить") },
			func(i widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(cars[i])
			},
		)

		popup := app.NewWindow("Список автомобилей")
		popup.SetContent(container.NewMax(carList))
		popup.Resize(fyne.NewSize(400, 300))
		popup.Show()
	})

	purchaseHistoryButton := widget.NewButton("История покупок", func() {

		// Реализация истории покупок

		rows, err := database.Query(`
        SELECT c.Brand, c.Model, c.YearOfRelease, chk.Price
        FROM Checks chk
        JOIN Cars c ON chk.ID_Car = c.ID_Car
        WHERE chk.ID_Client = ?
    `, currentClientID) // Используем ID текущего клиента
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
			var year int
			var price float64
			if err := rows.Scan(&brand, &model, &year, &price); err == nil {
				purchase := fmt.Sprintf("%s %s (%d), Цена: %.2f", brand, model, year, price)
				purchases = append(purchases, purchase)
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
	addFundsButton := widget.NewButton("Пополнить баланс", func() {
		// Реализация пополнения баланса

	})

	// Размещение кнопок
	clientWindow.SetContent(container.NewVBox(
		widget.NewLabel("Добро пожаловать, Клиент!"),
		browseCarsButton,
		purchaseHistoryButton,
		addFundsButton,
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
