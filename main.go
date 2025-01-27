package main

import (
	"fmt"
	"os"
	"tamerGoClient/pkg/auth"
	"tamerGoClient/pkg/info"
	"tamerGoClient/pkg/resource"
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
			resource.HandleResourceMenu()
		case 4:
			fmt.Println("Programdan çıkılıyor...")
			os.Exit(0)
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}
