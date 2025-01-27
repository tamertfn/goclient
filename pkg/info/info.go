package info

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"tamerGoClient/pkg/auth"
	"tamerGoClient/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func showInfoMenu() int {
	fmt.Println("\n=== Cluster Bilgileri Menüsü ===")
	if auth.GetActiveConnection() != "" {
		fmt.Printf("(Aktif Bağlantı: %s)\n", auth.GetActiveConnection())
	}
	fmt.Println("1. Namespace Listesi")
	fmt.Println("2. Node Bilgileri")
	fmt.Println("3. Pod Listesi")
	fmt.Println("4. Service Listesi")
	fmt.Println("5. Deployment Listesi")
	fmt.Println("6. ConfigMap Listesi")
	fmt.Println("7. Secret Listesi")
	fmt.Println("8. PersistentVolume Listesi")
	fmt.Println("9. PersistentVolumeClaim Listesi")
	fmt.Println("10. StatefulSet Listesi")
	fmt.Println("11. DaemonSet Listesi")
	fmt.Println("12. Ingress Listesi")
	fmt.Println("13. Ana Menüye Dön")
	fmt.Print("Seçiminiz (1-13): ")

	var choice int
	fmt.Scanf("%d", &choice)
	return choice
}

func waitForContinue() bool {
	fmt.Print("\nBaşka sorgu yapmak için 'c' tuşuna, ana menüye dönmek için 'q' tuşuna basın: ")
	var input string
	fmt.Scanf("%s", &input)
	for input != "c" && input != "q" {
		fmt.Print("Geçersiz tuş! Başka sorgu yapmak için 'c', ana menüye dönmek için 'q' tuşuna basın: ")
		fmt.Scanf("%s", &input)
	}
	return input == "c" // true ise devam et, false ise ana menüye dön
}

func HandleInfoMenu() {
	if auth.KubeClient == nil {
		fmt.Println("\nUyarı: Önce bir Kubernetes cluster'ına bağlanmalısınız!")
		return
	}

	for {
		choice := showInfoMenu()

		switch choice {
		case 1:
			listNamespaces()
		case 2:
			listNodes()
		case 3:
			listPods()
		case 4:
			listServices()
		case 5:
			listDeployments()
		case 6:
			listConfigMaps()
		case 7:
			listSecrets()
		case 8:
			listPersistentVolumes()
		case 9:
			listPersistentVolumeClaims()
		case 10:
			listStatefulSets()
		case 11:
			listDaemonSets()
		case 12:
			listIngresses()
		case 13:
			return
		default:
			fmt.Println("Geçersiz seçim!")
			continue
		}

		// waitForContinue'dan dönen değere göre ana menüye dön veya devam et
		if !waitForContinue() {
			return // Ana menüye dön
		}
	}
}

func listNamespaces() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaces, err := auth.KubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Namespace listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nNamespace Listesi:")
	fmt.Printf("%-30s %-15s %-15s %-20s %-20s\n",
		"İSİM", "DURUM", "POD SAYISI", "OLUŞTURULMA", "LABELS")

	for _, ns := range namespaces.Items {
		// Pod sayısını al
		pods, err := auth.KubeClient.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		podCount := "N/A"
		if err == nil {
			podCount = fmt.Sprintf("%d", len(pods.Items))
		}

		// Label'ları string'e çevir
		labels := []string{}
		for key, value := range ns.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", key, value))
		}
		labelStr := "N/A"
		if len(labels) > 0 {
			labelStr = strings.Join(labels, ",")
		}

		fmt.Printf("%-30s %-15s %-15s %-20s %-20s\n",
			ns.Name,
			string(ns.Status.Phase),
			podCount,
			ns.CreationTimestamp.Format("2006-01-02 15:04:05"),
			labelStr)
	}
}

