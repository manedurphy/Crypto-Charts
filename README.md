# Crypto Charts

## What this Application Does
This application gathers information on cryptocurrencies (Bitcoin, Ethereum, Doge) from the [crypto compare](https://www.cryptocompare.com/) API, and delivers that data to a frontend which displays the data on charts

## The Purpose
The purpose of this project was to better my understanding of various technologies that I have been learning. This includes working with `Protocol Buffers` to define an API, the `gRPC` framework, and `Kubernetes`.

## How it Works
The client, a ReactJS frontend, makes a request for the data it needs to display the charts. Because the browser cannot make a direct request to the gRPC server, I used the gRPC Gateway [library](https://github.com/grpc-ecosystem/grpc-gateway) to forward the request to the gRPC server over a TLS connection. The gRPC server then makes a call to the external API for the data needed, manipulates it and stores that information in a Redis store so that it does not need to continuously make calls to the external API.

# Kubernetes
My primary focus throughout the project was to experiment with various system architectures using Kubernetes. These are the working parts I have included in this application:

## Ingress Controller (Helm Chart)
When the browser makes the request for data, its request goes through an Nginx Ingress Controller that has defined the routes. A request to the root`"/"` path points to the Pod that serves the static HTML, CSS, and JS files. Once the JS has loaded, it makes a request to `"/api"`, where the Ingress Controller forwards that request to the gRPC gateway.

## Static Files (Deployment)
I have built a docker image of an Nginx server that serves the static files from my React build script.

## gRPC Gateway (Deployment)
The gRPC gateway receives requests from the Ingress Controller, and then to the gRPC server. It is the bridge between the server and client.

## Server (Deployment)
The server makes the request to the external API, shapes the data, and stores it in a Redis StatefulSet

## Redis (StatefulSet)
The Redis StatefulSet is configured with a master-slave system, where slave pods synchronize their data with the master pod. Another StatefulSet, the Sentinel, is also there to ensure that a new master pod is elected when the current one fails for whatever reason. These steps were to ensure a highly available Redis cluster.

## CSI-Secrets-Store (Helm Chart)
The [kubernetes-sigs/secrets-store-csi-driver](https://github.com/kubernetes-sigs/secrets-store-csi-driver) allows the user to store sensitive information externally from the cluster (or within). For this project, I used a `Vault` instance with the [hashicorp/vault-csi-provider](https://github.com/hashicorp/vault-csi-provider) to act as the bridge between the `CSI-Driver` and the `Vault` instance. The password the the Redis cluster is received from the `Vault` and mounted the file system of the `Server` pods.

# Future Features
1. Some time in the future I would like to add an option to see live data
2. Add more ways to visualize the data provided to the client


# Commands
To run locally in docker
```bash
make compose
```

To run in K8s using Kind
```bash
make crypto-charts
```
This creates the cluster and builds everything from scratch
