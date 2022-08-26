package webhosting

import (
	data "internal/datagathering"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// BuildHandleFunc builds an http handler function that has access to the cluster struct variable
func BuildHandleFunc(cluster *data.ClusterInfo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

func executeIndex(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset, config *rest.Config) {
	fileName := "html/index.html"
	t, err := template.ParseFiles(fileName)
	if err != nil {
		fmt.Println("Error parsing index template", err)
		return
	}

	err = t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		fmt.Println("Error executing index template", err)
		return
	}
}

func executeClusterInfo(w http.ResponseWriter, r *http.Request, cluster *data.ClusterInfo) {
	fileName := "html/clusterInfo.html"
	t, err := template.ParseFiles(fileName)
	if err != nil {
		fmt.Println("Error parsing clusterInfo template", err)
		return
	}

	err = t.ExecuteTemplate(w, "clusterInfo.html", *cluster)
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
	fileName := "html/nodeinfo.html"
	t, err := template.ParseFiles(fileName)
	if err != nil {
		fmt.Println("Error parsing nodeinfo template", err)
		return
	}

	err = t.ExecuteTemplate(w, "nodeinfo.html", struct {
		TimeUpdated string
		Node        data.NodeInfo
	}{
		TimeUpdated: cluster.TimeUpdated,
		Node:        cluster.Nodes[nodeIndex]})
	if err != nil {
		fmt.Println("Error executing nodeinfo template", err)
		return
	}
}