func listNodes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := auth.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Node listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nNode Listesi:")
	fmt.Printf("\n%-20s %-12s %-15s %-15s %-15s %-15s\n",
		"İSİM", "DURUM", "CPU", "MEMORY", "POD SAYISI", "OS")

	for _, node := range nodes.Items {
		// Node durumunu kontrol et
		status := "NotReady"
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status == "True" {
					status = "Ready"
				}
				break
			}
		}

		// Pod sayısını al
		fieldSelector := fmt.Sprintf("spec.nodeName=%s", node.Name)
		pods, _ := auth.KubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
		})

		// Kaynak kullanımını hesapla
		allocatableCPU := node.Status.Allocatable.Cpu().String()
		allocatableMemory := node.Status.Allocatable.Memory().String()

		fmt.Printf("%-20s %-12s %-15s %-15s %-15d %-15s\n",
			node.Name,
			status,
			allocatableCPU,
			allocatableMemory,
			len(pods.Items),
			node.Status.NodeInfo.OSImage)

		// Detaylı bilgileri göster
		fmt.Printf("  Kernel Version: %s\n", node.Status.NodeInfo.KernelVersion)
		fmt.Printf("  Container Runtime: %s\n", node.Status.NodeInfo.ContainerRuntimeVersion)
		fmt.Printf("  Kubelet Version: %s\n", node.Status.NodeInfo.KubeletVersion)
		fmt.Printf("  Architecture: %s\n", node.Status.NodeInfo.Architecture)
		fmt.Println()
	}
}

func listPods() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := auth.KubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Pod listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nPod Listesi:")
	fmt.Printf("%-5s %-30s %-15s %-12s %-15s %-15s\n",
		"NO", "İSİM", "NAMESPACE", "DURUM", "NODE", "IP")

	podList := make([]corev1.Pod, 0)
	for i, pod := range pods.Items {
		fmt.Printf("%-5d %-30s %-15s %-12s %-15s %-15s\n",
			i+1,
			pod.Name,
			pod.Namespace,
			string(pod.Status.Phase),
			pod.Spec.NodeName,
			pod.Status.PodIP)
		podList = append(podList, pod)
	}

	fmt.Print("\nPod detayları için pod numarası girin (0 için ana menü): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice > 0 && choice <= len(podList) {
		showPodDetails(podList[choice-1])
	}
}

func showPodDetails(pod corev1.Pod) {
	for {
		fmt.Printf("\n=== Pod Detayları: %s ===\n", pod.Name)
		fmt.Println("1. Genel Bilgiler")
		fmt.Println("2. Container Durumları")
		fmt.Println("3. Son Loglar")
		fmt.Println("4. Canlı Log Takibi")
		fmt.Println("5. Events")
		fmt.Println("6. Önceki Menü")
		fmt.Print("Seçiminiz (1-6): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			showPodGeneralInfo(pod)
		case 2:
			showPodContainerStatuses(pod)
		case 3:
			getPodLogs(pod.Name, pod.Namespace, false)
		case 4:
			getPodLogs(pod.Name, pod.Namespace, true)
		case 5:
			getPodEvents(pod)
		case 6:
			return
		default:
			fmt.Println("Geçersiz seçim!")
		}
	}
}

