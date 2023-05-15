/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"path/filepath"

	configset "github.com/openshift/client-go/config/clientset/versioned"
	routeset "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/spf13/cobra"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

type checkOptions struct {
	kubeconfig string
}

type nodemetrics struct {
	name      string
	capCpu    float64
	capMemory float64
}

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

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the overall health for an OpenShift cluster",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		obj := complete(cmd, args)
		run(obj)
	},
}

var (
	containerrestart int32 = 10
)

func init() {

	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")
	checkCmd.PersistentFlags().StringP("kubeconfig", "k", "", "(optional) Path for the kubeconfig file to be used")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func complete(cmd *cobra.Command, args []string) checkOptions {
	// Get kubeconfig flag
	kube, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		panic(err.Error())
	}
	// Use default kubeconfig if not passed via flag
	if kube == "" {
		kube = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		fmt.Printf("%s[Info]%s Using default kubeconfig: %s\n", colorYellow, colorReset, kube)

		// Check if file exists
	} else if _, err := os.Stat(kube); err != nil {
		panic(err.Error())
	}

	// Instantiate a checkOptions object
	obj := checkOptions{
		kubeconfig: kube,
	}

	return obj
}

func run(obj checkOptions) {

	// Build a new config from flag kubeconfig and instantiate a new clientset
	config, err := clientcmd.BuildConfigFromFlags("", obj.kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	clusterAdmin(clientset)
	capacityStatus(clientset, config)
	nodeStatus(clientset)
	etcdStatus(clientset)
	podStatus(clientset, containerrestart)
	eventStatus(clientset)
	alertsStatus(config)
	csrStatus(clientset)

	apiStatus(clientset)
	coStatus(config)
	mcpStatus(clientset)
	networkStatus(clientset)
	versionStatus(clientset)

}

func coStatus(config *rest.Config) {
	fmt.Println("Checking cluster Operators...")
	clientset, err := configset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clusteroperators, err := clientset.ConfigV1().ClusterOperators().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	counter := 0
	for _, co := range clusteroperators.Items {
		for _, condition := range co.Status.Conditions {
			if condition.Type == "Degraded" && condition.Status == "True" {
				fmt.Printf("  %s[Warning]%s clusteroperator %s is Degraded\n", colorRed, colorReset, co.GetName())
				counter++
			}
			if condition.Type == "Available" && condition.Status == "False" {
				fmt.Printf("  %s[Warning]%s clusteroperator %s is not available\n", colorRed, colorReset, co.GetName())
				counter++
			}
			if condition.Type == "Progressing" && condition.Status == "True" {
				fmt.Printf("  %s[Info]%s clusteroperator %s is in Progressing state\n", colorYellow, colorReset, co.GetName())
				counter++
			}
		}
	}
	if counter == 0 {
		fmt.Printf("  %s[Info]%s All clusteroperator are healthy\n", colorYellow, colorReset)
	}

}

func mcpStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking MCP...")
	// OpenShift API
}

func pingverifier() string {
	return `
	#!/bin/bash
	ping 8.8.8.8 -c 5 &>/dev/null && echo 0 || echo 1
	`
}

func networkStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking network...")

	fmt.Println(" - Checking Egress conectivity...")

	//cmd, err := exec.Command("oc", "run", "-i", "--rm=true", "network-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "curl", "--keepalive-time", "3", "-w", "%{exitcode}", "telnet://8.8.8.8:53").Output()
	cmd, err := exec.Command("oc", "run", "-i", "--rm=true", "network-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "ping 8.8.8.8 -c 3 &> /dev/null && echo OK").Output()
	if err != nil {
		fmt.Printf(err.Error())
	}

	if strings.Contains(string(cmd), "OK") {
		fmt.Printf("  %s[Info]%s There is internet connectivity\n", colorYellow, colorReset)
	} else {
		fmt.Printf("  %s[Warning]%s There is no internet connectivity\n", colorRed, colorReset)
	}

	fmt.Println(" - Checking DNS can resolve...")

	//cmd, err := exec.Command("oc", "run", "-i", "--rm=true", "network-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "curl", "--keepalive-time", "3", "-w", "%{exitcode}", "telnet://8.8.8.8:53").Output()
	cmddns, err := exec.Command("oc", "run", "-i", "--rm=true", "dns-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "sleep 3 && dig +short www.redhat.com &> /dev/null && echo OK").Output()
	if err != nil {
		fmt.Printf(err.Error())
	}

	if strings.Contains(string(cmddns), "OK") {
		fmt.Printf("  %s[Info]%s DNS can resolve www.redhat.com\n", colorYellow, colorReset)
	} else {
		fmt.Printf("  %s[Warning]%s DNS can not resolve www.redhat.com\n", colorRed, colorReset)
	}

	// Check DNS
	// Check certificate expiration
}

func versionStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking cluster version...")
	// check if cluster is EOL
	// Openshift API
}

