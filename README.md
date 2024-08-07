# Tilt and Helm Integration with Go Service, Prometheus, and Grafana

This project is a demonstration of how to use Tilt with Helm to manage a Go service that integrates with Prometheus for monitoring and Grafana for visualization. The setup allows for rapid local development and testing with a Kubernetes environment.

## Project Overview

The purpose of this project is to showcase the use of Tilt with Helm for deploying and managing a Go service in a Kubernetes environment. It includes:

- A Go service that exposes Prometheus metrics.
- Prometheus for scraping and storing metrics.
- Grafana for visualizing metrics.
- Helm charts for deploying the Go service, Prometheus, and Grafana.

## Prerequisites

- [Docker](https://www.docker.com/)
- [Kubernetes](https://kubernetes.io/)
- [Helm](https://helm.sh/)
- [Tilt](https://tilt.dev/)
- [Go](https://golang.org/)

## Usage

To start the development environment using Tilt, run:

### Starting Tilt

To start the development environment using Tilt, run:

```sh
tilt up
```