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
	"strconv"
	"strings"

	"github.com/fatih/color"
	configset "github.com/openshift/client-go/config/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Struct used for the clusterversion object
type versionResponse struct {
	Nodes []struct {
		Version string `json:"version"`
	} `json:"nodes"`
}

// Check if cluster is EOL
func versionStatus(config *rest.Config) {
	fmt.Print(color.New(color.Bold).Sprintln("Checking if cluster is EOL..."))

	// Create client set to interact with API
	clientset, err := configset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Get cluster version object
	clusterversion, err := clientset.ConfigV1().ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Variables to be used when checking the version
	currentChannel, _ := strconv.ParseFloat(strings.ReplaceAll(clusterversion.Spec.Channel, "stable-", ""), 64)
	var latestChannel float64
	openshiftApi := "https://api.openshift.com/api/upgrades_info/v1/graph?channel=stable-"

	// Determine lastest channel available
	var vResponse versionResponse
	for i := 0.01; i < 0.99; i = i + 0.01 {
		nextChannel := currentChannel + i
		apiUrl := openshiftApi + fmt.Sprintf("%.2f", nextChannel)
		resp, err := http.Get(apiUrl)
		if err != nil {
			panic(err.Error())
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal(body, &vResponse)
		if err != nil {
			panic(err.Error())
		}
		if len(vResponse.Nodes) == 0 {
			latestChannel = nextChannel - 0.01
			break
		}
	}

	// Print output
	if (latestChannel - currentChannel) >= 0.03 {
		fmt.Printf("  %s This cluster version (%v) might be out of support. Please double check the OpenShift Lifecycle page to confirm that.\n", color.RedString("[Warning]"), clusterversion.Spec.DesiredUpdate.Version)
	} else {
		fmt.Printf("  %s This cluster version (%v) is not more than 2 versions behind.\n", color.YellowString("[Info]"), clusterversion.Spec.DesiredUpdate.Version)
	}
	fmt.Println()

}
