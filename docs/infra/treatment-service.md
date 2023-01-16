# Treatment Service Deployment

## Overview
XP's Treatment Service is currently offered in two different versions:

- a standalone Treatment Service, *it can be configured to serve treatments to multiple clients/users with multiple 
  CaraML projects*
- a Treatment Service plugin, *deployed as a [Turing Router](https://github.com/caraml-dev/turing) experiment engine 
  plugin, to serve treatments to a single client/user with a single CaraML project*

## Central Treatment Service

### Requirements
- Configuration
- Google Cloud Provider (GCP) service account

### Configuration
The central Treatment Service deployment uses a variety of configuration that can be stored in one or multiple 
configuration files (`.yaml`) files. Some of such examples can be found here:

- https://github.com/caraml-dev/xp/treatment-service/config/example.yaml
- https://github.com/caraml-dev/xp/treatment-service/testdata/config1.yaml
- https://github.com/caraml-dev/xp/treatment-service/testdata/config2.yaml

As can be seen from the various examples, there are a number of values that need to be set, though not all of them 
are necessarily required. Some of these values, when left undefined/empty, will be automatically initialised with 
certain default values (see https://github.com/caraml-dev/xp/blob/main/treatment-service/config/config.go#L22).

### Google Cloud Provider (GCP) Service Account
[Google Cloud Pub/Sub](https://cloud.google.com/pubsub/docs/overview) is used to allow the Treatment Service to 
communicate with the Management Service to retrieve information about the experiments that are being run at any point 
of time. More specifically, the Treatment Service subscribes to a Pub/Sub topic that the Management Service 
publishes to whenever there are updates to actively running experiments.

Furthermore, the Treatment Service may also interact with [Google BigQuery](https://cloud.google.com/bigquery), if 
it has been set up, to perform logging of all the treatments it assigns in a specified table.

Hence, if access control has been set up for the Pub/Sub topic or for the BigQuery table, the Treatment Service needs 
to be authenticated using a service account key that has the necessary rights to subscribe to the aforementioned Pub/Sub 
topic or to push data to the said BigQuery table. It does so by accessing a `.json` 
[file containing the service account key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys), 
whose location is stored as a filepath in the `GOOGLE_APPLICATION_CREDENTIALS` environment variable. 

Depending on how the Treatment Service is deployed, there are a variety of ways to ensure that it has access to 
the service account key file and the environment variable. See the section below for more information.

Note that the Treatment Service does not currently support the usage of multiple GCP service accounts for 
authentication, i.e the Pub/Sub topic and the BigQuery table each requiring a **different** service account for 
authentication.

### Deploying the Treatment Service

#### As a Helm Release

*Some of these steps have been adapted from https://github.com/caraml-dev/xp/tree/main/infra/charts/treatment-service*

##### 1. Add the Helm Repository

```sh
$ helm repo add xp https://turing-ml.github.io/charts
```

##### 2. Install the Helm Chart

This command will install XP Treatment Service release named `xp-treatment` in the `default` namespace.
Default chart values will be used for the installation:
```shell
$ helm install xp-treatment xp/xp-treatment
```

You can (and most likely, should) override the default configuration with Helm chart values suitable for your 
installation. Refer to [Configuration](https://github.com/caraml-dev/xp/tree/main/infra/charts/treatment-service#configuration) section for the detailed description of available configuration keys.

You can also refer to [values.yaml](https://github.com/caraml-dev/xp/tree/main/infra/charts/treatment-service/values.yaml)
for a minimal configuration that needs to be provided for XP Treatment Service installation.

```shell
$ helm install xp-treatment xp/xp-treatment \
    --values=path/to/helm/chart/values/file.yaml
```

Notice that you can specify the *Treatment Service configuration* [values](#configuration) under the `xpTreatment` 
field in the Helm chart values file like below: 

```yaml
xpTreatment:
  config:
    Port: 8080
    ManagementService:
      URL: https://caraml-dev.io/api/xp/v1
      AuthorizationEnabled: true
    ...
```

These configuration values would be saved in a `.yaml` file within a 
[secret](https://kubernetes.io/docs/concepts/configuration/secret/) that gets mounted automatically onto the 
Treatment Service pod, where it will be read by the Treatment Service.

##### 2.1 Configure the Treatment Service to use a Google Cloud Provider (GCP) Service Account (Optional) 

In particular, should you need to use a GCP service account for authenticating, one recommended way to set this up 
is to utilise [volumes](https://kubernetes.io/docs/concepts/storage/volumes/) to mount the GCP service account key file, 
which itself should be saved within a [secret](https://kubernetes.io/docs/concepts/configuration/secret/) in the 
Kubernetes cluster, into the Treatment Service [pod](https://kubernetes.io/docs/concepts/workloads/pods/). Adding the 
environment variable `GOOGLE_APPLICATION_CREDENTIALS` with the path to the mounted key file would then allow the 
Treatment Service to access it.

When deploying the Treatment Service with Helm, however, it is convenient to specify `extraVolumes`, 
`extraVolumeMounts` and `extraEnvs` in the Helm chart values (see
[values.yaml](https://github.com/caraml-dev/xp/tree/main/infra/charts/treatment-service/values.yaml#L59)), as they 
automatically get added to the Treatment 
Service deployment's volumes, volume mounts and environment variables respectively, via templates in the Helm chart 
(see 
[deployment.yaml](https://github.com/caraml-dev/xp/tree/main/infra/charts/treatment-service/templates/deployment.yaml#L37)).

*Note that this is NOT the only one way to set the GCP service account key file to allow the Treatment Service to
access it. There are a wide variety of other methods, such as containerising the Treatment Service binary together with 
the GCP service account key file, but we will not be listing all of them here.*

##### 3. Update the Helm Release

To subsequently update a deployed release of `xp-treatment`, you may run something like the following, which updates 
not only the `xp-treatment`'s image but also its config values:

```shell
$ helm upgrade xp-treatment xp/xp-treatment \
    --set xpTreatment.image.tag=${NEW_VERSION_TAG} \
    --values=path/to/updated/helm/chart/values/file.yaml
```

## Treatment Service Plugin 
TBC