<div id="top"></div>

<h3 align="center">Kubernetes Cluster Introspection Tool</h3>

The purpose of this tool is to capture a detailed cluster inventory, including GPU data, package it into a transportable file/data structure, and make this information easily accessible. This tool is compatible with various Kubernetes deployment options (on prem, cloud, OpenShift, Tanzu, etc.), as well as many GPU configurations (vGPU, MIG, etc.). A more detailed design spec for this application can be found [here](https://docs.google.com/document/d/1vIZLLR46bY93l-tIpZa80LSDiUh_eNs7EQirZhB1Bx8/edit?usp=sharing).
</div>



<!-- TABLE OF CONTENTS -->
## Table of Contents
<ol>
  <li><a href="#prerequisites">Prerequisites</a></li>
  <li><a href="#installation">Installation</a></li>
  <li><a href="#usage">Usage</a></li>
</ol>

## Prerequisites

Before installing, ensure that the NVIDIA GPU Operator is installed and running smoothly on your cluster. More info on the GPU Operator can be found [here](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/overview.html).

## Installation

Various options can be specified at installation time through helm:
- Data Scrape Rate (Default: 24 hours)
    * Helm chart variable name: `rate` (integer)
    * Measured in hours
- Deploy Web Interface (Default: true)
    * Helm chart variable name: `web` (boolean)
- NodePort Value (Default: 30069)
    * Helm chart variable name: `nodePort` (integer)
- Path To Output File (Default: “inventory.csv”)
    * Helm chart variable name: `path` (string)
    * This variable must specify a file, not a directory. If the specified file does not exist, it will be created. If it already exists, the inventory data will be appended to it.

1. Add Helm Repo
   ```sh
   helm repo add introspection https://mattfeinberg.github.io/K8s-Introspection-Tool/helm/
   helm repo update

   ```
2. Install Helm Chart
   ```sh
   helm install introspec-tool introspection/introspection-chart \
   --create-namespace --namespace monitoring
   ```
This will use the default installation settings, but you can customize your installation by adding `--set <variable>=<value>` to the end of the installation command. This command will also deploy the introspection application in the monitoring namespace, and will create the namespace if it does not already exist. There are no restrictions on this, so feel free to deploy in another namespace if desired.

<!-- USAGE EXAMPLES -->
## Usage

To access the cluster inventory web page (if enabled), naviagte to <machine-IP>:<nodePort> on your web browser, where <machine-IP> is the IP address of any node. The default NodePort value is 30069.

To access the cluster inventory data file, run:

```sh
kubectl cp <namespace>/<pod>:<path to file> <destination path>
```

The default namespace is monitoring, and the default path is `/root/inventory.csv`. The <pod> name will change with each deployment, and you can find it by searching for the `introspection-tool` pod in the output of:

```sh
kubectl get pods -n monitoring
```
