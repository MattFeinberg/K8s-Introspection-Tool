package datagathering


//CLUSTER STRUCT
type ClusterInfo struct {
    K8sVersion        string
    CloudOrOnPrem     string
    CSP               string
    NumNodes          int
    // TODO: extra storing names here, but mayeb its useful
    NodeNames         []string
    NumUnhealthyNodes int
    Distribution      string
    TotalGPUs         int
    TotalNICs         int
    TotalMIG          int
    TimeUpdated       string
    GPUDist           map[string]int
    Nodes             []NodeInfo
}


//TODO: Delete unneeded elements of this struct

//NODE STRUCT
type NodeInfo struct {
    NodeName         string
    OSVersion        string
    OS               string
    VMOrBareMetal    string
    KubeletVersion   string
    ContainerRuntime string
    DriverVersion    string   `xml:"driver_version"`
    CudaVersion      string   `xml:"cuda_version"`
    AttachedGPUs     int      `xml:"attached_gpus"`
    AttachedNICs     int
    NVSMIExitCode    int
    VGPUHostDriver   string
    NICs             []string
    //TODO: This is GPU-level, I ahve it at node level rn
    //MIGProfiles      string
    GPUs             []struct{
        ProductName           string `xml:"product_name"`
        GPUPartNumber         string `xml:"gpu_part_number"`
        GPUSerialNumber       string `xml:"serial"`
        GSPFirmwareVersion    string `xml:"gsp_firmware_version"`
        GpuVirtualizationMode struct {
            VirtualizationMode string `xml:"virtualization_mode"`
            VGPUProfile        string
        } `xml:"gpu_virtualization_mode"`
        FbMemoryUsage         struct {
            Total    string `xml:"total"`
            Reserved string `xml:"reserved"`
            Used     string `xml:"used"`
            Free     string `xml:"free"`
        } `xml:"fb_memory_usage"`
        Bar1MemoryUsage       struct {
            Total string `xml:"total"`
            Used  string `xml:"used"`
            Free  string `xml:"free"`
        } `xml:"bar1_memory_usage"`
        MIGInfo               struct {
            MIGStatus string `xml:"current_mig"`
        } `xml:"mig_mode"`
        MIGDevices            struct {
            MIGDevice []struct {
                GPUInstanceID     string `xml:"gpu_instance_id"`
                ComputeInstanceID string `xml:"compute_instance_id"`
                DeviceAttributes  struct {
                    Shared struct {
                        MultiprocessorCount string `xml:"multiprocessor_count"`
                        CopyEngineCount     string `xml:"copy_engine_count"`
                        EncoderCount        string `xml:"encoder_count"`
                        DecoderCount        string `xml:"decoder_count"`
                        OfaCount            string `xml:"ofa_count"`
                        JpgCount            string `xml:"jpg_count"`
                    } `xml:"shared"`
                } `xml:"device_attributes"`
                FbMemoryUsage     struct {
                    Total    string `xml:"total"`
                    Reserved string `xml:"reserved"`
                    Used     string `xml:"used"`
                    Free     string `xml:"free"`
                } `xml:"fb_memory_usage"`
                Bar1MemoryUsage   struct {
                    Total string `xml:"total"`
                    Used  string `xml:"used"`
                    Free  string `xml:"free"`
                } `xml:"bar1_memory_usage"`
            } `xml:"mig_device"`
        } `xml:"mig_devices"`
    } `xml:"gpu"`
}

