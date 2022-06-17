# xp-management

---
![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square)
![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

A Helm chart for Kubernetes Deployment of the XP Managment Service

## Introduction

This Helm chart installs [Management Service](https://github.com/gojek/xp/management-service) and all its dependencies in a Kubernetes cluster.

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

This command will install XP Management Service release named `management-service` in the `default` namespace.
Default chart values will be used for the installation:
```shell
$ helm install xp xp/management-service
```

You can (and most likely, should) override the default configuration with values suitable for your installation.
Refer to [Configuration](#configuration) section for the detailed description of available configuration keys.

You can also refer to [values.minimal.yaml](./values.minimal.yaml) to check a minimal configuration that needs
to be provided for XP Management Service installation.

### Uninstalling the chart

To uninstall `management-service` release:
```shell
$ helm uninstall management-service
```

The command removes all the Kubernetes components associated with the chart and deletes the release,
except for postgresql PVC, those will have to be removed manually.

## Configuration

The following table lists the configurable parameters of the XP Management Service chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| global.mlp.encryption.key | string | `nil` | Global MLP Encryption Key to be used by all MLP components |
| global.sentry.dsn | string | `nil` | Global Sentry DSN value |
| swaggerUi.apiServer | string | `"http://127.0.0.1/v1"` | URL of API server |
| swaggerUi.image | object | `{"tag":"v3.47.1"}` | Docker tag for Swagger UI https://hub.docker.com/r/swaggerapi/swagger-ui |
| swaggerUi.service.externalPort | int | `8080` | Swagger UI Kubernetes service port number |
| swaggerUi.service.internalPort | int | `8081` | Swagger UI container port number |
| xpManagement.apiConfig | object | `{"sentryConfig":{"enabled":false}}` | XP Management Service server configuration. |
| xpManagement.extraArgs | list | `[]` | List of string containing additional XP Management Service server arguments. For example, multiple "-config" can be specified to use multiple config files |
| xpManagement.extraEnvs | list | `[]` | List of extra environment variables to add to XP Management Service server container |
| xpManagement.extraLabels | object | `{}` | List of extra labels to add to XP Management Service K8s resources |
| xpManagement.extraVolumeMounts | list | `[]` | Extra volume mounts to attach to XP Management Service server container. For example to mount the extra volume containing secrets |
| xpManagement.extraVolumes | list | `[]` | Extra volumes to attach to the Pod. For example, you can mount  additional secrets to these volumes |
| xpManagement.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| xpManagement.image.registry | string | `"docker.io/"` | Docker registry for XP Management Service image |
| xpManagement.image.repository | string | `"xp-management"` | Docker image repository for XP Management Service |
| xpManagement.image.tag | string | `"latest"` | Docker image tag for XP Management Service |
| xpManagement.ingress.class | string | `""` | Ingress class annotation to add to this Ingress rule, useful when there are multiple ingress controllers installed |
| xpManagement.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to XP Management Service |
| xpManagement.ingress.host | string | `""` | Set host value to enable name based virtual hosting. This allows routing HTTP traffic to multiple host names at the same IP address. If no host is specified, the ingress rule applies to all inbound HTTP traffic through  the IP address specified. https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting |
| xpManagement.labels | object | `{}` |  |
| xpManagement.livenessProbe.initialDelaySeconds | int | `60` | Liveness probe delay and thresholds |
| xpManagement.livenessProbe.path | string | `"/v1/internal/live"` | HTTP path for liveness check |
| xpManagement.livenessProbe.periodSeconds | int | `10` |  |
| xpManagement.livenessProbe.successThreshold | int | `1` |  |
| xpManagement.livenessProbe.timeoutSeconds | int | `5` |  |
| xpManagement.readinessProbe.initialDelaySeconds | int | `60` | Liveness probe delay and thresholds |
| xpManagement.readinessProbe.path | string | `"/v1/internal/ready"` | HTTP path for readiness check |
| xpManagement.readinessProbe.periodSeconds | int | `10` |  |
| xpManagement.readinessProbe.successThreshold | int | `1` |  |
| xpManagement.readinessProbe.timeoutSeconds | int | `5` |  |
| xpManagement.replicaCount | int | `1` |  |
| xpManagement.resources | object | `{}` | Resources requests and limits for XP Management Service. This should be set according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| xpManagement.sentry.dsn | string | `""` | Sentry DSN value used by both XP Management Service and XP UI |
| xpManagement.sentry.enabled | bool | `false` |  |
| xpManagement.service.externalPort | int | `8080` | XP Management Service Kubernetes service port number |
| xpManagement.service.internalPort | int | `8080` | XP Management Service container port number |
| xpManagement.serviceAccount.annotations | object | `{}` |  |
| xpManagement.serviceAccount.create | bool | `true` |  |
| xpManagement.serviceAccount.name | string | `""` |  |
| xpManagement.uiConfig | object | `{"apiConfig":{"mlpApiUrl":"/api/v1","xpApiUrl":"/api/xp/v1"},"appConfig":{"docsUrl":[{"href":"https://github.com/gojek/xp/tree/main/docs","label":"XP User Guide"}]},"authConfig":{"oauthClientId":""},"sentryConfig":{}}` | XP UI configuration. |
