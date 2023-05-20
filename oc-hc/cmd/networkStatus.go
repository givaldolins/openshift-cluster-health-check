/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// Wrapper function
func networkStatus() {
	fmt.Print(color.New(color.Bold).Sprintln("Checking network..."))

	checkEgress()

	checkDns()

}

func checkEgress() {
	// Run pod to test egress connectivity
	fmt.Println(" - Checking Egress conectivity...")
	cmd, err := exec.Command("oc", "run", "-i", "--rm=true", "network-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "ping 8.8.8.8 -c 3 &> /dev/null && echo OK").Output()
	if err != nil {
		panic(err.Error())
	}

	// Print output
	if strings.Contains(string(cmd), "OK") {
		fmt.Printf("  %s There is internet connectivity\n", color.YellowString("[Info]"))
	} else {
		fmt.Printf("  %s There is no internet connectivity\n", color.RedString("[Warning]"))
	}
	fmt.Println()
}

func checkDns() {
	// Run pod to test DNS resolution
	fmt.Println(" - Checking DNS can resolve...")
	cmddns, err := exec.Command("oc", "run", "-i", "--rm=true", "dns-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "sleep 3 && dig +short www.redhat.com &> /dev/null && echo OK").Output()
	if err != nil {
		panic(err.Error())
	}

	// Print output
	if strings.Contains(string(cmddns), "OK") {
		fmt.Printf("  %s DNS can resolve www.redhat.com\n", color.YellowString("[Info]"))
	} else {
		fmt.Printf("  %s DNS can not resolve www.redhat.com\n", color.RedString("[Warning]"))
	}
	fmt.Println()
}
