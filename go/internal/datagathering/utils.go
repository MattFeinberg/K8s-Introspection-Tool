package datagathering

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"strconv"
	"strings"
	"time"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// HandleDataUpdates scrapes all cluster data
// Break this down into more functions
// TODO: I dont like this functions name
func HandleDataUpdates() ClusterInfo {
	var cluster ClusterInfo
	clientset, config, err := GetK8s()
	if err != nil {
		fmt.Println("Error getting K8s info", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}

	// update nodes
	cluster.Nodes, err = updateNodes(clientset, config)
	if err != nil {
		fmt.Println("Error updating nodes\n", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}

    //TODO: Put the below in a function (to get cluster level data)

	//get K8s distribution
	cluster.Distribution, err = getDistribution(clientset, config)
	if err != nil {
		fmt.Println("Error getting distribution\n", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}

	//get cloud info
	cluster.CloudOrOnPrem, cluster.CSP, err = getCloudInfo(clientset, config)
	if err != nil {
		fmt.Println("Error getting cloud info\n", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}

	// get kubernetes version
	discovClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		fmt.Println("Error getting discovery client\n", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}
	k8sVersion, err := discovClient.ServerVersion()
	if err != nil {
		fmt.Println("Error getting server version (GitVersion)\n", err)
		//TODO: dont reutrn cluster in error case
		return cluster
	}
	cluster.K8sVersion = k8sVersion.GitVersion

	//TODO: combine range based for loops over nodes array
	//Count GPUs
	totalGPUs := 0
	totalUnhealthy := 0
	for _, node := range cluster.Nodes {
		totalGPUs = totalGPUs + node.AttachedGPUs
		if node.NVSMIExitCode != 0 {
			totalUnhealthy = totalUnhealthy + 1
		}
	}
	cluster.TotalGPUs = totalGPUs
	cluster.NumUnhealthyNodes = totalUnhealthy

	//Count NICs
	totalNICs := 0
	for _, node := range cluster.Nodes {
		totalNICs = totalNICs + node.AttachedNICs
	}
	cluster.TotalNICs = totalNICs

	//Count GPU Type Distribution
	cluster.GPUDist = make(map[string]int)
	for _, node := range cluster.Nodes {
		for _, GPU := range node.GPUs {
			if _, found := cluster.GPUDist[GPU.ProductName]; found {
				// type already in map, increment it
				cluster.GPUDist[GPU.ProductName] = cluster.GPUDist[GPU.ProductName] + 1
			} else {
				// type not in mat, set it to 1
				cluster.GPUDist[GPU.ProductName] = 1
			}
		}
	}

	// Count total MIG
	cluster.TotalMIG = 0
	for _, node := range cluster.Nodes {
		for _, GPU := range node.GPUs {
			if GPU.MIGInfo.MIGStatus == "Enabled" {
				cluster.TotalMIG = cluster.TotalMIG + 1
			}
		}
	}

	// TODO: How many GPU are in good/bad state?
	//       idea: Check NVSMI return value maybe?
	// get node names
	var names []string
	for _, elem := range cluster.Nodes {
		names = append(names, elem.NodeName)
	}

	// update rest of cluster info
	cluster.NumNodes = len(cluster.Nodes)
	cluster.NodeNames = names
	cluster.TimeUpdated = time.Now().Format("01-02-2006 15:04:05")

	// write to csv

    path := os.Getenv("FILEPATH")

	fmt.Println("printing")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("failed creating csv file\n", err)
		return cluster
	}
	csvWriter := csv.NewWriter(file)
	str, err := json.MarshalIndent(cluster, "", "    ")
	if err != nil {
		fmt.Print("error reading json\n", err)
		return cluster
	}
	row := []string{cluster.TimeUpdated, string(str)}
	if err := csvWriter.Write(row); err != nil {
		fmt.Print("error writing to file\n", err)
		return cluster
	}
	fmt.Println("DOne")
    csvWriter.Flush()
	file.Close()
	return cluster
}

// GetK8s returns Kubernetse clientest and config for the cluster the app is running on
func GetK8s() (*kubernetes.Clientset, *rest.Config, error) {

	// set up config and clientset for the cluster
	//    rules := clientcmd.NewDefaultClientConfigLoadingRules()
	//    kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	//    config, err := kubeconfig.ClientConfig()
	// the following line only works when running directly on cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("\nError in getting config")
		return nil, nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("\nError in getting access to K8s")
		return nil, nil, err
	}
	return clientset, config, err
}

// get output from nvidia-smi
func runNVSMI(clientset *kubernetes.Clientset, config *rest.Config, nodeName string) (string, int, error) {
	// find the device plugin daemonset pod associated with this node
	var podName string
	var podNamespace string
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(),
		metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		fmt.Println("Error getting pods")
		return "", 1, err
	}
	found := false
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "nvidia-device-plugin-daemonset") {
			podName = pod.Name
			podNamespace = pod.Namespace
			found = true
			break
		}
	}
	if !found {
		//TODO: indicate here this just means no GPUs on this node
		fmt.Println("Could not find nvidia device plugin daemondset pod",
			"\nThis may be the result of a node with no GPUs")
		return "", 1, nil
	}

	req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(podNamespace).SubResource("exec")
	command := []string{"/bin/bash", "-c", "nvidia-smi -q -x && echo -n $?"}
	option := &corev1.PodExecOptions{
		Command: command,
		//TODO: Past issue: MIGHT need to specify container
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		TTY:    false,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	// execute nvidia-smi command
	var output string
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Println("Error creating executor for nvidia-smi")
		return output, 1, err
	}

	// get output as a string
	buf := new(bytes.Buffer)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: buf,
		Stderr: nil,
		Tty:    false,
	})
	if err != nil {
		fmt.Println("Error getting output from command")
		return output, 1, err
	}
	output = buf.String()

	// parse output for NVSMI output and exit code
	idx := strings.LastIndex(output, "\n")
	nvSMIout := output[0:idx]
	exitCode, err := strconv.Atoi(output[idx+1 : len(output)])
	if err != nil {
		fmt.Println("Non-integer exit code\n")
		return "", 1, err
	}

	return nvSMIout, exitCode, nil
}

