package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"tamerGoClient/pkg/auth"
	"tamerGoClient/pkg/info"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const defaultPodYAML = `apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: client-access
  labels:
    app: my-app
spec:
  containers:
  - name: my-container
    image: nginx:latest
    ports:
    - containerPort: 80
    resources:
      limits:
        memory: 128Mi
        cpu: 500m
      requests:
        memory: 64Mi
        cpu: 250m
`

func openInEditor(content string) (string, error) {
	// Geçici dosya oluştur
	tmpfile, err := os.CreateTemp("", "k8s-*.yaml")
	if err != nil {
		return "", fmt.Errorf("geçici dosya oluşturulamadı: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// İçeriği geçici dosyaya yaz
	if err := os.WriteFile(tmpfile.Name(), []byte(content), 0644); err != nil {
		return "", fmt.Errorf("dosya yazılamadı: %v", err)
	}

	// Tercih edilen editörü belirle
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Varsayılan olarak vim kullan
	}

	// Editörü aç
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editör çalıştırılamadı: %v", err)
	}

	// Düzenlenmiş içeriği oku
	editedContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("düzenlenmiş içerik okunamadı: %v", err)
	}

	return string(editedContent), nil
}

func createFromYAML(yamlContent string) error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlContent), nil, nil)
	if err != nil {
		return fmt.Errorf("YAML ayrıştırılamadı: %v", err)
	}

	// Resource türüne göre create işlemi yap
	switch o := obj.(type) {
	case *corev1.Pod:
		_, err = auth.KubeClient.CoreV1().Pods(o.Namespace).Create(context.Background(), o, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		// Pod ismini ve namespace'i döndür
		return nil
	default:
		return fmt.Errorf("desteklenmeyen resource tipi")
	}
}

func HandleResourceMenu() {
	if auth.KubeClient == nil {
		fmt.Println("\nUyarı: Önce bir Kubernetes cluster'ına bağlanmalısınız!")
		return
	}

	for {
		fmt.Println("\n=== Instance Oluştur/Sil ===")
		fmt.Println("1. Pod Yönetim Menüsü")
		fmt.Println("2. Ana Menüye Dön")
		fmt.Print("Seçiminiz (1-2): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			handlePodMenu()
		case 2:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func handlePodMenu() {
	for {
		fmt.Println("\n=== Pod Yönetim Menüsü ===")
		fmt.Println("1. Pod Oluştur")
		fmt.Println("2. Pod Sil")
		fmt.Println("3. Podları Listele")
		fmt.Println("4. Önceki Menüye Dön")
		fmt.Print("Seçiminiz (1-4): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			createPod()
		case 2:
			deletePod()
		case 3:
			info.ListPods()
		case 4:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func deletePod() {
	// Mevcut podları listele
	ctx := context.Background()
	pods, err := auth.KubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Pod listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nMevcut Podlar:")
	fmt.Printf("%-5s %-30s %-20s %-12s\n", "NO", "İSİM", "NAMESPACE", "DURUM")

	podList := make([]corev1.Pod, 0)
	for i, pod := range pods.Items {
		fmt.Printf("%-5d %-30s %-20s %-12s\n",
			i+1,
			pod.Name,
			pod.Namespace,
			string(pod.Status.Phase))
		podList = append(podList, pod)
	}

	fmt.Print("\nSilmek istediğiniz pod'un numarasını girin (0 için iptal): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice > 0 && choice <= len(podList) {
		selectedPod := podList[choice-1]
		fmt.Printf("\nPod'u silmek istediğinizden emin misiniz? (%s/%s) [e/h]: ",
			selectedPod.Namespace, selectedPod.Name)

		var confirm string
		fmt.Scanf("%s", &confirm)

		if confirm == "e" {
			err := auth.KubeClient.CoreV1().Pods(selectedPod.Namespace).Delete(ctx, selectedPod.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("Pod silinemedi: %v\n", err)
			} else {
				fmt.Println("Pod başarıyla silindi! Etkilerin görüntülenmesi için bir süre bekleyiniz...")
				time.Sleep(3 * time.Second)
				info.ListPods()
			}
		}
	}
}

func createPod() {
	fmt.Println("\nVarsayılan Pod YAML editörde açılacak...")
	editedContent, err := openInEditor(defaultPodYAML)
	if err != nil {
		fmt.Printf("Editör hatası: %v\n", err)
		return
	}

	// YAML'dan pod bilgilerini çıkar
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(editedContent), nil, nil)
	if err != nil {
		fmt.Printf("YAML parse hatası: %v\n", err)
		return
	}

	if pod, ok := obj.(*corev1.Pod); ok {
		// Pod'u oluştur
		if err := createFromYAML(editedContent); err != nil {
			fmt.Printf("Pod oluşturma hatası: %v\n", err)
		} else {
			fmt.Println("Pod başarıyla oluşturuldu! 2 saniye sonra detaylar görüntülenecek...")

			// Pod'un oluşmasını bekle ve güncel bilgileri al
			time.Sleep(2 * time.Second)
			createdPod, err := auth.KubeClient.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Pod detayları alınamadı: %v\n", err)
				return
			}
			info.ShowPodDetails(*createdPod)
		}
	}
}
