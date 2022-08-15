# xp-treatment

---
![Version: 0.1.2](https://img.shields.io/badge/Version-0.1.2-informational?style=flat-square)
![AppVersion: v0.7.0](https://img.shields.io/badge/AppVersion-v0.7.0-informational?style=flat-square)

A Helm chart for Kubernetes Deployment of the XP Treatment Service

## Introduction

This Helm chart installs [Treatment Service](https://github.com/gojek/xp/treatment-service) and all its dependencies in a Kubernetes cluster.

## Prerequisites

To use the charts here, [Helm](https://helm.sh/) must be configured for your
Kubernetes cluster. Setting up Kubernetes and Helm is outside the scope of
this README. Please refer to the Kubernetes and Helm documentation.

- **Helm 3.0+** – This chart was tested with Helm v3.7.1, but it is also expected to work with earlier Helm versions
- **Kubernetes 1.18+** – This chart was tested with GKE v1.20.x, but it's possible it works with earlier k8s versions too.

## Installation

### Add Helm repository

```sh
$ helm repo add xp https://turing-ml.github.io/charts
```

### Installing the chart

This command will install XP Treatment Service release named `treatment-service` in the `default` namespace.
Default chart values will be used for the installation:
```shell
$ helm install xp xp/treatment-service
```

You can (and most likely, should) override the default configuration with values suitable for your installation.
Refer to [Configuration](#configuration) section for the detailed description of available configuration keys.

You can also refer to [values.minimal.yaml](./values.minimal.yaml) to check a minimal configuration that needs
to be provided for XP Treatment Service installation.

### Uninstalling the chart

To uninstall `treatment-service` release:
```shell
$ helm uninstall treatment-service
```

The command removes all the Kubernetes components associated with the chart and deletes the release,
except for postgresql PVC, those will have to be removed manually.

## Configuration

The following table lists the configurable parameters of the XP Treatment Service chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| swaggerUi.apiServer | string | `"http://127.0.0.1/v1"` | URL of API server |
| swaggerUi.enabled | bool | `false` |  |
| swaggerUi.image | object | `{"tag":"v3.47.1"}` | Docker tag for Swagger UI https://hub.docker.com/r/swaggerapi/swagger-ui |
| swaggerUi.service.externalPort | int | `8080` | Swagger UI Kubernetes service port number |
| swaggerUi.service.internalPort | int | `8081` | Swagger UI container port number |
| xpTreatment.autoscaling.enabled | bool | `false` |  |
| xpTreatment.autoscaling.maxReplicas | int | `2` |  |
| xpTreatment.autoscaling.minReplicas | int | `1` |  |
| xpTreatment.autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| xpTreatment.autoscaling.targetMemoryUtilizationPercentage | int | `80` |  |
| xpTreatment.extraEnvs | list | `[]` | List of extra environment variables to add to XP Treatment Service server container |
| xpTreatment.extraLabels | object | `{}` | List of extra labels to add to XP Treatment Service K8s resources |
| xpTreatment.extraVolumeMounts | list | `[]` | Extra volume mounts to attach to XP Treatment Service server container. For example to mount the extra volume containing secrets |
| xpTreatment.extraVolumes | list | `[]` | Extra volumes to attach to the Pod. For example, you can mount  additional secrets to these volumes |
| xpTreatment.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| xpTreatment.image.registry | string | `"ghcr.io"` | Docker registry for XP Treatment Service image |
| xpTreatment.image.repository | string | `"gojek/turing-experiments/xp-treatment"` | Docker image repository for XP Treatment Service |
| xpTreatment.image.tag | string | `"v0.7.0"` | Docker image tag for XP Treatment Service |
| xpTreatment.ingress.class | string | `""` | Ingress class annotation to add to this Ingress rule,  useful when there are multiple ingress controllers installed |
| xpTreatment.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to XP Treatment Service |
| xpTreatment.ingress.host | string | `""` | Set host value to enable name based virtual hosting. This allows routing HTTP traffic to multiple host names at the same IP address. If no host is specified, the ingress rule applies to all inbound HTTP traffic through  the IP address specified. https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting |
| xpTreatment.labels | object | `{}` |  |
| xpTreatment.livenessProbe.initialDelaySeconds | int | `60` | Liveness probe delay and thresholds |
| xpTreatment.livenessProbe.path | string | `"/v1/internal/health/live"` | HTTP path for liveness check |
| xpTreatment.livenessProbe.periodSeconds | int | `10` |  |
| xpTreatment.livenessProbe.successThreshold | int | `1` |  |
| xpTreatment.livenessProbe.timeoutSeconds | int | `5` |  |
| xpTreatment.nodeSelector | object | `{}` |  |
| xpTreatment.readinessProbe.initialDelaySeconds | int | `60` | Liveness probe delay and thresholds |
| xpTreatment.readinessProbe.path | string | `"/v1/internal/health/ready"` | HTTP path for readiness check |
| xpTreatment.readinessProbe.periodSeconds | int | `10` |  |
| xpTreatment.readinessProbe.successThreshold | int | `1` |  |
| xpTreatment.readinessProbe.timeoutSeconds | int | `5` |  |
| xpTreatment.replicaCount | int | `1` |  |
| xpTreatment.resources | object | `{}` | Resources requests and limits for XP Treatment Service. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| xpTreatment.service.externalPort | int | `8080` | XP Treatment Service Kubernetes service port number |
| xpTreatment.service.internalPort | int | `8080` | XP Treatment Service container port number |
| xpTreatment.service.type | string | `"ClusterIP"` |  |
