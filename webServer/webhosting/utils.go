package webhosting

import(
    data "introspec-proj/webServer/datagathering"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "fmt"
    "net/http"
    "html/template"
    "strings"
    "strconv"
    "context"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)


//func testHandler(name string) http.HandlerFunc {
//    return func(w http.ResponseWriter, r *http.Request) {
//        w.Write([]byte("Hi, from Service: " + name))
//    }
//}

func BuildHandleFunc(cluster *data.ClusterInfo) (func(w http.ResponseWriter, r *http.Request)) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the K8s info and nodeList
        // note: how often should this be updated?
        //       might have to move it inside each case
        //       or case function
        clientset, config, err := data.GetK8s()
        if err != nil {
            fmt.Println("Error getting K8s info", err)
            return
        }


        switch r.URL.Path {
            case "/":
                executeIndex(w, r, clientset, config)
            case "/home":
                executeIndex(w, r, clientset, config)
            case "/clusterinfo":
                executeClusterInfo(w, r, cluster)
            case "/nodeinfo":
                executeNodeInfo(w, r, cluster)
        }
    }
}
//func HandleFunc(cluster data.ClusterInfo) {
//    return func(w http.ResponseWriter, r *http.Request) {
//        // Get the K8s info and nodeList
//        // note: how often should this be updated?
//        //       might have to move it inside each case
//        //       or case function
//        clientset, config, err := data.GetK8s()
//        if err != nil {
//            fmt.Println("Error getting K8s info", err)
//            return
//        }
//
//
//        switch r.URL.Path {
//            case "/":
//                executeIndex(w, r, clientset, config)
//            case "/home":
//                executeIndex(w, r, clientset, config)
//            case "/clusterinfo":
//                executeClusterInfo(w, r, cluster)
//            case "/nodeinfo":
//                executeNodeInfo(w, r, cluster)
//        }
//    }
//}

// maybe make a function that takes a filename and does this...
func executeIndex(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset, config *rest.Config) {
    fileName := "index.html"
    t, err := template.ParseFiles(fileName)
    if err != nil {
        fmt.Println("Error parsing index template", err)
        return
    }

    // get prometheus and grafana IP
    // maybe make this a function
    namespace := "introspection-resources"
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Println("Error getting pods", err)
        return
    }

    IPs :=  TelemetryIPs{GrafanaEnabled: false, PrometheusEnabled: false,}
    for _, pod := range pods.Items {
        if !IPs.GrafanaEnabled && strings.HasPrefix(pod.Name, "grafana") {
            IPs.GrafanaIP = "http://" + pod.Status.HostIP + ":32000"
            IPs.GrafanaEnabled = true
        }
        if !IPs.PrometheusEnabled && strings.HasPrefix(pod.Name, "prometheus") {
            IPs.PrometheusIP = "http://" + pod.Status.HostIP + ":30070"
            IPs.PrometheusEnabled = true
        }
        if IPs.GrafanaEnabled && IPs.PrometheusEnabled {
            break
        }
    }
    //note: add error if grafana is enabled and prometheus isnt
    //      this should not be possible due to helm chart


    err = t.ExecuteTemplate(w, fileName, IPs)
    if err != nil {
        fmt.Println("Error executing index template", err)
        return
    }
}

func executeClusterInfo(w http.ResponseWriter, r *http.Request, cluster *data.ClusterInfo) {
    fileName := "clusterInfo.html"
    t, err := template.ParseFiles(fileName)
    if err != nil {
        fmt.Println("Error parsing clusterInfo template", err)
        return
    }

    err = t.ExecuteTemplate(w, fileName, *cluster)
    if err != nil {
        fmt.Println("Error executing clusterInfo template", err)
        return
    }
}

func executeNodeInfo(w http.ResponseWriter, r *http.Request, cluster *data.ClusterInfo) {

    nodeIndexString := r.FormValue("idx")
    nodeIndex, err := strconv.Atoi(nodeIndexString)
    if err != nil {
        fmt.Println("Node index not of type int")
    }
    fileName := "nodeinfo.html"
    t, err := template.ParseFiles(fileName)
    if err != nil {
        fmt.Println("Error parsing nodeinfo template", err)
        return
    }

    err = t.ExecuteTemplate(w, fileName, struct{TimeUpdated string; Node data.NodeInfo}{
                                                TimeUpdated: cluster.TimeUpdated,
                                                Node: cluster.Nodes[nodeIndex]})
    if err != nil {
        fmt.Println("Error executing nodeinfo template", err)
        return
    }
}
