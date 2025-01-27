package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamerGoClient/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	KubeClient       *kubernetes.Clientset
	envManager       *config.EnvManager
	activeConnection string // Aktif bağlantı bilgisini tutacak
)

func init() {
	envManager = config.NewEnvManager(".env")
	if err := envManager.Load(); err != nil {
		fmt.Printf("Uyarı: .env dosyası yüklenirken hata oluştu: %v\n", err)
	}
}

type ServiceAccountConfig struct {
	ServerURL string
	Token     string
	CAData    []byte
	Timeout   time.Duration
}

func showAuthMenu() int {
	fmt.Println("\n=== Kimlik Doğrulama Yöntemleri ===")
	if activeConnection != "" {
		fmt.Printf("(Aktif Bağlantı: via %s)\n", activeConnection)
	}
	fmt.Println("1. Service Account ile Bağlan")
	fmt.Println("2. In-Cluster Bağlantı")
	fmt.Println("3. Kubeconfig ile Bağlan")
	fmt.Println("4. Ana Menüye Dön")
	fmt.Print("Seçiminiz (1-4): ")

	var choice int
	fmt.Scanf("%d", &choice)
	return choice
}

func showServiceAccountMenu() int {
	fmt.Println("\n=== Service Account Bağlantı Menüsü ===")
	if activeConnection != "" && strings.HasPrefix(activeConnection, "ServiceAccount") {
		fmt.Printf("(Aktif Bağlantı: via %s)\n", activeConnection)
	}
	fmt.Println("\nGerekli .env değerleri:")
	fmt.Println("- API_SERVER: Kubernetes API sunucu adresi")
	fmt.Println("- K8S_TOKEN: Service Account token değeri")
	fmt.Println("- CA_CERT_PATH: CA sertifika dosyasının yolu")
	fmt.Println("\nSeçenekler:")
	fmt.Println("1. Bağlantıyı Yapılandır")
	fmt.Println("2. .env Dosyasını Düzenle")
	fmt.Println("3. Önceki Menüye Dön")
	fmt.Print("Seçiminiz (1-3): ")

	var choice int
	fmt.Scanf("%d", &choice)
	return choice
}

func showKubeconfigMenu() int {
	fmt.Println("\n=== Kubeconfig Bağlantı Menüsü ===")
	if activeConnection != "" && strings.HasPrefix(activeConnection, "Kubeconfig") {
		fmt.Printf("(Aktif Bağlantı: %s)\n", activeConnection)
	}
	fmt.Println("\nGerekli .env değeri:")
	fmt.Println("- KUBECONFIG_PATH: Kubeconfig dosyasının yolu")
	fmt.Println("\nSeçenekler:")
	fmt.Println("1. Bağlantıyı Yapılandır")
	fmt.Println("2. .env Dosyasını Düzenle")
	fmt.Println("3. Önceki Menüye Dön")
	fmt.Print("Seçiminiz (1-3): ")

	var choice int
	fmt.Scanf("%d", &choice)
	return choice
}

func waitForMainMenu() {
	fmt.Print("\nAna menüye dönmek için herhangi bir tuşa basın: ")
	var input string
	fmt.Scanf("%s", &input)
}

