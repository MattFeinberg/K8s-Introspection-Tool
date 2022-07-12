//TODO:
// how often to update info? when only updating at cluster info, error:
// check logs - seems like the nodesptr slice is not being updated

/*
 * Go program to gather information on a Kubernetes cluster
 * Gathers information on each node, then uses nvidia-smi and
 * dcgm-exporter to get GPU metrics for the cluster
 */
package main

import (
        "fmt"
        "os"
        "context"
        "bytes"
        "net/http"
        "html/template"
        "io"
        "strconv"
        "strings"
        "encoding/xml"
        //"time"
        //"bytes"
        //appsv1 "k8s.io/api/apps/v1"
        corev1 "k8s.io/api/core/v1"
        //"k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        //"k8s.io/apimachinery/pkg/types"
        //"k8s.io/apimachinery/pkg/util/intstr"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/kubernetes/scheme"
        "k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/tools/remotecommand"
        "k8s.io/client-go/rest"
        //"k8s.io/client-go/discovery"
        //"k8s.io/apimachinery/pkg/runtime"
        //ctrl "sigs.k8s.io/controller-runtime"
        //client "sigs.k8s.io/controller-runtime/pkg/client"
        //"sigs.k8s.io/controller-runtime/pkg/log"
        //"github.com/prometheus/client_golang/prometheus/promhttp"
       )

func main() {
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
    http.HandleFunc("/", handleFunc)
    http.ListenAndServe(":8000", nil)
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
    // Get the K8s info and nodeList
    // note: how often should this be updated?
    //       might have to move it inside each case
    //       or case function
    clientset, config, err := getK8s()
    if err != nil {
        fmt.Println("Error getting K8s info", err)
        return
    }

    var nodes []NodeInfo

    switch r.URL.Path {
        case "/":
            executeIndex(w, r, clientset, config)
        case "/home":
            executeIndex(w, r, clientset, config)
        case "/clusterinfo":
            executeClusterInfo(w, r, clientset, config, &nodes)
        case "/nodeinfo":
            executeNodeInfo(w, r, clientset, config, nodes)
    }
}

type TelemetryIPs struct {
    GrafanaEnabled bool
    PrometheusEnabled bool
    GrafanaIP string
    PrometheusIP string
}
// maybe make a function that takes a filename and does this...
func executeIndex(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset, config *rest.Config) {
    fileName := "index.html"
    t, err := template.ParseFiles(fileName)
    if err != nil {
        fmt.Println("Error parsing index template", err)
        return
    }

    // get prometheus and grafana IP
    namespace := "introspection-resources"
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Println("Error getting pods", err)
        return
    }

    // maybe make this a function
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

type ClusterInfo struct {
    NumNodes int
    NodeNames []string
    Distribution string
    TotalGPUs int
}
func executeClusterInfo(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset,
                        config *rest.Config, nodesPtr *[]NodeInfo) {
    fileName := "clusterInfo.html"
    t, err := template.ParseFiles(fileName)
    if err != nil {
        fmt.Println("Error parsing clusterInfo template", err)
        return
    }

    // gather cluster data
    // desired variables:
    *nodesPtr, err = updateNodes(clientset, config)
    if err != nil {
        fmt.Println("Error updating nodes")
        return
    }

    //TODO: count GPUs
    var totalGPUs int

    // get node names
    var names []string
    for _, elem := range *nodesPtr {
        names = append(names, elem.NodeName)
    }
    // tentative: get total GPUs (need multiple nodes?)
    // run nvsmi
    totalGPUs = 1

    Cluster := ClusterInfo{
                   NumNodes: len(*nodesPtr),
                   NodeNames: names,
                   Distribution: "Dont know yet",
                   TotalGPUs: totalGPUs,
               }
    err = t.ExecuteTemplate(w, fileName, Cluster)
    if err != nil {
        fmt.Println("Error executing clusterInfo template", err)
        return
    }
}

