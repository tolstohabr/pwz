package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"PWZ1.0/internal/models"

	"PWZ1.0/internal/storage"
)

const (
	dateFormat = "2006-01-02"
)

func Run(storage *storage.FileStorage, scanner *bufio.Scanner) {
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("Завершение работы.")
			break
		}

		args := strings.Split(input, " ")
		switch args[0] {
		case "help":
			printHelp()
		case "accept-order":
			handleAcceptOrder(storage, args[1:])
		case "return-order":
			handleReturnOrder(storage, args[1:])
		case "process-order":
			handleProcessOrders(storage, args[1:])
		case "list-orders":
			handleListOrders(storage, args[1:])
		case "list-returns":
			handleListReturns(storage, args[1:])
		case "order-history":
			handleOrderHistory()
		case "import-orders":
			handleImportOrders(storage, args[1:])
		case "scroll-orders":
			handleScrollOrders(storage, args[1:])
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}

func printHelp() {
	fmt.Println("Список команд:")
	fmt.Println("  help")
	fmt.Println("  accept-order     Принять заказ от курьера")
	fmt.Println("  return-order     Вернуть заказ") //удалить
	fmt.Println("  process-order   	Выдать или принять возврат")
	fmt.Println("  list-orders    	Получить список заказов")
	fmt.Println("  list-returns    	Получить список возвратов")
	fmt.Println("  order-history   	Получить историю заказов")
	fmt.Println("  import-orders   	Импорт заказов из файла")
	fmt.Println("  scroll-orders   	Прокрутка")
	fmt.Println("  exit             Выйти из приложения")
}

// handleAcceptOrder Принять заказ от курьера
func handleAcceptOrder(storage *storage.FileStorage, args []string) {
	var orderID, userID, expiresStr string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--order-id":
			if i+1 < len(args) {
				orderID = args[i+1]
				i++
			}
		case "--user-id":
			if i+1 < len(args) {
				userID = args[i+1]
				i++
			}
		case "--expires":
			if i+1 < len(args) {
				expiresStr = args[i+1]
				i++
			}
		}
	}

	switch {
	case orderID == "":
		fmt.Println("ERROR: VALIDATION_FAILED: отсутствует orderID")
		return
	case userID == "":
		fmt.Println("ERROR: VALIDATION_FAILED: отсутствует userID")
		return
	case expiresStr == "":
		fmt.Println("ERROR: VALIDATION_FAILED: отсутствует expiresStr")
		return
	}

	expiresAt, err := time.Parse(dateFormat, expiresStr)
	if err != nil {
		fmt.Println("ERROR: VALIDATION_FAILED: Неверный формат даты")
		return
	}

	err = AcceptOrder(storage, orderID, userID, expiresAt)
	if err != nil {
		fmt.Println("ERROR:", err.Error())
	} else {
		fmt.Println("ORDER_ACCEPTED:", orderID)
	}
}

// handleReturnOrder Вернуть заказ
func handleReturnOrder(storage *storage.FileStorage, args []string) {
	var orderID string

	for i := 0; i < len(args); i++ {
		if args[i] == "--order-id" && i+1 < len(args) {
			orderID = args[i+1]
			i++
		}
	}

	if orderID == "" {
		fmt.Println("ERROR: VALIDATION_FAILED: Параметр --order-id обязателен.")
		return
	}

	err := ReturnOrder(storage, orderID)
	if err != nil {
		fmt.Println("ERROR: INTERNAL_ERROR: ", err.Error())
	} else {
		fmt.Println("ORDER_RETURNED:", orderID)
	}
}

// handleProcessOrders Выдать или принять возврат
func handleProcessOrders(storage storage.Storage, args []string) {
	var userID, action, orderIDsStr string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userID = args[i+1]
				i++
			}
		case "--action":
			if i+1 < len(args) {
				action = args[i+1]
				i++
			}
		case "--order-ids":
			if i+1 < len(args) {
				orderIDsStr = args[i+1]
				i++
			}
		}
	}

	if userID == "" || action == "" || orderIDsStr == "" {
		fmt.Println("ERROR: VALIDATION_FAILED: Все параметры обязательны")
		return
	}

	orderIDs := strings.Split(orderIDsStr, ",")
	results := ProcessOrders(storage, userID, action, orderIDs)

	for _, res := range results {
		fmt.Println(res)
	}
}