func updateNodes(clientset *kubernetes.Clientset, config *rest.Config) ([]NodeInfo, error) {
	//get list of all nodes
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error in getting list of cluster.Nodes")
		return nil, err
	}

	var nodes []NodeInfo

	//for each node:
	//create a NodeInfo instance
	//populate instance with nodespec/status information
	//run NVSMI to populate rest of node
	//append to nodes

	//for each node:
	for _, node := range nodeList.Items {
		var info NodeInfo
		//populate instance with nodespec/status information
		info.OSVersion = node.Status.NodeInfo.OSImage
		info.OS = node.Status.NodeInfo.OperatingSystem
		info.KubeletVersion = node.Status.NodeInfo.KubeletVersion
		info.ContainerRuntime = node.Status.NodeInfo.ContainerRuntimeVersion
		info.NodeName = node.Name

		// find VM or bare metal by getting labels
		// TODO: Some nodes will omit this label - so this isn't 100% reliable
		labels := node.Labels
		hyperLabel := "feature.node.kubernetes.io/cpu-cpuid.HYPERVISOR"
		if value, present := labels[hyperLabel]; value == "true" {
			info.VMOrBareMetal = "VM"
		} else if present {
			info.VMOrBareMetal = "Bare Metal"
		} else {
			// label not present
			info.VMOrBareMetal = "Unknown"
		}

		//run NVSMI to populate rest of node
		NVSMIOutput, exitCode, err := runNVSMI(clientset, config, node.Name)
		if err != nil {
			fmt.Println("Error running NVSMI")
			return nil, err
		}
		xmlOutput := []byte(NVSMIOutput)
		xml.Unmarshal(xmlOutput, &info)
		info.NVSMIExitCode = exitCode

		//Chop up GPU Names (for vgpu profiles)
		for i := range info.GPUs {
			name := info.GPUs[i].ProductName
			if dashIndex := strings.Index(name, "-"); dashIndex != -1 {
				info.GPUs[i].ProductName = name[0:dashIndex]
				fields := strings.Fields(name)
				info.GPUs[i].GpuVirtualizationMode.VGPUProfile = fields[len(fields)-1]
			}
		}

		// Get vgpu host driver
		vGPUHostDriverLabel := "nvidia.com/vgpu.host-driver-version"
		if value, present := labels[vGPUHostDriverLabel]; present {
			info.VGPUHostDriver = value
		} else {
			info.VGPUHostDriver = "N/A"
		}

		//handle Mellanox NICs
		info.NICs, err = readNICs(clientset, config, node.Name)
		if err != nil {
			fmt.Println("Error reading NICs")
			return nodes, err
		}
		info.AttachedNICs = len(info.NICs)

		//append to nodes
		nodes = append(nodes, info)
	}

	//return nodes
	return nodes, nil
}

