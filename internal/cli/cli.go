package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"

	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
)

const (
	dateFormat = "2006-01-02"
)

func Run(storage *storage.FileStorage, scanner *bufio.Scanner) {
	ctx := context.Background()

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
			handleAcceptOrder(ctx, storage, args[1:])
		case "return-order":
			handleReturnOrder(ctx, storage, args[1:])
		case "process-order":
			handleProcessOrders(ctx, storage, args[1:])
		case "list-orders":
			handleListOrders(ctx, storage, args[1:])
		case "list-returns":
			handleListReturns(storage, args[1:])
		case "order-history":
			handleOrderHistory(ctx)
		case "import-orders":
			handleImportOrders(ctx, storage, args[1:])
		case "scroll-orders":
			handleScrollOrders(ctx, storage, args[1:])
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}

func printHelp() {
	fmt.Println("Список команд:")
	fmt.Println("  help")
	fmt.Println("  accept-order     Принять заказ от курьера")
	fmt.Println("  return-order     Вернуть заказ") //удалить значит
	fmt.Println("  process-order   	Выдать или принять возврат")
	fmt.Println("  list-orders    	Получить список заказов")
	fmt.Println("  list-returns    	Получить список возвратов")
	fmt.Println("  order-history   	Получить историю заказов")
	fmt.Println("  import-orders   	Импорт заказов из файла")
	fmt.Println("  scroll-orders   	Прокрутка")
	fmt.Println("  exit             Выйти из приложения")
}

// handleAcceptOrder Принять заказ от курьера
func handleAcceptOrder(ctx context.Context, storage *storage.FileStorage, args []string) {
	var orderID, userID, expiresStr string
	var weight, price float64
	var package_type models.PackageType

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
		case "--weight":
			if i+1 < len(args) {
				var err error
				weight, err = strconv.ParseFloat(args[i+1], 64)
				if err != nil || weight <= 0 {
					logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "некорректный вес")
					return
				}
				i++
			}
		case "--price":
			if i+1 < len(args) {
				var err error
				price, err = strconv.ParseFloat(args[i+1], 64)
				if err != nil || price <= 0 {
					logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "некорректная цена")
					return
				}
				i++
			}
		case "--package":
			if i+1 < len(args) {
				package_type = models.PackageType(args[i+1])
				i++
			}
		}

	}

	switch {
	case orderID == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует orderID")
		return
	case userID == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	case expiresStr == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует expiresStr")
		return
	}

	expiresAt, err := time.Parse(dateFormat, expiresStr)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "неверный формат даты")
		return
	}
	newOrder, err := AcceptOrder(ctx, storage, orderID, userID, weight, price, expiresAt, package_type)
	//TODO: ВОТ ТАК НАДО ОШИБКИ СРАВНИВАТЬ которые приходят откуда-то
	if errors.Is(err, domainErrors.ErrWeightTooHeavy) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "вес слишком большой")
	} else if err != nil {
		logger.LogErrorWithCode(ctx, err, "такой заказ уже существует или срок хранения в прошлом")

	} else {
		fmt.Println("ORDER_ACCEPTED:", orderID)
		fmt.Println("PACKAGE:", package_type)
		fmt.Println("TOTAL_PRICE:", newOrder.Price)
	}
}

// handleReturnOrder Вернуть заказ
func handleReturnOrder(ctx context.Context, storage *storage.FileStorage, args []string) {
	var orderID string

	for i := 0; i < len(args); i++ {
		if args[i] == "--order-id" && i+1 < len(args) {
			orderID = args[i+1]
			i++
		}
	}

	if orderID == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует orderID")
		return
	}

	err := ReturnOrder(storage, orderID)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "заказ у клиента или время хранения еще не истекло")
	} else {
		fmt.Println("ORDER_RETURNED:", orderID)
	}
}

// handleProcessOrders Выдать или принять возврат
func handleProcessOrders(ctx context.Context, storage storage.Storage, args []string) {
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
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствуют необходимые параметры")
		return
	}

	orderIDs := strings.Split(orderIDsStr, ",")
	results := ProcessOrders(ctx, storage, userID, action, orderIDs)

	for _, res := range results {
		fmt.Println(res)
	}
}

// handleListOrders Получить список заказов
func handleListOrders(ctx context.Context, storage storage.Storage, args []string) {
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
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	}

	orders := ListOrders(ctx, storage, userID, inPvzOnly, lastCount, page, limit)
	for _, o := range orders {
		fmt.Printf("ORDER: %s %s %s %s %s %f %f\n", o.ID, o.UserID, o.Status, o.ExpiresAt.Format(dateFormat), o.PackageType, o.Weight, o.Price)
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
					fmt.Printf("ERROR: STRCONV_FAILED: %v", err)
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
		returnedAt := "Нет данных"
		if o.IssuedAt != nil {
			returnedAt = o.IssuedAt.Format(dateTimeFormat)
		}
		fmt.Printf("RETURN: %s %s %s\n", o.ID, o.UserID, returnedAt)
	}
	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
}

// handleOrderHistory Получить историю заказов
func handleOrderHistory(ctx context.Context) {
	file, err := os.Open("order_history.json")
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOpenFiled, "не открывается файл")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var record map[string]string
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			fmt.Printf("ERROR: JSON_FAILED: %v\n", err)
			continue
		}
		fmt.Printf("HISTORY: %s %s %s\n", record["order_id"], record["status"], record["timestamp"])
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("ERROR: READ_FAILED: %v", err)
	}
}

// handleImportOrders Импорт заказов из файла
func handleImportOrders(ctx context.Context, storage storage.Storage, args []string) {
	var filePath string

	for i := 0; i < len(args); i++ {
		if args[i] == "--file" && i+1 < len(args) {
			filePath = args[i+1]
			i++
		}
	}
	if filePath == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "параметр --file обязателен")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "параметр --file обязателен")
		return
	}

	var orders []models.Order
	err = json.Unmarshal(data, &orders)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrJsonFiled, "некорректный JSON")
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
func handleScrollOrders(ctx context.Context, storage storage.Storage, args []string) {
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
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	}

	lastID := "0"

	reader := bufio.NewReader(os.Stdin)

	for {
		orders, nextLastID := ScrollOrders(storage, userID, lastID, limit)

		for _, o := range orders {
			fmt.Printf("ORDER: %s %s %s %s %s %f %f\n", o.ID, o.UserID, o.Status, o.ExpiresAt.Format(dateFormat), o.PackageType, o.Weight, o.Price)
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
			logger.LogErrorWithCode(ctx, domainErrors.ErrReadFiled, "ошибка чтения")
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