// handleListOrders Получить список заказов
func handleListOrders(storage storage.Storage, args []string) {
	var userID string
	var inPvzOnly bool
	var lastCount int
	var page, limit int

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userID = args[i+1]
				i++
			}
		case "--in-pvz":
			inPvzOnly = true
		case "--last":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err == nil {
					lastCount = n
				}
				i++
			}
		case "--page":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err == nil {
					page = n
				}
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err == nil {
					limit = n
				}
				i++
			}
		}
	}

	if userID == "" {
		fmt.Println("ERROR: VALIDATION_FAILED: --user-id обязателен")
		return
	}

	orders := ListOrders(storage, userID, inPvzOnly, lastCount, page, limit)
	for _, o := range orders {
		fmt.Printf("ORDER: %s %s %s %s\n", o.ID, o.UserID, o.Status, o.ExpiresAt.Format("2006-01-02"))
	}
	fmt.Printf("TOTAL: %d\n", len(orders))
}

// handleListReturns Получить список возвратов
func handleListReturns(storage storage.Storage, args []string) {
	var err error
	var page, limit int
	page = 0
	limit = 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--page":
			if i+1 < len(args) {
				page, err = strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Printf("ERROR: VALIDATION_FAILED: %v", err)
				}
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				limit, _ = strconv.Atoi(args[i+1])
				i++
			}
		}
	}

	returns := ListReturns(storage, page, limit)
	for _, o := range returns {
		returnedAt := "N/A"
		if o.IssuedAt != nil {
			returnedAt = o.IssuedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("RETURN: %s %s %s\n", o.ID, o.UserID, returnedAt)
	}
	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
}

// handleOrderHistory Получить историю заказов
func handleOrderHistory() {
	file, err := os.Open("order_history.json")
	if err != nil {
		fmt.Println("ERROR: OPEN_FAILED: не открывается файл", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var record map[string]string
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			fmt.Printf("ERROR: JSON_FAILED: \n", err)
			continue
		}
		fmt.Printf("HISTORY: %s %s %s\n", record["order_id"], record["status"], record["timestamp"])
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR: чтение файла:", err)
	}
}

// handleImportOrders Импорт заказов из файла
func handleImportOrders(storage storage.Storage, args []string) {
	var filePath string

	for i := 0; i < len(args); i++ {
		if args[i] == "--file" && i+1 < len(args) {
			filePath = args[i+1]
			i++
		}
	}

	if filePath == "" {
		fmt.Println("ERROR: VALIDATION_FAILED: Параметр --file обязателен")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("ERROR: READ_FILE_ERROR: не удалось прочитать файл:", err)
		return
	}

	var orders []models.Order
	err = json.Unmarshal(data, &orders)
	if err != nil {
		fmt.Println("ERROR: INVALID: некорректный JSON:", err)
		return
	}

	imported := 0
	for _, o := range orders {
		err := storage.SaveOrder(o)
		if err != nil {
			fmt.Printf("ERROR: IMPORT_FAILED: не удалось импортировать заказ %s: %v\n", o.ID, err)
			continue
		}
		imported++
	}

	fmt.Printf("IMPORTED: %d\n", imported)
}

// handleScrollOrders прокрутка
func handleScrollOrders(storage storage.Storage, args []string) {
	var userID string
	var limit = 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userID = args[i+1]
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err == nil {
					limit = n
				}
				i++
			}
		}
	}

	if userID == "" {
		fmt.Println("ERROR: VALIDATION_FAILED: --user-id обязателен")
		return
	}

	lastID := "0"

	reader := bufio.NewReader(os.Stdin)

	for {
		orders, nextLastID := ScrollOrders(storage, userID, lastID, limit)

		for _, o := range orders {
			fmt.Printf("ORDER: %s %s %s %s\n", o.ID, o.UserID, o.Status, o.ExpiresAt.Format("2006-01-02"))
		}

		if nextLastID != "" && nextLastID != lastID {
			fmt.Printf("NEXT: %s\n", nextLastID)
		} else {
			fmt.Println("NEXT: ")
			fmt.Println("Больше заказов нет")
			break
		}

		fmt.Print("> ")
		cmdLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения ввода:", err)
			break
		}
		cmdLine = strings.TrimSpace(cmdLine)

		if cmdLine == "exit" {
			break
		} else if cmdLine == "next" {
			lastID = nextLastID
			continue
		} else {
			fmt.Println("Команда не распознана. Введите next или exit")
		}
	}
}