func showPodGeneralInfo(pod corev1.Pod) {
	fmt.Printf("\nPod Adı: %s\n", pod.Name)
	fmt.Printf("Namespace: %s\n", pod.Namespace)
	fmt.Printf("Node: %s\n", pod.Spec.NodeName)
	fmt.Printf("IP: %s\n", pod.Status.PodIP)
	fmt.Printf("Oluşturulma: %s\n", pod.CreationTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Durum: %s\n", pod.Status.Phase)
	fmt.Printf("QoS Sınıfı: %s\n", pod.Status.QOSClass)

	if len(pod.Spec.Containers) > 0 {
		fmt.Println("\nContainer'lar:")
		for _, container := range pod.Spec.Containers {
			fmt.Printf("- %s (Image: %s)\n", container.Name, container.Image)
			if len(container.Ports) > 0 {
				fmt.Print("  Portlar: ")
				for _, port := range container.Ports {
					fmt.Printf("%d/%s ", port.ContainerPort, port.Protocol)
				}
				fmt.Println()
			}
		}
	}
}

func showPodContainerStatuses(pod corev1.Pod) {
	fmt.Printf("\nContainer Durumları - Pod: %s\n", pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		fmt.Printf("\nContainer: %s\n", status.Name)
		fmt.Printf("Ready: %v\n", status.Ready)
		fmt.Printf("Restart Count: %d\n", status.RestartCount)

		if status.State.Running != nil {
			fmt.Printf("Durum: Running (Başlangıç: %s)\n",
				status.State.Running.StartedAt.Format("2006-01-02 15:04:05"))
		} else if status.State.Waiting != nil {
			fmt.Printf("Durum: Waiting (Sebep: %s)\n", status.State.Waiting.Reason)
		} else if status.State.Terminated != nil {
			fmt.Printf("Durum: Terminated (Sebep: %s, Kod: %d)\n",
				status.State.Terminated.Reason,
				status.State.Terminated.ExitCode)
		}
	}
}

func getPodLogs(podName, namespace string, follow bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	// Pod bilgilerini al
	pod, err := auth.KubeClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Pod bilgileri alınamadı: %v\n", err)
		return
	}

	// Birden fazla container varsa seçim yaptır
	containerName := ""
	if len(pod.Spec.Containers) > 1 {
		fmt.Println("\nContainer Listesi:")
		for i, container := range pod.Spec.Containers {
			fmt.Printf("%d. %s\n", i+1, container.Name)
		}
		fmt.Print("\nContainer seçin (1-" + fmt.Sprint(len(pod.Spec.Containers)) + "): ")
		var choice int
		fmt.Scanf("%d", &choice)

		if choice > 0 && choice <= len(pod.Spec.Containers) {
			containerName = pod.Spec.Containers[choice-1].Name
		} else {
			fmt.Println("Geçersiz seçim!")
			return
		}
	} else if len(pod.Spec.Containers) == 1 {
		containerName = pod.Spec.Containers[0].Name
	}

	podLogOpts := corev1.PodLogOptions{
		Container: containerName,
		Follow:    follow,
		TailLines: utils.Int64(100),
	}

	req := auth.KubeClient.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		fmt.Printf("Log stream açılamadı: %v\n", err)
		return
	}
	defer podLogs.Close()

	if follow {
		fmt.Println("\nCanlı log takibi başladı. Çıkmak için 'q' tuşuna basın...")

		// Kullanıcı inputu için goroutine
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				char, _, err := reader.ReadRune()
				if err != nil {
					continue
				}
				if char == 'q' {
					cancel() // Context'i iptal et
					return
				}
			}
		}()
	}

	scanner := bufio.NewScanner(podLogs)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("\nLog takibi sonlandırıldı.")
			return
		default:
			fmt.Println(scanner.Text())
			if !follow {
				if scanner.Text() == "" {
					break
				}
			}
		}
	}
}

func getPodEvents(pod corev1.Pod) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := auth.KubeClient.CoreV1().Events(pod.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name),
	})
	if err != nil {
		fmt.Printf("Events alınamadı: %v\n", err)
		return
	}

	fmt.Printf("\nPod Events - %s:\n", pod.Name)
	fmt.Printf("%-20s %-12s %-20s %s\n", "ZAMAN", "TİP", "SEBEP", "MESAJ")
	for _, event := range events.Items {
		fmt.Printf("%-20s %-12s %-20s %s\n",
			event.LastTimestamp.Format("2006-01-02 15:04:05"),
			event.Type,
			event.Reason,
			event.Message)
	}
}

func listServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services, err := auth.KubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Service listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nService Listesi:")
	fmt.Printf("%-25s %-15s %-10s %-15s %-15s %-20s\n",
		"İSİM", "NAMESPACE", "TİP", "CLUSTER-IP", "EXTERNAL-IP", "PORTS")

	for _, svc := range services.Items {
		// Port bilgilerini formatla
		ports := []string{}
		for _, port := range svc.Spec.Ports {
			portStr := fmt.Sprintf("%d:%d/%s",
				port.Port,
				port.TargetPort.IntVal,
				port.Protocol)
			ports = append(ports, portStr)
		}

		externalIP := "N/A"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
		}

		fmt.Printf("%-25s %-15s %-10s %-15s %-15s %-20s\n",
			svc.Name,
			svc.Namespace,
			string(svc.Spec.Type),
			svc.Spec.ClusterIP,
			externalIP,
			strings.Join(ports, ","))

		// Selector bilgilerini göster
		if len(svc.Spec.Selector) > 0 {
			fmt.Println("  Selectors:")
			for key, value := range svc.Spec.Selector {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}
}

