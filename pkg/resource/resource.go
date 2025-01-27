package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"tamerGoClient/pkg/auth"
	"tamerGoClient/pkg/info"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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

const defaultDeploymentYAML = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: client-access
  labels:
    app: my-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
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
            cpu: 250m`

const defaultServiceYAML = `apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: client-access
  labels:
    app: my-app
spec:
  selector:
    app: my-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
  type: ClusterIP`

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
	case *appsv1.Deployment:
		_, err = auth.KubeClient.AppsV1().Deployments(o.Namespace).Create(context.Background(), o, metav1.CreateOptions{})
	case *corev1.Service:
		_, err = auth.KubeClient.CoreV1().Services(o.Namespace).Create(context.Background(), o, metav1.CreateOptions{})
	default:
		return fmt.Errorf("desteklenmeyen resource tipi")
	}
	return err
}

func HandleResourceMenu() {
	if auth.KubeClient == nil {
		fmt.Println("\nUyarı: Önce bir Kubernetes cluster'ına bağlanmalısınız!")
		return
	}

	for {
		fmt.Println("\n=== Instance Oluştur/Sil ===")
		fmt.Println("1. Pod Yönetim Menüsü")
		fmt.Println("2. Deployment Yönetim Menüsü")
		fmt.Println("3. Service Yönetim Menüsü")
		fmt.Println("4. Ana Menüye Dön")
		fmt.Print("Seçiminiz (1-4): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			handlePodMenu()
		case 2:
			handleDeploymentMenu()
		case 3:
			handleServiceMenu()
		case 4:
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

func handleDeploymentMenu() {
	for {
		fmt.Println("\n=== Deployment Yönetim Menüsü ===")
		fmt.Println("1. Deployment Oluştur")
		fmt.Println("2. Deployment Sil")
		fmt.Println("3. Deploymentları Listele")
		fmt.Println("4. Önceki Menüye Dön")
		fmt.Print("Seçiminiz (1-4): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			createDeployment()
		case 2:
			deleteDeployment()
		case 3:
			info.ListDeploymentsWithDetails()
		case 4:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func handleServiceMenu() {
	for {
		fmt.Println("\n=== Service Yönetim Menüsü ===")
		fmt.Println("1. Service Oluştur")
		fmt.Println("2. Service Sil")
		fmt.Println("3. Service'leri Listele")
		fmt.Println("4. Önceki Menüye Dön")
		fmt.Print("Seçiminiz (1-4): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			createService()
		case 2:
			deleteService()
		case 3:
			info.ListServicesWithDetails()
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

func createDeployment() {
	fmt.Println("\nVarsayılan Deployment YAML editörde açılacak...")
	editedContent, err := openInEditor(defaultDeploymentYAML)
	if err != nil {
		fmt.Printf("Editör hatası: %v\n", err)
		return
	}

	// YAML'dan deployment bilgilerini çıkar
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(editedContent), nil, nil)
	if err != nil {
		fmt.Printf("YAML parse hatası: %v\n", err)
		return
	}

	if deployment, ok := obj.(*appsv1.Deployment); ok {
		// Deployment'ı oluştur
		createdDeployment, err := auth.KubeClient.AppsV1().Deployments(deployment.Namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Deployment oluşturma hatası: %v\n", err)
			return
		}

		fmt.Println("Deployment başarıyla oluşturuldu! 3 saniye sonra detaylar görüntülenecek...")
		time.Sleep(3 * time.Second)

		// Güncel deployment bilgilerini al ve göster
		updatedDeployment, err := auth.KubeClient.AppsV1().Deployments(createdDeployment.Namespace).Get(context.Background(), createdDeployment.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Deployment detayları alınamadı: %v\n", err)
			return
		}
		info.ShowDeploymentDetails(*updatedDeployment)
	}
}

func deleteDeployment() {
	// Mevcut deploymentları listele
	ctx := context.Background()
	deployments, err := auth.KubeClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Deployment listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nMevcut Deploymentlar:")
	fmt.Printf("%-5s %-30s %-20s %-10s\n", "NO", "İSİM", "NAMESPACE", "REPLICAS")

	deployList := make([]appsv1.Deployment, 0)
	for i, deploy := range deployments.Items {
		fmt.Printf("%-5d %-30s %-20s %d/%d\n",
			i+1,
			deploy.Name,
			deploy.Namespace,
			deploy.Status.ReadyReplicas,
			deploy.Status.Replicas)
		deployList = append(deployList, deploy)
	}

	fmt.Print("\nSilmek istediğiniz deployment'ın numarasını girin (0 için iptal): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice > 0 && choice <= len(deployList) {
		selectedDeploy := deployList[choice-1]
		fmt.Printf("\nDeployment'ı silmek istediğinizden emin misiniz? (%s/%s) [e/h]: ",
			selectedDeploy.Namespace, selectedDeploy.Name)

		var confirm string
		fmt.Scanf("%s", &confirm)

		if confirm == "e" {
			err := auth.KubeClient.AppsV1().Deployments(selectedDeploy.Namespace).Delete(ctx, selectedDeploy.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("Deployment silinemedi: %v\n", err)
			} else {
				fmt.Println("Deployment başarıyla silindi! Etkilerin görüntülenmesi için bir süre bekleyiniz...")
				time.Sleep(3 * time.Second)
				info.ListDeploymentsWithDetails()
			}
		}
	}
}

func createService() {
	fmt.Println("\nVarsayılan Service YAML editörde açılacak...")
	editedContent, err := openInEditor(defaultServiceYAML)
	if err != nil {
		fmt.Printf("Editör hatası: %v\n", err)
		return
	}

	// YAML'dan service bilgilerini çıkar
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(editedContent), nil, nil)
	if err != nil {
		fmt.Printf("YAML parse hatası: %v\n", err)
		return
	}

	if service, ok := obj.(*corev1.Service); ok {
		// Service'i oluştur
		createdService, err := auth.KubeClient.CoreV1().Services(service.Namespace).Create(context.Background(), service, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Service oluşturma hatası: %v\n", err)
			return
		}

		fmt.Println("Service başarıyla oluşturuldu! 2 saniye sonra detaylar görüntülenecek...")
		time.Sleep(2 * time.Second)

		// Güncel service bilgilerini al ve göster
		updatedService, err := auth.KubeClient.CoreV1().Services(createdService.Namespace).Get(context.Background(), createdService.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Service detayları alınamadı: %v\n", err)
			return
		}
		info.ShowServiceDetails(*updatedService)
	}
}

func deleteService() {
	// Mevcut service'leri listele
	ctx := context.Background()
	services, err := auth.KubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Service listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nMevcut Service'ler:")
	fmt.Printf("%-5s %-30s %-20s %-15s %-15s\n",
		"NO", "İSİM", "NAMESPACE", "TYPE", "CLUSTER-IP")

	svcList := make([]corev1.Service, 0)
	for i, svc := range services.Items {
		fmt.Printf("%-5d %-30s %-20s %-15s %-15s\n",
			i+1,
			svc.Name,
			svc.Namespace,
			svc.Spec.Type,
			svc.Spec.ClusterIP)
		svcList = append(svcList, svc)
	}

	fmt.Print("\nSilmek istediğiniz service'in numarasını girin (0 için iptal): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice > 0 && choice <= len(svcList) {
		selectedSvc := svcList[choice-1]
		fmt.Printf("\nService'i silmek istediğinizden emin misiniz? (%s/%s) [e/h]: ",
			selectedSvc.Namespace, selectedSvc.Name)

		var confirm string
		fmt.Scanf("%s", &confirm)

		if confirm == "e" {
			err := auth.KubeClient.CoreV1().Services(selectedSvc.Namespace).Delete(ctx, selectedSvc.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("Service silinemedi: %v\n", err)
			} else {
				fmt.Println("Service başarıyla silindi! Etkilerin görüntülenmesi için bir süre bekleyiniz...")
				time.Sleep(2 * time.Second)
				info.ListServicesWithDetails()
			}
		}
	}
}