func executeNodeInfo(w http.ResponseWriter, r *http.Request,
                     clientset *kubernetes.Clientset, config *rest.Config, nodesPtr []NodeInfo) {
//    //gather node info
//    //gatherNodeInfo()
//    nodeName := r.FormValue("name")

    nodesPtr, err := updateNodes(clientset, config)

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

//    // Get node info
//    // do this from handlefunc
//    listOpts := metav1.ListOptions{}
//    nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), listOpts)
//    if err != nil {
//        fmt.Println(err, "\nError in getting list of cluster Nodes")
//        return
//    }
//
//    // run nvidia-smi on this node
//    var OutputStruct NodeInfo
//    output, err := runNVSMI(clientset, config, nodeName)
//    xmlOutput := []byte(output)
//    xml.Unmarshal(xmlOutput, &OutputStruct)
//
//    // this will change once I get multiple nodes
//    node := (nodeList.Items)[nodeIndex].Status.NodeInfo
//    OutputStruct.OSVersion = node.OSImage
//    OutputStruct.K8sVersion = node.KubeletVersion
//    OutputStruct.ContainerRuntime = node.ContainerRuntimeVersion
//    OutputStruct.NodeName = nodeName
//    // still have yet to find CSP
//    //OutputStruct.Provider = (nodeList.Items)[nodeIndex].Spec.ProviderID

    err = t.ExecuteTemplate(w, fileName, nodesPtr[nodeIndex])
    if err != nil {
        fmt.Println("Error executing nodeinfo template", err)
        return
    }
}


func getK8s() (*kubernetes.Clientset, *rest.Config, error) {

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
    namespace := "gpu-operator-resources"
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(),
                 metav1.ListOptions{ FieldSelector: "spec.nodeName=" + nodeName,})
    if err != nil {
        fmt.Println("Error getting pods")
        return "", err
    }
    found := false
    for _, pod := range pods.Items {
        if strings.HasPrefix(pod.Name, "nvidia-driver-daemonset") {
            podName = pod.Name
            found = true
            break
        }
    }
    if !found {
        fmt.Println("Could not find nvidia driver daemondset pod")
        return "", nil
    }

    req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
           Namespace(namespace).SubResource("exec")
    option := &corev1.PodExecOptions{
        Command: []string{"nvidia-smi", "-q", "-x"},
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
        fmt.Println("Error in getting list of cluster Nodes")
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

func getInfo() {
    // set up config and clientset for the cluster
    rules := clientcmd.NewDefaultClientConfigLoadingRules()
    kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
    config, err := kubeconfig.ClientConfig()
    // the following line only works when running directly on cluster
    //config, err := rest.InClusterConfig()
    if err != nil {
        fmt.Println(err, "\nError in getting config")
        return
    }

    // create the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Println(err, "\nError in getting access to K8s")
        return
    }

    //Gather info on each node:
    listOpts := metav1.ListOptions{}
    nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), listOpts)
    if err != nil {
        fmt.Println(err, "\nError in getting list of cluster Nodes")
        return
    }
    // For each node, just print NodeInfo
    for idx, node := range nodeList.Items {
        fmt.Print("Node ", idx, " Detected\n")
        fmt.Printf("%+v\n\n", node.Status.NodeInfo)
    }

    // get output from nvidia-smi
    // note: podName will likely change for each cluster
    //       how can i get pod name without unique tag
    // note: How does this change when it's a multi-node cluster
    //       we might have multiple of these pods - 1 per node?
    podName := "nvidia-driver-daemonset-5ghmt"
    namespace := "gpu-operator-resources"
    req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
           Namespace(namespace).SubResource("exec")
    option := &corev1.PodExecOptions{
        Command: []string{"nvidia-smi", "-q"},
        Stdin:   false,
        Stdout:  true,
        Stderr:  true,
        TTY:     false,
    }
    req.VersionedParams(
        option,
        scheme.ParameterCodec,
    )
    // execute nvidia-smi command using request
    exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
    if err != nil {
        fmt.Println(err, "\nError creating executor for nvidia-smi")
        return
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
        fmt.Println(err, "\nError getting output from command")
        return
    }
    str := buf.String()
    fmt.Println(str)

    //
    // get output from dcgm-exporter
    //

    // find the dcgm-exporter ClusterIP service
    serviceList, err := clientset.CoreV1().Services("gpu-operator-resources").List(context.TODO(), listOpts)
    if err != nil {
        fmt.Println(err, "\nError in getting list of cluster Nodes")
        return
    }

    found := false
    var clusterIP string
    for _, svc := range serviceList.Items {
        if svc.Namespace == "gpu-operator-resources" && svc.Name == "nvidia-dcgm-exporter" {
            found = true
            clusterIP = svc.Spec.ClusterIP
            break
        }
    }
    if !found {
        fmt.Println(err, "\nFailed to get dcgm-exporter service")
        return
    }

    // Use an HTTP request to get metrics from DCGM
    httpTarget := "http://" + clusterIP + ":9400" + "/metrics"
    resp, err := http.Get(httpTarget)
    if err != nil {
        fmt.Println(err, "\nError getting dcgm metrics")
        return
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err, "\nError reading dcgm-exporter metrics from HTTP")
        return
    }
    str = string(body)
    fmt.Println(str)
    defer resp.Body.Close()
}


