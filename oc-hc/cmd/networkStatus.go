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
func networkStatus() error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking network..."))

	err := checkDNS()
	if err != nil {
		return err
	}

	err = checkEgress()
	if err != nil {
		return err
	}

	return nil
}

func checkEgress() error {
	// Run pod to test egress connectivity
	fmt.Println(" - Checking Egress conectivity to the internet...")

	// Make sure the egress-tester pod doesn't exist and clean up variable
	cleanup, _ := exec.Command("oc", "delete", "pod", "egress-tester", "-n", "openshift-monitoring").Output()
	_ = cleanup

	// run the egress tester
	cmd, err := exec.Command("oc", "run", "-i", "--rm=true", "egress-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "curl https://www.redhat.com &> /dev/null && echo OK").Output()
	if err != nil {
		return err
	}

	// Print output
	if strings.Contains(string(cmd), "OK") {
		fmt.Printf("  %s There is internet connectivity to www.redhat.com\n", color.YellowString("[Info]"))
	} else {
		fmt.Printf("  %s There is no internet connectivity to www.redhat.com\n", color.RedString("[Warning]"))
	}
	fmt.Println()
	return nil
}

func checkDNS() error {
	// Run pod to test DNS resolution
	fmt.Println(" - Checking if DNS can resolve external names...")

	// Make sure the egress-tester pod doesn't exist and clean up variable
	cleanup, _ := exec.Command("oc", "delete", "pod", "dns-tester", "-n", "openshift-monitoring").Output()
	_ = cleanup

	cmddns, err := exec.Command("oc", "run", "-i", "--rm=true", "dns-tester", "-n", "openshift-monitoring", "--image", "registry.redhat.io/openshift4/network-tools-rhel8", "--", "/bin/bash", "-c", "sleep 3 && dig +short www.redhat.com &> /dev/null && echo OK").Output()
	if err != nil {
		return err
	}

	// Print output
	if strings.Contains(string(cmddns), "OK") {
		fmt.Printf("  %s DNS can resolve www.redhat.com\n", color.YellowString("[Info]"))
	} else {
		fmt.Printf("  %s DNS can not resolve www.redhat.com\n", color.RedString("[Warning]"))
	}
	fmt.Println()
	return nil
}
