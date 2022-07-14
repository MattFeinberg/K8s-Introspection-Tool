package datagathering

import(
    "time"
    "fmt"
    "encoding/csv"
    "encoding/json"
    "strings"
    "bytes"
    "os"
    "context"
    "encoding/xml"
     "k8s.io/client-go/kubernetes"
     corev1 "k8s.io/api/core/v1"
     metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
     "k8s.io/client-go/kubernetes/scheme"
     "k8s.io/client-go/tools/clientcmd"
     "k8s.io/client-go/tools/remotecommand"
     "k8s.io/client-go/rest"
)

//var cluster ClusterInfo
//
//func GetCluster() (ClusterInfo) {
//    return cluster
//}

// Break this down into more functions
// TODO: I dont like this functions name
func HandleDataUpdates() (ClusterInfo) {
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

    //Count GPUs
    totalGPUs := 0
    for _, node := range cluster.Nodes {
        totalGPUs = totalGPUs + node.AttachedGpus
    }

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
            if GPU.MigMode.CurrentMig == "Enabled" {
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
    cluster.Distribution = "Dont know yet"
    cluster.TotalGPUs = totalGPUs
    cluster.TimeUpdated = time.Now().Format("01-02-2006 15:04:05")

    // write to csv

    csvFile, err := os.OpenFile("data.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Print("failed creating csv file\n", err)
        return cluster
    }
    defer csvFile.Close()

    csvWriter := csv.NewWriter(csvFile)
    str, err := json.Marshal(cluster)
    row := []string{"timestamp here", string(str)}
    if err := csvWriter.Write(row); err != nil {
        fmt.Print("error writing to file\n", err)
        return cluster
    }

    return cluster
}



func GetK8s() (*kubernetes.Clientset, *rest.Config, error) {

    // set up config and clientset for the cluster
    rules := clientcmd.NewDefaultClientConfigLoadingRules()
    kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
    config, err := kubeconfig.ClientConfig()
    // the following line only works when running directly on cluster
    //config, err := rest.InClusterConfig()
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
func runNVSMI(clientset *kubernetes.Clientset, config *rest.Config, nodeName string) (string, error){
    // find the driver daemonset associated with this node
    var podName string
    var podNamespace string
    //TODO: namespaces might be unique to a cluster?
    //namespace := "nvidia-gpu-operator"
    pods, err := clientset.CoreV1().Pods("").List(context.TODO(),
                 metav1.ListOptions{ FieldSelector: "spec.nodeName=" + nodeName,})
    if err != nil {
        fmt.Println("Error getting pods")
        return "", err
    }
    found := false
    for _, pod := range pods.Items {
        if strings.HasPrefix(pod.Name, "nvidia-driver-daemonset") {
            podName = pod.Name
            podNamespace = pod.Namespace
            found = true
            break
        }
    }
    if !found {
        fmt.Println("Could not find nvidia driver daemondset pod")
        return "", nil
    }

    req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
           Namespace(podNamespace).SubResource("exec")
    option := &corev1.PodExecOptions{
        Command: []string{"nvidia-smi", "-q", "-x"},
        //TODO: confirm that it's always this container?
        Container: "nvidia-driver-ctr",
        Stdin:   false,
        Stdout:  true,
        Stderr:  true,
        TTY:     false,
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
        return output, err
    }

    // get output as a string
    buf := new(bytes.Buffer)
    err = exec.Stream(remotecommand.StreamOptions{
        Stdin:  os.Stdin,
        Stdout: buf,
        Stderr: nil,
        Tty: false,
    })
    if err != nil {
        fmt.Println("Error getting output from command")
        return output, err
    }
    output = buf.String()
    return output, nil
}

func updateNodes(clientset *kubernetes.Clientset, config *rest.Config) ([]NodeInfo, error){
    //get list of all nodes
    nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Println("Error in getting list of cluster.Nodes")
        return nil, err
    }

    // this will get returned, and nodesPtr will point to it
    var nodes []NodeInfo

    //for each node:
        //create a NodeInfo instance
        //populate instance with nodespec/status information
        //run NVSMI to populate rest of node
        //append to nodes

    for _, node := range nodeList.Items {
        var info NodeInfo
        //populate instance with nodespec/status information
        info.OSVersion = node.Status.NodeInfo.OSImage
        info.K8sVersion = node.Status.NodeInfo.KubeletVersion
        info.ContainerRuntime = node.Status.NodeInfo.ContainerRuntimeVersion
        info.NodeName = node.Name
        // still have yet to find CSP
        //OutputStruct.Provider = (nodeList.Items)[nodeIndex].Spec.ProviderID

        //run NVSMI to populate rest of node
        NVSMIOutput, err := runNVSMI(clientset, config, node.Name)
        if err != nil {
            fmt.Println("Error running NVSMI")
            return nil, err
        }
        xmlOutput := []byte(NVSMIOutput)
        xml.Unmarshal(xmlOutput, &info)

        //append to nodes
        nodes = append(nodes, info)
    }

    //return nodes
    return nodes, nil
}
