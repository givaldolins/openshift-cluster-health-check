/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// Structs for machineConfigPool
type mcpResponse struct {
	Items []struct {
		Status   *mcpStatus   `json:"status"`
		Metadata *mcpMetadata `json:"metadata"`
	} `json:"items"`
}
type mcpMetadata struct {
	Name string `json:"name"`
}
type mcpConditions struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
type mcpStatus struct {
	Conditions           []*mcpConditions `json:"conditions"`
	DegradedMachineCount int              `json:"degradedMachineCount"`
	MachineCount         int              `json:"machineCount"`
	ReadyMachineCount    int              `json:"readyMachineCount"`
	UpdatedMachineCount  int              `json:"updatedMachineCount"`
}

// Function to check MCP
func machineConfigPoolStatus() error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking MCP..."))
	fmt.Println("  - Checking if MCP is rolling the nodes...")

	// Get MCP json
	cmdOut, err := exec.Command("oc", "get", "mcp", "-ojson").Output()
	if err != nil {
		return err
	}
	data := mcpResponse{}
	err = json.Unmarshal(cmdOut, &data)
	if err != nil {
		return err
	}

	// Create a new table for printing output
	table := table.New("  NAME", "UPDATING", "DEGRADED", "MACHINECOUNT", "READYMACHINECOUNT", "UPDATEDMACHINECOUNT", "DEGRADEDMACHINECOUNT").WithPadding(5)

	// Check MCP status
	warning := false
	updating := "False"
	degraded := "False"
	for _, mcp := range data.Items {
		for _, condition := range mcp.Status.Conditions {
			if condition.Type == "Updating" && condition.Status == "True" {
				warning = true
				updating = "True"
			}

			if condition.Type == "Degraded" || condition.Type == "NodeDegrade" || condition.Type == "RenderDegraded" {
				if condition.Status == "True" {
					warning = true
					degraded = "True"
				}
			}
		}
		table.AddRow("  "+mcp.Metadata.Name, updating, degraded, mcp.Status.MachineCount, mcp.Status.ReadyMachineCount, mcp.Status.UpdatedMachineCount, mcp.Status.DegradedMachineCount)
	}

	// Print output
	if warning {
		fmt.Printf("  %s One or more machineconfigpool may require your attention\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All machineconfigpools look good\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()

	return nil
}