func capacityStatus(clientset *kubernetes.Clientset, config *rest.Config) {
	fmt.Println("Checking capacity...")
	nodes := allocatableResources(clientset)
	currentUtilization(clientset, config, nodes)
}

func allocatableResources(clientset *kubernetes.Clientset) []nodemetrics {
	fmt.Println(" - Checking allocated resources...")
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 1, '\t', 0)
	defer table.Flush()

	rNode := []nodemetrics{}

	for _, node := range nodeList.Items {
		capCpu := node.Status.Capacity.Cpu().AsApproximateFloat64()
		capMem := node.Status.Capacity.Memory().AsApproximateFloat64()
		allocCpu := node.Status.Allocatable.Cpu().AsApproximateFloat64()
		allocMem := node.Status.Allocatable.Memory().AsApproximateFloat64()
		percentCpu := 100 - (allocCpu * 100 / capCpu)
		percentMem := 100 - (allocMem * 100 / capMem)
		rNode = append(rNode, nodemetrics{capCpu: capCpu, capMemory: capMem, name: node.GetName()})

		if percentCpu >= 80 || percentMem >= 80 {
			fmt.Fprintf(table, "  %s[Warning]%s Node %s\t CPU: %.0f%%\t Memory: %.0f%%\n", colorRed, colorReset, node.GetName(), percentCpu, percentMem)
		} else {
			fmt.Fprintf(table, "  %s[Info]%s Node %s\t CPU: %.0f%%\t Memory: %.0f%%\n", colorYellow, colorReset, node.GetName(), percentCpu, percentMem)
		}

	}
	return rNode
}

func currentUtilization(clientset *kubernetes.Clientset, config *rest.Config, nodeList []nodemetrics) {
	fmt.Println(" - Checking current resources use...")

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 1, '\t', 0)
	defer table.Flush()

	metricClientSet, err := metricsv1beta.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodesUtilization, err := metricClientSet.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodesUtilization.Items {
		cpuUtilization := node.Usage.Cpu().AsApproximateFloat64()
		memoryUtilization := node.Usage.Memory().AsApproximateFloat64()
		for _, item := range nodeList {
			if node.Name == item.name {
				percentCPU := cpuUtilization * 100 / item.capCpu
				percentMemory := memoryUtilization * 100 / item.capMemory
				if percentCPU >= 80 || percentMemory >= 80 {
					fmt.Fprintf(table, "  %s[Warning]%s Node %s\t CPU: %.0f%%\t Memory: %.0f%%\n", colorRed, colorReset, node.GetName(), percentCPU, percentMemory)
				} else {
					fmt.Fprintf(table, "  %s[Info]%s Node %s\t CPU: %.0f%%\t Memory: %.0f%%\n", colorYellow, colorReset, node.GetName(), percentCPU, percentMemory)
				}
			}
		}
	}
}

func nodeStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking nodes status...")
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	checkTaints(nodes)
	checkConditions(nodes)
	fmt.Println()
}

func checkConditions(nodes *corev1.NodeList) {

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 1, '\t', 0)
	defer table.Flush()

	fmt.Print("\n - Checking node conditions...")
	for _, node := range nodes.Items {
		nodeName := node.GetName()
		for i, condition := range node.Status.Conditions {
			kind := condition.Type
			status := condition.Status
			if (kind == "Ready" && status != "True") || (kind != "Ready" && status == "True") {
				if i == 0 {
					fmt.Fprintf(table, "\n  %s[Warning]%s Node %s\t", colorRed, colorReset, nodeName)
				}
				fmt.Fprintf(table, "%s: %s\t", kind, status)
			} else {
				if i == 0 {
					fmt.Fprintf(table, "\n  %s[Info]%s Node %s\t", colorYellow, colorReset, nodeName)
				}
				fmt.Fprintf(table, "%s: %s\t", kind, status)
			}
		}
	}
}

