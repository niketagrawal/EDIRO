# EDIRO

EDIRO is an Edge-driven IoT Resource-aware Orchestration Framework for Edge Computing. It enables an edge infrastructure comprising of a cluster of interconnected edge nodes to operate as an autonomous entity with minimal or no dependence on the cloud. The development of EDIRO is a means to experimentally evaluate the proposal of an edge-driven architecture for edge computing that aims to overcome the shortcomings of the existing approaches of utilizing edge computing for IoT. Please refer the [EDIRO paper](https://dl.acm.org/doi/abs/10.1145/3360468.3368179) for more details on this proposal and the description of the system architecture of EDIRO.


## Overview

EDIRO is specifically designed for utilization in IoT use cases characterized by client requests that request on-demand services from the edge infrastructure. The target use case dealt with in this regard is that of connected vehicles, where mobile clients, i.e., vehicles, demand  provisioning of ephemeral workloads on the edge servers. These workloads require an input or an ‘IoT resource’ to execute upon and render the desired results. EDIRO facilitates collaborative processing by enabling the network of edge nodes to source these IoT resources from other clients in the vicinity that may contribute them. In a typical connected vehicles scenario, a vehicle’s query about the condition of the road ahead of it is serviced by executing a workload on the nearby edge compute node installed on a traffic light that utilizes the input, i.e., high definition (HD) maps and sensor data offloaded on it by other vehicles in the vicinity.     


## System requirements for build and execution

EDIRO requires the following to be installed on the host machine.

- [A working Golang environment](https://golang.org/doc/install)
- [Docker Communty Edition](https://docs.docker.com/install/linux/docker-ce/ubuntu/)
- [gRPC and protoc plugin for Golang](https://grpc.io/docs/quickstart/go/)
