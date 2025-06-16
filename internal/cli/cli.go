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
	"PWZ1.0/internal/service"

	"PWZ1.0/internal/tools/logger"
)

const (
	dateFormat = "2006-01-02"
)

type CLI struct {
	orderService service.OrderService
	scanner      *bufio.Scanner
}

func NewCLI(orderService service.OrderService, scanner *bufio.Scanner) *CLI {
	return &CLI{
		orderService: orderService,
		scanner:      scanner,
	}
}

func (c *CLI) Run() {
	ctx := context.Background()
	for {
		fmt.Print("> ")
		if !c.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(c.scanner.Text())
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
			handleAcceptOrder(ctx, c.orderService, args[1:])
		case "return-order":
			handleReturnOrder(ctx, c.orderService, args[1:])
		case "process-order":
			handleProcessOrders(ctx, c.orderService, args[1:])
		case "list-orders":
			handleListOrders(ctx, c.orderService, args[1:])
		case "list-returns":
			handleListReturns(c.orderService, args[1:])
		case "order-history":
			//handleOrderHistory(ctx)
			handleOrderHistory(c.orderService, args[1:])

		case "import-orders":
			handleImportOrders(ctx, c.orderService, args[1:])
		case "scroll-orders":
			handleScrollOrders(ctx, c.orderService, args[1:])
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
func handleAcceptOrder(ctx context.Context, orderService service.OrderService, args []string) {
	var orderIDStr, userIDStr, expiresStr string
	var orderID, userID uint64
	var weight, price float32
	var packageType models.PackageType

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--order-id":
			if i+1 < len(args) {
				orderIDStr = args[i+1]
				i++
			}
		case "--user-id":
			if i+1 < len(args) {
				userIDStr = args[i+1]
				i++
			}
		case "--expires":
			if i+1 < len(args) {
				expiresStr = args[i+1]
				i++
			}
		case "--weight":
			if i+1 < len(args) {
				weight64, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || weight64 <= 0 {
					logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "некорректный вес")
					return
				}
				weight = float32(weight64)
				i++
			}
		case "--price":
			if i+1 < len(args) {
				price64, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || price64 <= 0 {
					logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "некорректная цена")
					return
				}
				price = float32(price64)
				i++
			}
		case "--package":
			if i+1 < len(args) {
				packageType = models.PackageType(args[i+1])
				i++
			}
		}
	}

	switch {
	case orderIDStr == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует orderID")
		return
	case userIDStr == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	case expiresStr == "":
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует expiresStr")
		return
	}

	var err error
	orderID, err = strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "orderID должен быть числом")
		return
	}

	userID, err = strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "userID должен быть числом")
		return
	}

	expiresAt, err := time.Parse(dateFormat, expiresStr)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "неверный формат даты")
		return
	}

	newOrder, err := orderService.AcceptOrder(ctx, orderID, userID, weight, price, expiresAt, packageType)
	if errors.Is(err, domainErrors.ErrInvalidPackage) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrInvalidPackage, "некорректная упаковка")
	} else if errors.Is(err, domainErrors.ErrWeightTooHeavy) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrWeightTooHeavy, "вес слишком большой")
	} else if err != nil {
		logger.LogErrorWithCode(ctx, err, "такой заказ уже существует или срок хранения в прошлом")
	} else {
		fmt.Println("ORDER_ACCEPTED:", orderID)
		fmt.Println("PACKAGE:", packageType)
		fmt.Println("TOTAL_PRICE:", newOrder.Price)
	}
}

// handleReturnOrder Вернуть заказ
func handleReturnOrder(ctx context.Context, orderService service.OrderService, args []string) {
	var orderIDStr string

	for i := 0; i < len(args); i++ {
		if args[i] == "--order-id" && i+1 < len(args) {
			orderIDStr = args[i+1]
			i++
		}
	}

	if orderIDStr == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует orderID")
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "неверный формат orderID")
		return
	}

	resp, err := orderService.ReturnOrder(orderID)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "заказ у клиента или время хранения еще не истекло")
		return
	}
	fmt.Printf("ORDER_RETURNED: ID=%d STATUS=%s\n", resp.OrderID, resp.Status)
}

// handleProcessOrders Выдать или принять возврат
func handleProcessOrders(ctx context.Context, orderService service.OrderService, args []string) {
	var userIDStr, actionStr, orderIDsStr string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userIDStr = args[i+1]
				i++
			}
		case "--action":
			if i+1 < len(args) {
				actionStr = args[i+1]
				i++
			}
		case "--order-ids":
			if i+1 < len(args) {
				orderIDsStr = args[i+1]
				i++
			}
		}
	}

	if userIDStr == "" || actionStr == "" || orderIDsStr == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствуют необходимые параметры")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "userID должен быть числом")
		return
	}

	orderIDsStrSlice := strings.Split(orderIDsStr, ",")
	orderIDs := make([]uint64, 0, len(orderIDsStrSlice))
	for _, idStr := range orderIDsStrSlice {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, fmt.Sprintf("некорректный orderID: %s", idStr))
			return
		}
		orderIDs = append(orderIDs, id)
	}

	actionType := models.ParseActionType(actionStr)
	if actionType == models.ActionTypeUnspecified {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "неизвестный тип действия")
		return
	}

	result := orderService.ProcessOrders(ctx, userID, actionType, orderIDs)

	for _, id := range result.Processed {
		fmt.Printf("PROCESSED: %d\n", id)
	}
	for _, id := range result.Errors {
		fmt.Printf("ERROR: %d\n", id)
	}
}