func listDeployments() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployments, err := auth.KubeClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Deployment listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nDeployment Listesi:")
	fmt.Printf("%-25s %-15s %-10s %-10s %-15s %-15s\n",
		"İSİM", "NAMESPACE", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")

	for _, deploy := range deployments.Items {
		age := time.Since(deploy.CreationTimestamp.Time).Round(time.Second)

		fmt.Printf("%-25s %-15s %d/%d %-10d %-15d %-15s\n",
			deploy.Name,
			deploy.Namespace,
			deploy.Status.ReadyReplicas,
			deploy.Status.Replicas,
			deploy.Status.UpdatedReplicas,
			deploy.Status.AvailableReplicas,
			age.String())

		// Detaylı bilgileri göster
		fmt.Printf("  Strategy: %s\n", deploy.Spec.Strategy.Type)
		if len(deploy.Spec.Template.Spec.Containers) > 0 {
			fmt.Println("  Containers:")
			for _, container := range deploy.Spec.Template.Spec.Containers {
				fmt.Printf("    - %s:\n", container.Name)
				fmt.Printf("      Image: %s\n", container.Image)
				if len(container.Resources.Requests) > 0 {
					fmt.Printf("      Requests: CPU=%s, Memory=%s\n",
						container.Resources.Requests.Cpu().String(),
						container.Resources.Requests.Memory().String())
				}
				if len(container.Resources.Limits) > 0 {
					fmt.Printf("      Limits: CPU=%s, Memory=%s\n",
						container.Resources.Limits.Cpu().String(),
						container.Resources.Limits.Memory().String())
				}
			}
		}
		fmt.Println()
	}
}

func listConfigMaps() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configmaps, err := auth.KubeClient.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("ConfigMap listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nConfigMap Listesi:")
	fmt.Printf("%-30s %-20s %-10s %-20s\n", "İSİM", "NAMESPACE", "DATA", "AGE")

	for _, cm := range configmaps.Items {
		age := time.Since(cm.CreationTimestamp.Time).Round(time.Second)
		fmt.Printf("%-30s %-20s %-10d %-20s\n",
			cm.Name,
			cm.Namespace,
			len(cm.Data),
			age.String())

		// ConfigMap içeriğini göster
		if len(cm.Data) > 0 {
			fmt.Println("  Data Keys:")
			for key := range cm.Data {
				fmt.Printf("    - %s\n", key)
			}
		}
		fmt.Println()
	}
}

func listSecrets() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secrets, err := auth.KubeClient.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Secret listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nSecret Listesi:")
	fmt.Printf("%-30s %-20s %-15s %-10s %-20s\n", "İSİM", "NAMESPACE", "TYPE", "DATA", "AGE")

	for _, secret := range secrets.Items {
		age := time.Since(secret.CreationTimestamp.Time).Round(time.Second)
		fmt.Printf("%-30s %-20s %-15s %-10d %-20s\n",
			secret.Name,
			secret.Namespace,
			secret.Type,
			len(secret.Data),
			age.String())

		// Secret key'lerini göster (değerleri göstermeden)
		if len(secret.Data) > 0 {
			fmt.Println("  Data Keys:")
			for key := range secret.Data {
				fmt.Printf("    - %s\n", key)
			}
		}
		fmt.Println()
	}
}

func listPersistentVolumes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pvs, err := auth.KubeClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("PersistentVolume listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nPersistentVolume Listesi:")
	fmt.Printf("%-30s %-15s %-15s %-15s %-15s\n", "İSİM", "CAPACITY", "ACCESS MODES", "STATUS", "CLAIM")

	for _, pv := range pvs.Items {
		claim := "N/A"
		if pv.Spec.ClaimRef != nil {
			claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		fmt.Printf("%-30s %-15s %-15s %-15s %-15s\n",
			pv.Name,
			pv.Spec.Capacity.Storage().String(),
			accessModesToString(pv.Spec.AccessModes),
			string(pv.Status.Phase),
			claim)

		// Storage class ve diğer detayları göster
		fmt.Printf("  StorageClass: %s\n", pv.Spec.StorageClassName)
		fmt.Printf("  Reclaim Policy: %s\n", pv.Spec.PersistentVolumeReclaimPolicy)
		fmt.Println()
	}
}

