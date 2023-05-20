/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	routeset "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/rodaine/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Structs for alerts
type AlertResponse struct {
	Status string `json:"status"`
	Data   []struct {
		Labels *AlertLabels
		Status *AlertStatus
	} `json:"data"`
}
type AlertLabels struct {
	Alertname string `json:"alertname"`
	Namespace string `json:"namespace"`
	Severity  string `json:"severity"`
}
type AlertStatus struct {
	State string `json:"state"`
}

// Function to print all current firing alerts
func alertsStatus(config *rest.Config) {
	fmt.Print(color.New(color.Bold).Sprintln("Checking alerts..."))

	// Create clientset to get Alertmanager route
	clientset, err := routeset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	route, err := clientset.RouteV1().Routes("openshift-monitoring").Get(context.TODO(), "alertmanager-main", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Define URL for alerts endpoint
	alertmanagerURL := "https://" + route.Spec.Host + "/api/v1/alerts"

	// Get user token
	cmdOut, err := exec.Command("oc", "whoami", "-t").Output()
	if err != nil {
		panic(err.Error())
	}
	bearerToken := strings.TrimSuffix(string(cmdOut), "\n")

	// Request all current alerts
	req, err := http.NewRequest("GET", alertmanagerURL, nil)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	// Set body variable with server response
	var alerts AlertResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		panic(err.Error())
	}

	// Create a new table for printing alerts
	table := table.New("  ALERTNAME", "NAMESPACE", "SEVERITY", "STATE").WithPadding(5)

	// Print output
	if len(alerts.Data) > 0 {
		fmt.Printf("  %s Found Alerts in firing state\n", color.RedString("[Warning]"))
		fmt.Printf("")
		for _, value := range alerts.Data {
			name := value.Labels.Alertname
			state := value.Status.State
			namespace := value.Labels.Namespace
			severity := value.Labels.Severity
			table.AddRow("  "+name, namespace, severity, state)
		}
		table.Print()
	} else {
		fmt.Printf("  %s There is no Alerts in AlertManager in firing state at this time\n", color.YellowString("[Info]"))
	}
	fmt.Println()
}
