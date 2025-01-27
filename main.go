package main

import (
	"fmt"
	"os"
	"tamerGoClient/pkg/auth"
	"tamerGoClient/pkg/create"
	"tamerGoClient/pkg/info"
	"tamerGoClient/pkg/ui"
)

func main() {
	for {
		choice := ui.ShowMainMenu()

		switch choice {
		case 1:
			auth.HandleAuthMenu()
		case 2:
			info.HandleInfoMenu()
		case 3:
			create.HandleCreateMenu()
		case 4:
			fmt.Println("Programdan çıkılıyor...")
			os.Exit(0)
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}
