package main

import (
	"bufio"
	"fmt"
	"os"

	"PWZ1.0/internal/cli"
	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
)

func main() {
	fmt.Println("Введите команду или 'help' для списка доступных команд.")

	storage := storage.NewFileStorage("orders.json")
	scanner := bufio.NewScanner(os.Stdin)

	logger.InitLogger()
	cli.Run(storage, scanner)
}

//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
//TODO: ЧЕКПОИНТ