func checkTaints(nodes *corev1.NodeList) {

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 1, '\t', 0)
	defer table.Flush()

	fmt.Print(" - Checking for node taints...")

	for _, node := range nodes.Items {
		nodeName := node.GetName()
		for i, taint := range node.Spec.Taints {
			if taint.Effect == "NoSchedule" && taint.Key == "node.kubernetes.io/unschedulable" {
				if i == 0 {
					fmt.Fprintf(table, "\n  %s[Warning]%s Node %s\t", colorRed, colorReset, nodeName)
				}
				fmt.Fprintf(table, "%s: %s", taint.Key, taint.Effect)
			} else {
				if i == 0 {
					fmt.Fprintf(table, "\n  %s[Info]%s Node %s\t", colorYellow, colorReset, nodeName)
				}
				fmt.Fprintf(table, "%s: %s", taint.Key, taint.Effect)
			}
		}
	}
}

func etcdStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking ETCD...")
	etcds, err := clientset.CoreV1().ComponentStatuses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 2, '\t', 0)
	defer table.Flush()

	for _, etcd := range etcds.Items {
		if strings.Contains(etcd.Name, "etcd") {
			for _, condition := range etcd.Conditions {
				if condition.Type == "Healthy" && condition.Status != "True" {
					fmt.Fprintf(table, "  %s[Warning]%s %s\t %s: %s\n", colorRed, colorReset, etcd.Name, condition.Type, condition.Status)
				} else {
					fmt.Fprintf(table, "  %s[Info]%s %s\t  %s: %s\n", colorYellow, colorReset, etcd.Name, condition.Type, condition.Status)
				}
			}
		}
	}
}

func alertsStatus(config *rest.Config) {
	fmt.Println("Checking alerts...")

	clientset, err := routeset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	route, err := clientset.RouteV1().Routes("openshift-monitoring").Get(context.TODO(), "alertmanager-main", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	//// HERE Get url from route
	alertmanagerURL := "https://" + route.Spec.Host + "/api/v1/alerts"

	cmdOut, err := exec.Command("oc", "whoami", "-t").Output()
	if err != nil {
		panic(err.Error())
	}
	bearerToken := strings.TrimSuffix(string(cmdOut), "\n")

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

	var alerts AlertResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		panic(err.Error())
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 6, 1, '\t', 0)
	defer table.Flush()

	if len(alerts.Data) > 0 {
		fmt.Printf("  %s[Warning]%s Found Alerts in firing state\n", colorRed, colorReset)
		for _, value := range alerts.Data {
			name := value.Labels.Alertname
			state := value.Status.State
			namespace := value.Labels.Namespace
			severity := value.Labels.Severity
			fmt.Fprintf(table, "    - Name: %v\t Namespace: %v\t Severity: %v\t State: %v\n", name, namespace, severity, state)
		}
	} else {
		fmt.Printf("  %s[Info]%s There is no Alerts in AlertManager in firing state at this time\n", colorYellow, colorReset)
	}
}

func podStatus(clientset *kubernetes.Clientset, restartNumber int32) {
	fmt.Println("Checking pods status...")
	// Get a list of pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	counter := 0
	for _, pod := range pods.Items {
		// Print pods that are not running or succeeded
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			fmt.Printf("  %s[Warning]%s The status of Pod %s in namespace %s is %s\n", colorRed, colorReset, pod.GetName(), pod.GetNamespace(), pod.Status.Phase)
			counter++
		}
		// Print pods that have restarted more than a given number
		for _, container := range pod.Status.ContainerStatuses {
			if container.RestartCount > restartNumber {
				fmt.Printf("  %s[Warning]%s Container %s in Pod %s in namespace %s as restarted more than %d times\n", colorRed, colorReset, container.Name, pod.GetName(), pod.GetNamespace(), restartNumber)
				counter++
			}
		}
	}
	if counter == 0 {
		fmt.Printf("  %s[Info]%s There is no pod in failed state or restarted more than %d times\n", colorYellow, colorReset, restartNumber)
	}
}

func eventStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking events...")
	events, err := clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 1, 5, 1, '\t', 0)
	defer table.Flush()

	counter := 0
	for _, event := range events.Items {
		if event.Type == "Warning" {
			counter++
			if event.InvolvedObject.Namespace == "" {
				if len(event.Message) < 80 {
					fmt.Fprintf(table, "  %s[Warning]%s Namespace:%s\t\t Event: %s\t Message%s\n", colorRed, colorReset, "", event.Reason, event.Message)
				} else {
					fmt.Fprintf(table, "  %s[Warning]%s Namespace:%s\t\t Event: %s\t Message%s\n", colorRed, colorReset, "", event.Reason, event.Message[:80]+"...")
				}
			} else {
				if len(event.Message) < 80 {
					fmt.Fprintf(table, "  %s[Warning]%s Namespace: %s\t\t Event: %s\t Message:%s\n", colorRed, colorReset, event.InvolvedObject.Namespace, event.Reason, event.Message)
				} else {
					fmt.Fprintf(table, "  %s[Warning]%s Namespace: %s\t\t Event: %s\t Message:%s\n", colorRed, colorReset, event.InvolvedObject.Namespace, event.Reason, event.Message[:80]+"...")
				}
			}
		}
	}
	if counter == 0 {
		fmt.Fprintf(table, "  %s[Info]%s There is no warning events at this time\n", colorYellow, colorReset)
	}

}

func csrStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking CSRs status...")
	// Get a list of CSRs
	csrs, err := clientset.CertificatesV1().CertificateSigningRequests().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	counter := 0

csrLoop:
	for _, csr := range csrs.Items {
		if len(csr.Status.Conditions) == 0 {
			fmt.Printf("  %s[Warning]%s CSR %s is not in Approved state\n", colorRed, colorReset, csr.GetName())
			counter++
			continue csrLoop
		}
		for _, condition := range csr.Status.Conditions {
			//fmt.Printf("%v", condition)
			if !(condition.Type == "Approved" && condition.Status == "True") {
				fmt.Printf("  %s[Warning]%s CSR %s is not in Approved state\n", colorRed, colorReset, csr.GetName())
				counter++
			}
		}
	}
	if counter == 0 {
		fmt.Printf("  %s[Info]%s There is no pending CSR at this time\n", colorYellow, colorReset)
	}
}

func clusterAdmin(clientset *kubernetes.Clientset) {
	review := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:     "*",
				Group:    "*",
				Resource: "*",
			},
		},
	}
	auth, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &review, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	if !auth.Status.Allowed {
		fmt.Printf("%s[Error]%s User is not a cluster-admin. %s\n", colorRed, colorReset, auth.Status.Reason)
		panic("This needs to be run by a user with cluster-admin privileges")
	}
}

func apiStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking API...")
	apipods, err := clientset.CoreV1().Pods("openshift-apiserver").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=openshift-apiserver-a"})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(" - Checking OpenShift API server pods readiness...")
	for _, apipod := range apipods.Items {
		cmd, err := exec.Command("oc", "exec", "-it", apipod.GetName(), "-n", "openshift-apiserver", "-c", "openshift-apiserver", "--", "curl", "-k", "https://localhost:8443/readyz").Output()
		if err != nil {
			panic(err.Error())
		}
		if stdout := string(cmd); stdout != "ok" {
			fmt.Printf("  %s[Warning]%s Pod %s is not ready\n", colorRed, colorReset, apipod.GetName())
		} else {
			fmt.Printf("  %s[Info]%s Pod %s is ready\n", colorYellow, colorReset, apipod.GetName())
		}
	}
	kubepods, err := clientset.CoreV1().Pods("openshift-kube-apiserver").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=openshift-kube-apiserver"})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(" - Checking OpenShift Kube API server pods readiness...")
	for _, kubepod := range kubepods.Items {
		cmd, err := exec.Command("oc", "exec", "-it", kubepod.GetName(), "-n", "openshift-kube-apiserver", "-c", "kube-apiserver", "--", "curl", "-k", "https://localhost:6443/readyz").Output()

		if err != nil {
			panic(err.Error())
		}
		if stdout := string(cmd); stdout != "ok" {
			fmt.Printf("  %s[Warning]%s Pod %s is not ready\n", colorRed, colorReset, kubepod.GetName())
		} else {
			fmt.Printf("  %s[Info]%s Pod %s is ready\n", colorYellow, colorReset, kubepod.GetName())
		}
	}
	// Deprecated APIs are not eing used TBD
}
