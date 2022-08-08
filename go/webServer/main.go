/*
 * Go program to gather information on a Kubernetes cluster
 * Gathers information on each node, then uses nvidia-smi and
 * dcgm-exporter to get GPU metrics for the cluster
 */


//TODO: Error handling - where do i quit, where do i keep
//      running but indicate something is missing
package main

import (
        data "introspec-proj/webServer/datagathering"
        web "introspec-proj/webServer/webhosting"
        "os"
        "net/http"
        "time"
        "strconv"
        "fmt"
)

func main() {
    // initial data gathering
    // TODO: Make a constructor for cluster maybe?
    cluster := data.HandleDataUpdates()

    // periodic data gathering
    var rate int
    rate, err := strconv.Atoi(os.Getenv("RATE"))
    if err != nil {
        fmt.Println("Error reading rate variable as integer")
    }
    // default rate is 24 (hours)
    ticker := time.NewTicker(time.Duration(rate) * time.Hour)
    quit := make(chan struct{})
    go func() {
        for {
           select {
            case <- ticker.C:
                // do stuff
                cluster = data.HandleDataUpdates()
            case <- quit:
                ticker.Stop()
                return
            }
        }
     }()

    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
    //handleFunc := web.BuildHandleFunc(&cluster)
    //http.HandleFunc("/", handleFunc)
    //http.ListenAndServe(":8000", nil)
}


// this was the blueprint of most of the data gathering
// keeping around until I've used all i need from it

//func getInfo() {
//    // set up config and clientset for the cluster
//    rules := clientcmd.NewDefaultClientConfigLoadingRules()
//    kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
//    config, err := kubeconfig.ClientConfig()
//    // the following line only works when running directly on cluster
//    //config, err := rest.InClusterConfig()
//    if err != nil {
//        fmt.Println(err, "\nError in getting config")
//        return
//    }
//
//    // create the clientset
//    clientset, err := kubernetes.NewForConfig(config)
//    if err != nil {
//        fmt.Println(err, "\nError in getting access to K8s")
//        return
//    }
//
//    //Gather info on each node:
//    listOpts := metav1.ListOptions{}
//    nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), listOpts)
//    if err != nil {
//        fmt.Println(err, "\nError in getting list of cluster.Nodes")
//        return
//    }
//    // For each node, just print NodeInfo
//    for idx, node := range nodeList.Items {
//        fmt.Print("Node ", idx, " Detected\n")
//        fmt.Printf("%+v\n\n", node.Status.NodeInfo)
//    }
//
//    // get output from nvidia-smi
//    // note: podName will likely change for each cluster
//    //       how can i get pod name without unique tag
//    // note: How does this change when it's a multi-node cluster
//    //       we might have multiple of these pods - 1 per node?
//    podName := "nvidia-driver-daemonset-5ghmt"
//    namespace := "gpu-operator-resources"
//    req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
//           Namespace(namespace).SubResource("exec")
//    option := &corev1.PodExecOptions{
//        Command: []string{"nvidia-smi", "-q"},
//        Stdin:   false,
//        Stdout:  true,
//        Stderr:  true,
//        TTY:     false,
//    }
//    req.VersionedParams(
//        option,
//        scheme.ParameterCodec,
//    )
//    // execute nvidia-smi command using request
//    exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
//    if err != nil {
//        fmt.Println(err, "\nError creating executor for nvidia-smi")
//        return
//    }
//
//    // get output as a string
//    buf := new(bytes.Buffer)
//    err = exec.Stream(remotecommand.StreamOptions{
//        Stdin:  os.Stdin,
//        Stdout: buf,
//        Stderr: nil,
//        Tty: false,
//    })
//    if err != nil {
//        fmt.Println(err, "\nError getting output from command")
//        return
//    }
//    str := buf.String()
//    fmt.Println(str)
//
//    //
//    // get output from dcgm-exporter
//    //
//
//    // find the dcgm-exporter ClusterIP service
//    serviceList, err := clientset.CoreV1().Services("gpu-operator-resources").List(context.TODO(), listOpts)
//    if err != nil {
//        fmt.Println(err, "\nError in getting list of cluster.Nodes")
//        return
//    }
//
//    found := false
//    var clusterIP string
//    for _, svc := range serviceList.Items {
//        if svc.Namespace == "gpu-operator-resources" && svc.Name == "nvidia-dcgm-exporter" {
//            found = true
//            clusterIP = svc.Spec.ClusterIP
//            break
//        }
//    }
//    if !found {
//        fmt.Println(err, "\nFailed to get dcgm-exporter service")
//        return
//    }
//
//    // Use an HTTP request to get metrics from DCGM
//    httpTarget := "http://" + clusterIP + ":9400" + "/metrics"
//    resp, err := http.Get(httpTarget)
//    if err != nil {
//        fmt.Println(err, "\nError getting dcgm metrics")
//        return
//    }
//    body, err := io.ReadAll(resp.Body)
//    if err != nil {
//        fmt.Println(err, "\nError reading dcgm-exporter metrics from HTTP")
//        return
//    }
//    str = string(body)
//    fmt.Println(str)
//    defer resp.Body.Close()
//}
//
