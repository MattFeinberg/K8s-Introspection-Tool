<div id="top"></div>

<h3 align="center">Kubernetes Cluster Introspection Tool</h3>

### Note: README is unfinished

<p align="center">
This tool allows you to gather a detailed inventory on a GPU-equipped Kubernetes cluster. You can deploy the tool through a helm chart, and view the cluster inventory via HTTP or through a generated JSON file.
</p>
</div>



<!-- TABLE OF CONTENTS -->
## Table of Contents
<ol>
  <li><a href="#prerequisites">Prerequisites</a></li>
  <li><a href="#installation">Installation</a></li>
  <li><a href="#usage">Usage</a></li>
</ol>

## Prerequisites

Before installing, ensure that the NVIDIA GPU Operator is installed and running smoothly on your cluster. More info on the GPU Operator can be found here: https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/overview.html

## Installation

1. Clone the repo
   ```sh
   git clone https://github.com/MattFeinberg/K8s-Introspection-Tool.git
   ```
2. Install helm chart
   ```sh
   helm install
   ```
3. ...
   ```sh
   next steps
   ```

<!-- USAGE EXAMPLES -->
## Usage

The cluster inventory web interface, if enabled through helm, is availible after installation at <machine-ip>:30069, or at the specified NodePort at installation.






<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