func handleServiceAccountMenu() {
	for {
		choice := showServiceAccountMenu()

		switch choice {
		case 1:
			connectWithServiceAccount()
			return // Bağlantı başarılı olduğunda direkt ana menüye dön
		case 2:
			envManager.ShowEnvMenu()
		case 3:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func handleKubeconfigMenu() {
	for {
		choice := showKubeconfigMenu()

		switch choice {
		case 1:
			connectWithKubeconfig()
			return // Bağlantı başarılı olduğunda direkt ana menüye dön
		case 2:
			envManager.ShowEnvMenu()
		case 3:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func HandleAuthMenu() {
	for {
		choice := showAuthMenu()

		switch choice {
		case 1:
			handleServiceAccountMenu()
			if KubeClient != nil {
				return // Bağlantı başarılı olduğunda ana menüye dön
			}
		case 2:
			connectInCluster()
			if KubeClient != nil {
				return // Bağlantı başarılı olduğunda ana menüye dön
			}
		case 3:
			handleKubeconfigMenu()
			if KubeClient != nil {
				return // Bağlantı başarılı olduğunda ana menüye dön
			}
		case 4:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func connectWithServiceAccount() {
	// .env'den değerleri oku
	serverURL := envManager.Get("API_SERVER")
	token := envManager.Get("K8S_TOKEN")
	caPath := envManager.Get("CA_CERT_PATH")

	// Değerler eksikse kullanıcıdan al
	if serverURL == "" {
		fmt.Print("Kubernetes API Server URL: ")
		fmt.Scanf("%s", &serverURL)
		envManager.Set("API_SERVER", serverURL)
	}

	if token == "" {
		fmt.Print("Service Account Token: ")
		fmt.Scanf("%s", &token)
		envManager.Set("K8S_TOKEN", token)
	}

	if caPath == "" {
		fmt.Print("CA Sertifika dosya yolu: ")
		fmt.Scanf("%s", &caPath)
		envManager.Set("CA_CERT_PATH", caPath)
	}

	// Değişiklikleri kaydet
	if err := envManager.Save(); err != nil {
		fmt.Printf("Uyarı: .env dosyası kaydedilemedi: %v\n", err)
	}

	// CA sertifikasını oku
	caData, err := os.ReadFile(caPath)
	if err != nil {
		fmt.Printf("CA sertifikası okunamadı: %v\n", err)
		return
	}

	// Rest config oluştur
	config := &rest.Config{
		Host:        serverURL,
		BearerToken: token,
		Timeout:     30 * time.Second,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   caData,
			Insecure: false,
		},
	}

	// Client oluştur
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Client oluşturulamadı: %v\n", err)
		return
	}

	// Bağlantıyı test et
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Bağlantı testi başarısız: %v\n", err)
		return
	}

	KubeClient = client
	serverURL = envManager.Get("API_SERVER")
	activeConnection = fmt.Sprintf("ServiceAccount (%s)", serverURL)
	fmt.Println("Service Account ile bağlantı başarılı!")
	waitForMainMenu()

}

func connectInCluster() {
	fmt.Println("\n=== In-Cluster Bağlantı ===")
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("In-cluster config oluşturulamadı: %v\n", err)
		return
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Client oluşturulamadı: %v\n", err)
		return
	}

	KubeClient = client
	activeConnection = "In-Cluster"
	fmt.Println("In-cluster bağlantı başarılı!")
	waitForMainMenu()
}

func connectWithKubeconfig() {
	// .env'den kubeconfig yolunu oku
	kubeconfigPath := envManager.Get("KUBECONFIG_PATH")
	defaultPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if kubeconfigPath == "" {
		fmt.Printf("Kubeconfig dosya yolu [varsayılan: %s]: ", defaultPath)
		fmt.Scanf("%s", &kubeconfigPath)

		if kubeconfigPath == "" {
			kubeconfigPath = defaultPath
		}

		envManager.Set("KUBECONFIG_PATH", kubeconfigPath)
		if err := envManager.Save(); err != nil {
			fmt.Printf("Uyarı: .env dosyası kaydedilemedi: %v\n", err)
		}
	}

	// Dosyanın varlığını kontrol et
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		fmt.Printf("Hata: %s dosyası bulunamadı\n", kubeconfigPath)
		return
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Kubeconfig yüklenemedi: %v\n", err)
		return
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Client oluşturulamadı: %v\n", err)
		return
	}

	// Bağlantı başarılı olduğunda

	KubeClient = client

	// Kubeconfig'den cluster bilgisini al
	kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		fmt.Printf("Kubeconfig yüklenemedi: %v\n", err)
		return
	}
	currentContext := kubeconfig.CurrentContext
	if context, exists := kubeconfig.Contexts[currentContext]; exists {
		if cluster, exists := kubeconfig.Clusters[context.Cluster]; exists {
			activeConnection = fmt.Sprintf("Kubeconfig (%s)", cluster.Server)
		}
	}

	fmt.Printf("Kubeconfig ile bağlantı başarılı! (%s)\n", kubeconfigPath)
	waitForMainMenu()
}

// Aktif bağlantı bilgisini dışarıya açan fonksiyon
func GetActiveConnection() string {
	return activeConnection
}
