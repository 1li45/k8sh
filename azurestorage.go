package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func listStorageObjects(client *kubernetes.Clientset) error {
	// List Persistent Volumes
	persistentVolumes, err := client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("could not list Persistent Volumes: %v", err)
	}

	fmt.Println("Persistent Volumes in AKS:")
	for _, pv := range persistentVolumes.Items {
		if strings.HasPrefix(pv.Name, "pvc-") {

			fmt.Printf("Name: %s, Capacity: %v, Status: %s\n", pv.Name, pv.Spec.Capacity.Storage().Value(), pv.Status.Phase)
		}
	}

	fmt.Println()

	// Wait for user input to select a persistent volume
	var selection int
	fmt.Print("View 12-hour R/W Data of Persistent Volume: ")
	fmt.Scanln(&selection)

	// Call getAzureDiskResourceGroup with the selected persistent volume name

	persistentVolume := persistentVolumes.Items[selection-1]
	azureName := "kubernetes-dynamic-" + persistentVolume.Name
	resourceGroup := getAzureDiskResourceGroup(azureName)
	resourceID := getAzureDiskID(azureName, resourceGroup)

	fmt.Println(getAzureDiskMetrics(resourceID))

	return nil
}

func getAzureDiskResourceGroup(diskName string) string {
	cmd := exec.Command("az", "disk", "list",
		"--query", fmt.Sprintf("[?name=='%s'].resourceGroup | [0]", diskName),
		"-o", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	resourceGroup := strings.TrimSpace(string(output))
	return resourceGroup
}

func getAzureDiskID(diskName, resourceGroup string) string {
	cmd := exec.Command("az", "disk", "show",
		"--name", diskName,
		"--resource-group", resourceGroup,
		"--query", "id",
		"-o", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	diskID := strings.TrimSpace(string(output))
	return diskID
}

func getAzureDiskMetrics(resourceID string) string {
	startTime := time.Now().UTC().Add(-12 * time.Hour).Format("2006-01-02T15:04:00Z")
	endTime := time.Now().UTC().Format("2006-01-02T15:04:00Z")

	cmd := exec.Command("az", "monitor", "metrics", "list",
		"--resource", resourceID,
		"--metric", "Composite Disk Read Operations/sec,Composite Disk Write Operations/sec",
		"--interval", "PT1H",
		"--start-time", startTime,
		"--end-time", endTime,
		"-o", "table")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	metrics := strings.TrimSpace(string(output))
	return metrics
}