// handleListOrders Получить список заказов
func handleListOrders(ctx context.Context, orderService service.OrderService, args []string) {
	var userIDStr string
	var userID uint64
	var inPvzOnly bool
	var lastId uint32
	var page, limit uint32

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userIDStr = args[i+1]
				i++
			}
		case "--in-pvz":
			inPvzOnly = true
		case "--last":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					lastId = uint32(n)
				}
				i++
			}
		case "--page":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					page = uint32(n)
				}
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					limit = uint32(n)
				}
				i++
			}
		}
	}

	if userIDStr == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	}

	var err error
	userID, err = strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "userID должен быть числом")
		return
	}

	orders, total := orderService.ListOrders(ctx, userID, inPvzOnly, lastId, page, limit)
	for _, o := range orders {
		fmt.Printf("ORDER: %d %d %s %s %s %.2f %.2f\n",
			o.ID,
			o.UserID,
			o.Status,
			o.ExpiresAt.Format(dateFormat),
			o.PackageType,
			o.Weight,
			o.Price,
		)
	}
	fmt.Printf("TOTAL: %d\n", total)
}

// handleListReturns Получить список возвратов
func handleListReturns(orderService service.OrderService, args []string) {
	var page uint32 = 0
	var limit uint32 = 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--page":
			if i+1 < len(args) {
				p, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Printf("ERROR: invalid --page value: %v\n", err)
					return
				}
				page = uint32(p)
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Printf("ERROR: invalid --limit value: %v\n", err)
					return
				}
				limit = uint32(l)
				i++
			}
		}
	}

	req := service.ListReturnsRequest{
		Pagination: service.Pagination{
			Page:        page,
			CountOnPage: limit,
		},
	}

	res := orderService.ListReturns(req)
	for _, o := range res.Returns {
		approxReturnTime := o.ExpiresAt.Add(-service.ExpiredTime)
		fmt.Printf("RETURN: %d %d %s\n", o.ID, o.UserID, approxReturnTime.Format(service.DateTimeFormat))
	}
	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
}

// handleOrderHistory Получить историю заказов
func handleOrderHistory(orderService service.OrderService, args []string) {
	ctx := context.Background()
	var page, limit uint32

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--page":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					page = uint32(n)
				}
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					limit = uint32(n)
				}
				i++
			}
		default:
			logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed,
				fmt.Sprintf("неизвестный флаг: %s", args[i]))
			fmt.Println("Использование: history [--page N] [--limit M]")
			return
		}
	}

	if limit == 0 {
		limit = 10
	}

	req := service.GetHistoryRequest{
		Pagination: service.Pagination{
			Page:        page,
			CountOnPage: limit,
		},
	}

	history := orderService.GetHistory(req)

	if len(history.History) == 0 {
		fmt.Println("История изменений не найдена")
		return
	}

	for _, h := range history.History {
		fmt.Printf("%d %s %s\n",
			h.OrderID,
			h.Status,
			h.CreatedAt.Format(service.DateTimeFormat))
	}
	fmt.Printf("Всего записей: %d\n", len(history.History))
}

// handleImportOrders Импорт заказов из файла
func handleImportOrders(ctx context.Context, orderService service.OrderService, args []string) {
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
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "не удалось прочитать файл")
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
		err := orderService.SaveOrder(o)
		if err != nil {
			fmt.Printf("ERROR: IMPORT_FAILED: не удалось импортировать заказ %d: %v\n", o.ID, err)
			continue
		}
		imported++
	}

	fmt.Printf("IMPORTED: %d\n", imported)
}

// handleScrollOrders прокрутка
func handleScrollOrders(ctx context.Context, orderService service.OrderService, args []string) {
	var userIDStr string
	limit := 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user-id":
			if i+1 < len(args) {
				userIDStr = args[i+1]
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

	if userIDStr == "" {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "отсутствует userID")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "userID должен быть числом")
		return
	}

	var lastID uint64 = 0

	reader := bufio.NewReader(os.Stdin)

	for {
		orders, nextLastID := orderService.ScrollOrders(userID, lastID, limit)

		for _, o := range orders {
			fmt.Printf("ORDER: %d %d %s %s %s %.2f %.2f\n",
				o.ID,
				o.UserID,
				o.Status,
				o.ExpiresAt.Format(dateFormat),
				o.PackageType,
				o.Weight,
				o.Price,
			)
		}

		if nextLastID != 0 && nextLastID != lastID {
			fmt.Printf("NEXT: %d\n", nextLastID)
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