func getDistribution(clientset *kubernetes.Clientset, config *rest.Config) (string, error) {
	//get list of all nodes
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error in getting list of cluster.Nodes")
		return "", err
	}

	//only need to check one node
	labels := nodeList.Items[0].Labels

	// distributions to try to detect:
	distributions := []string{"Tanzu", "OpenShift", "K3s", "RKE"}
	distributionsLower := []string{"tanzu", "openshift", "k3s", ".rke."}

	// grep style search for distribution
	var match string
	matches := 0
	for idx, distribution := range distributionsLower {
		for label := range labels {
			if strings.Contains(label, distribution) {
				//found a match
				match = distributions[idx]
				matches = matches + 1
			}
		}
	}

	// If no match, it's standard distribution
	if matches == 0 {
		match = "Standard"
	} else if matches > 1 {
		// multiple distributions detected, cant confirm one
        //TODO: report the two that it found
		match = "Unknown"
	}
	return match, nil
}

func getCloudInfo(clientset *kubernetes.Clientset, config *rest.Config) (string, string, error) {
	//get list of all nodes
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error in getting list of cluster.Nodes")
		return "", "", err
	}

	//only need to check one node
	anns := nodeList.Items[0].Annotations
	labels := nodeList.Items[0].Labels

	// grep style search for "cloud"
	for ann := range anns {
		if strings.Contains(ann, "cloud") || strings.Contains(ann, "Cloud") {
			return "Cloud", nodeList.Items[0].Spec.ProviderID, nil
		}
	}
	for label := range labels {
		if strings.Contains(label, "cloud") || strings.Contains(label, "Cloud") {
			return "Cloud", nodeList.Items[0].Spec.ProviderID, nil
		}
	}

	// if we reach here without returning, no cloud detected
	return "On Prem", "N/A", nil
}

func readNICs(clientset *kubernetes.Clientset, config *rest.Config, nodeName string) ([]string, error) {
	// find the mofed pod associated with this node
	var podName string
	var podNamespace string
	var NICs []string
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(),
		metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		fmt.Println("Error getting pods")
		return NICs, err
	}
	found := false
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "mofed") {
			podName = pod.Name
			podNamespace = pod.Namespace
			found = true
			break
		}
	}
	if !found {
		//no mellanox NICs on cluster
		return NICs, nil
	}

	// if found, we have Mellanox NICS, so run lspci

	req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(podNamespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: []string{"lspci"},
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	// execute lspci command
	var output string
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Println("Error creating executor for lspci")
		return NICs, err
	}

	// get output as a string
	buf := new(bytes.Buffer)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: buf,
		Stderr: nil,
		Tty:    false,
	})
	if err != nil {
		fmt.Println("Error getting output from lspci command")
		return NICs, err
	}
	output = buf.String()

	// search for the NICs in lspci output
	scanner := bufio.NewScanner(strings.NewReader(output))
	inserted := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "Mellanox") {
			continue
		}
		// get first 5 chars: "xx:xx" this is "bus num:device num"
		cardID := line[0:5]

		// avoid inserting a card twice:
		if _, present := inserted[cardID]; !present {
			inserted[cardID] = true

			//get card name
			name := line[strings.Index(line, "Mellanox"):len(line)]
			NICs = append(NICs, name)
		}
	}

	return NICs, nil
}
