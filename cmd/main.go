package main

import (
	"bufio"
	"fmt"
	"os"

	"PWZ1.0/internal/cli"
	"PWZ1.0/internal/service"
	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
)

func main() {
	fmt.Println("Введите команду или 'help' для списка доступных команд.")

	logger.InitLogger()
	scanner := bufio.NewScanner(os.Stdin)

	storage := storage.NewFileStorage("orders.json")
	orderService := service.NewOrderService(storage)
	cliApp := cli.NewCLI(orderService, scanner)
	cliApp.Run()
}
