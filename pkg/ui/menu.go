package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"tamerGoClient/pkg/auth"
)

func ShowMainMenu() int {
	fmt.Println("\n=== Kubernetes Yönetim Arayüzü ===")
	if auth.GetActiveConnection() != "" {
		fmt.Printf("(Aktif Bağlantı: %s)\n", auth.GetActiveConnection())
	}
	fmt.Println("1. Kimlik Doğrulama")
	fmt.Println("2. Cluster Bilgileri")
	fmt.Println("3. Instance Oluştur")
	fmt.Println("4. Çıkış")
	fmt.Print("Seçiminiz (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0
	}
	return choice
}