func listPersistentVolumeClaims() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pvcs, err := auth.KubeClient.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("PersistentVolumeClaim listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nPersistentVolumeClaim Listesi:")
	fmt.Printf("%-30s %-20s %-15s %-15s %-15s\n", "İSİM", "NAMESPACE", "STATUS", "VOLUME", "CAPACITY")

	for _, pvc := range pvcs.Items {
		capacity := "N/A"
		if pvc.Status.Capacity != nil {
			capacity = pvc.Status.Capacity.Storage().String()
		}

		fmt.Printf("%-30s %-20s %-15s %-15s %-15s\n",
			pvc.Name,
			pvc.Namespace,
			string(pvc.Status.Phase),
			pvc.Spec.VolumeName,
			capacity)

		fmt.Printf("  StorageClass: %s\n", *pvc.Spec.StorageClassName)
		fmt.Printf("  Access Modes: %s\n", accessModesToString(pvc.Spec.AccessModes))
		fmt.Println()
	}
}

func listStatefulSets() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statefulsets, err := auth.KubeClient.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("StatefulSet listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nStatefulSet Listesi:")
	fmt.Printf("%-30s %-20s %-10s %-15s %-15s\n", "İSİM", "NAMESPACE", "READY", "AGE", "SERVICE NAME")

	for _, sts := range statefulsets.Items {
		age := time.Since(sts.CreationTimestamp.Time).Round(time.Second)
		fmt.Printf("%-30s %-20s %d/%d %-15s %-15s\n",
			sts.Name,
			sts.Namespace,
			sts.Status.ReadyReplicas,
			sts.Status.Replicas,
			age.String(),
			sts.Spec.ServiceName)

		// Volume claim templates
		if len(sts.Spec.VolumeClaimTemplates) > 0 {
			fmt.Println("  Volume Claim Templates:")
			for _, vct := range sts.Spec.VolumeClaimTemplates {
				fmt.Printf("    - %s (%s)\n", vct.Name, vct.Spec.Resources.Requests.Storage().String())
			}
		}
		fmt.Println()
	}
}

func listDaemonSets() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	daemonsets, err := auth.KubeClient.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("DaemonSet listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nDaemonSet Listesi:")
	fmt.Printf("%-30s %-20s %-15s %-15s %-15s\n", "İSİM", "NAMESPACE", "DESIRED", "CURRENT", "READY")

	for _, ds := range daemonsets.Items {
		fmt.Printf("%-30s %-20s %-15d %-15d %-15d\n",
			ds.Name,
			ds.Namespace,
			ds.Status.DesiredNumberScheduled,
			ds.Status.CurrentNumberScheduled,
			ds.Status.NumberReady)

		// Node selector bilgilerini göster
		if len(ds.Spec.Template.Spec.NodeSelector) > 0 {
			fmt.Println("  Node Selectors:")
			for key, value := range ds.Spec.Template.Spec.NodeSelector {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}
}

func listIngresses() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ingresses, err := auth.KubeClient.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Ingress listesi alınamadı: %v\n", err)
		return
	}

	fmt.Println("\nIngress Listesi:")
	fmt.Printf("%-30s %-20s %-20s %-30s\n", "İSİM", "NAMESPACE", "CLASS", "HOSTS")

	for _, ing := range ingresses.Items {
		ingressClass := "N/A"
		if ing.Spec.IngressClassName != nil {
			ingressClass = *ing.Spec.IngressClassName
		}

		hosts := []string{}
		for _, rule := range ing.Spec.Rules {
			hosts = append(hosts, rule.Host)
		}

		fmt.Printf("%-30s %-20s %-20s %-30s\n",
			ing.Name,
			ing.Namespace,
			ingressClass,
			strings.Join(hosts, ","))

		// TLS ve Path bilgilerini göster
		if len(ing.Spec.TLS) > 0 {
			fmt.Println("  TLS:")
			for _, tls := range ing.Spec.TLS {
				fmt.Printf("    - Secret Name: %s\n", tls.SecretName)
				fmt.Printf("      Hosts: %s\n", strings.Join(tls.Hosts, ", "))
			}
		}

		fmt.Println("  Rules:")
		for _, rule := range ing.Spec.Rules {
			fmt.Printf("    - Host: %s\n", rule.Host)
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					fmt.Printf("      Path: %s -> %s:%d\n",
						path.Path,
						path.Backend.Service.Name,
						path.Backend.Service.Port.Number)
				}
			}
		}
		fmt.Println()
	}
}

// Yardımcı fonksiyon
func accessModesToString(modes []corev1.PersistentVolumeAccessMode) string {
	strs := make([]string, len(modes))
	for i, mode := range modes {
		strs[i] = string(mode)
	}
	return strings.Join(strs, ",")
}
