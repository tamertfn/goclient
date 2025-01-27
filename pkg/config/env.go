package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type EnvManager struct {
	filePath       string
	envMap         map[string]string
	predefinedKeys []string
}

func NewEnvManager(filePath string) *EnvManager {
	return &EnvManager{
		filePath: filePath,
		envMap:   make(map[string]string),
		predefinedKeys: []string{
			"API_SERVER",
			"K8S_TOKEN",
			"CA_CERT_PATH",
			"KUBECONFIG_PATH",
		},
	}
}

func (em *EnvManager) Load() error {
	file, err := os.Open(em.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Dosya yoksa boş map ile devam et
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			em.envMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return scanner.Err()
}

func (em *EnvManager) Save() error {
	file, err := os.Create(em.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range em.envMap {
		_, err := fmt.Fprintf(file, "%s=%s\n", key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (em *EnvManager) Set(key, value string) {
	em.envMap[key] = value
}

func (em *EnvManager) Get(key string) string {
	return em.envMap[key]
}

func (em *EnvManager) ShowEnvMenu() {
	for {
		fmt.Println("\n=== .env Dosyası Yönetimi ===")
		fmt.Println("1. Mevcut Değerleri Göster")
		fmt.Println("2. Değer Ekle/Güncelle")
		fmt.Println("3. Önceki Menüye Dön")

		var choice int
		fmt.Print("Seçiminiz (1-3): ")
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			em.displayCurrentValues()
		case 2:
			em.updateValue()
		case 3:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func (em *EnvManager) displayCurrentValues() {
	fmt.Println("\nMevcut .env Değerleri:")
	if len(em.envMap) == 0 {
		fmt.Println("Henüz hiç değer yok.")
		return
	}
	for key, value := range em.envMap {
		fmt.Printf("%s=%s\n", key, value)
	}
}

func (em *EnvManager) updateValue() {
	fmt.Println("\nMevcut anahtarlar:")
	for i, key := range em.predefinedKeys {
		currentValue := em.Get(key)
		if currentValue == "" {
			currentValue = "<boş>"
		}
		fmt.Printf("%d. %s = %s\n", i+1, key, currentValue)
	}

	var choice int
	fmt.Print("\nGüncellenecek değerin numarası (1-4): ")
	fmt.Scanf("%d", &choice)

	if choice < 1 || choice > len(em.predefinedKeys) {
		fmt.Println("Geçersiz seçim!")
		return
	}

	key := em.predefinedKeys[choice-1]
	fmt.Printf("Yeni değer (%s): ", key)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		em.Set(key, value)
		if err := em.Save(); err != nil {
			fmt.Printf("Hata: Değerler kaydedilemedi: %v\n", err)
			return
		}
		fmt.Println("Değer başarıyla kaydedildi!")
	}
}
