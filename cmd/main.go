package main

import (
	"bufio"
	"fmt"
	"os"

	"PWZ1.0/internal/cli"
	"PWZ1.0/internal/storage"
)

func main() {
	fmt.Println("Введите команду или 'help' для списка доступных команд.")

	storage := storage.NewFileStorage("orders.json")
	scanner := bufio.NewScanner(os.Stdin)

	cli.Run(storage, scanner)
}
