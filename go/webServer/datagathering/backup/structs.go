package datagathering


//CLUSTER STRUCT
type ClusterInfo struct {
    Version           string
    NumNodes          int
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
    NumGPUs          string
    Provider         string
    DriverVersion    string   `xml:"driver_version"`
    CudaVersion      string   `xml:"cuda_version"`
    AttachedGpus     int      `xml:"attached_gpus"`
    AttachedNICs     int
    NVSMIExitCode    int
    VGPUHostDriver   string
    NICs             []string
    //TODO: This is GPU-level, I ahve it at node level rn
    //MIGProfiles      string
    GPUs             []struct{
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
        MigDevices struct {
            Text      string `xml:",chardata"`
            MigDevice []struct {
                Text              string `xml:",chardata"`
                Index             string `xml:"index"`
                GpuInstanceID     string `xml:"gpu_instance_id"`
                ComputeInstanceID string `xml:"compute_instance_id"`
                DeviceAttributes  struct {
                    Text   string `xml:",chardata"`
                    Shared struct {
                        Text                string `xml:",chardata"`
                        MultiprocessorCount string `xml:"multiprocessor_count"`
                        CopyEngineCount     string `xml:"copy_engine_count"`
                        EncoderCount        string `xml:"encoder_count"`
                        DecoderCount        string `xml:"decoder_count"`
                        OfaCount            string `xml:"ofa_count"`
                        JpgCount            string `xml:"jpg_count"`
                    } `xml:"shared"`
                } `xml:"device_attributes"`
                EccErrorCount struct {
                    Text          string `xml:",chardata"`
                    VolatileCount struct {
                        Text              string `xml:",chardata"`
                        SramUncorrectable string `xml:"sram_uncorrectable"`
                    } `xml:"volatile_count"`
                } `xml:"ecc_error_count"`
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
            } `xml:"mig_device"`
        } `xml:"mig_devices"`
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
            VgpuProfile        string
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