//GPU STRUCT
type NodeInfo struct {
    NodeName         string
    OSVersion        string
    K8sVersion       string
    ContainerRuntime string
    NumGPUs          string
    Provider         string
    XMLName          xml.Name `xml:"nvidia_smi_log"`
    Text             string   `xml:",chardata"`
    Timestamp        string   `xml:"timestamp"`
    DriverVersion    string   `xml:"driver_version"`
    CudaVersion      string   `xml:"cuda_version"`
    AttachedGpus     int   `xml:"attached_gpus"`
    GPUs            []struct{
        Text                string `xml:",chardata"`
        ID                  string `xml:"id,attr"`
        ProductName         string `xml:"product_name"`
        ProductBrand        string `xml:"product_brand"`
        ProductArchitecture string `xml:"product_architecture"`
        DisplayMode         string `xml:"display_mode"`
        DisplayActive       string `xml:"display_active"`
        PersistenceMode     string `xml:"persistence_mode"`
        MigMode             struct {
            Text       string `xml:",chardata"`
            CurrentMig string `xml:"current_mig"`
            PendingMig string `xml:"pending_mig"`
        } `xml:"mig_mode"`
        MigDevices               string `xml:"mig_devices"`
        AccountingMode           string `xml:"accounting_mode"`
        AccountingModeBufferSize string `xml:"accounting_mode_buffer_size"`
        DriverModel              struct {
            Text      string `xml:",chardata"`
            CurrentDm string `xml:"current_dm"`
            PendingDm string `xml:"pending_dm"`
        } `xml:"driver_model"`
        Serial         string `xml:"serial"`
        Uuid           string `xml:"uuid"`
        MinorNumber    string `xml:"minor_number"`
        VbiosVersion   string `xml:"vbios_version"`
        MultigpuBoard  string `xml:"multigpu_board"`
        BoardID        string `xml:"board_id"`
        GpuPartNumber  string `xml:"gpu_part_number"`
        GpuModuleID    string `xml:"gpu_module_id"`
        InforomVersion struct {
            Text       string `xml:",chardata"`
            ImgVersion string `xml:"img_version"`
            OemObject  string `xml:"oem_object"`
            EccObject  string `xml:"ecc_object"`
            PwrObject  string `xml:"pwr_object"`
        } `xml:"inforom_version"`
        GpuOperationMode struct {
            Text       string `xml:",chardata"`
            CurrentGom string `xml:"current_gom"`
            PendingGom string `xml:"pending_gom"`
        } `xml:"gpu_operation_mode"`
        GspFirmwareVersion    string `xml:"gsp_firmware_version"`
        GpuVirtualizationMode struct {
            Text               string `xml:",chardata"`
            VirtualizationMode string `xml:"virtualization_mode"`
            HostVgpuMode       string `xml:"host_vgpu_mode"`
        } `xml:"gpu_virtualization_mode"`
        Ibmnpu struct {
            Text                string `xml:",chardata"`
            RelaxedOrderingMode string `xml:"relaxed_ordering_mode"`
        } `xml:"ibmnpu"`
        Pci struct {
            Text           string `xml:",chardata"`
            PciBus         string `xml:"pci_bus"`
            PciDevice      string `xml:"pci_device"`
            PciDomain      string `xml:"pci_domain"`
            PciDeviceID    string `xml:"pci_device_id"`
            PciBusID       string `xml:"pci_bus_id"`
            PciSubSystemID string `xml:"pci_sub_system_id"`
            PciGpuLinkInfo struct {
                Text    string `xml:",chardata"`
                PcieGen struct {
                    Text           string `xml:",chardata"`
                    MaxLinkGen     string `xml:"max_link_gen"`
                    CurrentLinkGen string `xml:"current_link_gen"`
                } `xml:"pcie_gen"`
                LinkWidths struct {
                    Text             string `xml:",chardata"`
                    MaxLinkWidth     string `xml:"max_link_width"`
                    CurrentLinkWidth string `xml:"current_link_width"`
                } `xml:"link_widths"`
            } `xml:"pci_gpu_link_info"`
            PciBridgeChip struct {
                Text           string `xml:",chardata"`
                BridgeChipType string `xml:"bridge_chip_type"`
                BridgeChipFw   string `xml:"bridge_chip_fw"`
            } `xml:"pci_bridge_chip"`
            ReplayCounter         string `xml:"replay_counter"`
            ReplayRolloverCounter string `xml:"replay_rollover_counter"`
            TxUtil                string `xml:"tx_util"`
            RxUtil                string `xml:"rx_util"`
        } `xml:"pci"`
        FanSpeed              string `xml:"fan_speed"`
        PerformanceState      string `xml:"performance_state"`
        ClocksThrottleReasons struct {
            Text                                          string `xml:",chardata"`
            ClocksThrottleReasonGpuIdle                   string `xml:"clocks_throttle_reason_gpu_idle"`
            ClocksThrottleReasonApplicationsClocksSetting string `xml:"clocks_throttle_reason_applications_clocks_setting"`
            ClocksThrottleReasonSwPowerCap                string `xml:"clocks_throttle_reason_sw_power_cap"`
            ClocksThrottleReasonHwSlowdown                string `xml:"clocks_throttle_reason_hw_slowdown"`
            ClocksThrottleReasonHwThermalSlowdown         string `xml:"clocks_throttle_reason_hw_thermal_slowdown"`
            ClocksThrottleReasonHwPowerBrakeSlowdown      string `xml:"clocks_throttle_reason_hw_power_brake_slowdown"`
            ClocksThrottleReasonSyncBoost                 string `xml:"clocks_throttle_reason_sync_boost"`
            ClocksThrottleReasonSwThermalSlowdown         string `xml:"clocks_throttle_reason_sw_thermal_slowdown"`
            ClocksThrottleReasonDisplayClocksSetting      string `xml:"clocks_throttle_reason_display_clocks_setting"`
        } `xml:"clocks_throttle_reasons"`
        FbMemoryUsage struct {
            Text     string `xml:",chardata"`
            Total    string `xml:"total"`
            Reserved string `xml:"reserved"`
            Used     string `xml:"used"`
            Free     string `xml:"free"`
        } `xml:"fb_memory_usage"`
        Bar1MemoryUsage struct {
            Text  string `xml:",chardata"`
            Total string `xml:"total"`
            Used  string `xml:"used"`
            Free  string `xml:"free"`
        } `xml:"bar1_memory_usage"`
        ComputeMode string `xml:"compute_mode"`
        Utilization struct {
            Text        string `xml:",chardata"`
            GpuUtil     string `xml:"gpu_util"`
            MemoryUtil  string `xml:"memory_util"`
            EncoderUtil string `xml:"encoder_util"`
            DecoderUtil string `xml:"decoder_util"`
        } `xml:"utilization"`
        EncoderStats struct {
            Text           string `xml:",chardata"`
            SessionCount   string `xml:"session_count"`
            AverageFps     string `xml:"average_fps"`
            AverageLatency string `xml:"average_latency"`
        } `xml:"encoder_stats"`
        FbcStats struct {
            Text           string `xml:",chardata"`
            SessionCount   string `xml:"session_count"`
            AverageFps     string `xml:"average_fps"`
            AverageLatency string `xml:"average_latency"`
        } `xml:"fbc_stats"`
        EccMode struct {
            Text       string `xml:",chardata"`
            CurrentEcc string `xml:"current_ecc"`
            PendingEcc string `xml:"pending_ecc"`
        } `xml:"ecc_mode"`
        EccErrors struct {
            Text     string `xml:",chardata"`
            Volatile struct {
                Text              string `xml:",chardata"`
                SramCorrectable   string `xml:"sram_correctable"`
                SramUncorrectable string `xml:"sram_uncorrectable"`
                DramCorrectable   string `xml:"dram_correctable"`
                DramUncorrectable string `xml:"dram_uncorrectable"`
            } `xml:"volatile"`
            Aggregate struct {
                Text              string `xml:",chardata"`
                SramCorrectable   string `xml:"sram_correctable"`
                SramUncorrectable string `xml:"sram_uncorrectable"`
                DramCorrectable   string `xml:"dram_correctable"`
                DramUncorrectable string `xml:"dram_uncorrectable"`
            } `xml:"aggregate"`
        } `xml:"ecc_errors"`
        RetiredPages struct {
            Text                        string `xml:",chardata"`
            MultipleSingleBitRetirement struct {
                Text            string `xml:",chardata"`
                RetiredCount    string `xml:"retired_count"`
                RetiredPagelist string `xml:"retired_pagelist"`
            } `xml:"multiple_single_bit_retirement"`
            DoubleBitRetirement struct {
                Text            string `xml:",chardata"`
                RetiredCount    string `xml:"retired_count"`
                RetiredPagelist string `xml:"retired_pagelist"`
            } `xml:"double_bit_retirement"`
            PendingBlacklist  string `xml:"pending_blacklist"`
            PendingRetirement string `xml:"pending_retirement"`
        } `xml:"retired_pages"`
        RemappedRows string `xml:"remapped_rows"`
        Temperature  struct {
            Text                   string `xml:",chardata"`
            GpuTemp                string `xml:"gpu_temp"`
            GpuTempMaxThreshold    string `xml:"gpu_temp_max_threshold"`
            GpuTempSlowThreshold   string `xml:"gpu_temp_slow_threshold"`
            GpuTempMaxGpuThreshold string `xml:"gpu_temp_max_gpu_threshold"`
            GpuTargetTemperature   string `xml:"gpu_target_temperature"`
            MemoryTemp             string `xml:"memory_temp"`
            GpuTempMaxMemThreshold string `xml:"gpu_temp_max_mem_threshold"`
        } `xml:"temperature"`
        SupportedGpuTargetTemp struct {
            Text             string `xml:",chardata"`
            GpuTargetTempMin string `xml:"gpu_target_temp_min"`
            GpuTargetTempMax string `xml:"gpu_target_temp_max"`
        } `xml:"supported_gpu_target_temp"`
        PowerReadings struct {
            Text               string `xml:",chardata"`
            PowerState         string `xml:"power_state"`
            PowerManagement    string `xml:"power_management"`
            PowerDraw          string `xml:"power_draw"`
            PowerLimit         string `xml:"power_limit"`
            DefaultPowerLimit  string `xml:"default_power_limit"`
            EnforcedPowerLimit string `xml:"enforced_power_limit"`
            MinPowerLimit      string `xml:"min_power_limit"`
            MaxPowerLimit      string `xml:"max_power_limit"`
        } `xml:"power_readings"`
        Clocks struct {
            Text          string `xml:",chardata"`
            GraphicsClock string `xml:"graphics_clock"`
            SmClock       string `xml:"sm_clock"`
            MemClock      string `xml:"mem_clock"`
            VideoClock    string `xml:"video_clock"`
        } `xml:"clocks"`
        ApplicationsClocks struct {
            Text          string `xml:",chardata"`
            GraphicsClock string `xml:"graphics_clock"`
            MemClock      string `xml:"mem_clock"`
        } `xml:"applications_clocks"`
        DefaultApplicationsClocks struct {
            Text          string `xml:",chardata"`
            GraphicsClock string `xml:"graphics_clock"`
            MemClock      string `xml:"mem_clock"`
        } `xml:"default_applications_clocks"`
        MaxClocks struct {
            Text          string `xml:",chardata"`
            GraphicsClock string `xml:"graphics_clock"`
            SmClock       string `xml:"sm_clock"`
            MemClock      string `xml:"mem_clock"`
            VideoClock    string `xml:"video_clock"`
        } `xml:"max_clocks"`
        MaxCustomerBoostClocks struct {
            Text          string `xml:",chardata"`
            GraphicsClock string `xml:"graphics_clock"`
        } `xml:"max_customer_boost_clocks"`
        ClockPolicy struct {
            Text             string `xml:",chardata"`
            AutoBoost        string `xml:"auto_boost"`
            AutoBoostDefault string `xml:"auto_boost_default"`
        } `xml:"clock_policy"`
        Voltage struct {
            Text         string `xml:",chardata"`
            GraphicsVolt string `xml:"graphics_volt"`
        } `xml:"voltage"`
        SupportedClocks struct {
            Text              string `xml:",chardata"`
            SupportedMemClock []struct {
                Text                   string   `xml:",chardata"`
                Value                  string   `xml:"value"`
                SupportedGraphicsClock []string `xml:"supported_graphics_clock"`
            } `xml:"supported_mem_clock"`
        } `xml:"supported_clocks"`
        Processes struct {
            Text        string `xml:",chardata"`
            ProcessInfo struct {
                Text              string `xml:",chardata"`
                GpuInstanceID     string `xml:"gpu_instance_id"`
                ComputeInstanceID string `xml:"compute_instance_id"`
                Pid               string `xml:"pid"`
                Type              string `xml:"type"`
                ProcessName       string `xml:"process_name"`
                UsedMemory        string `xml:"used_memory"`
            } `xml:"process_info"`
        } `xml:"processes"`
        AccountedProcesses string `xml:"accounted_processes"`
    } `xml:"gpu"`
}

