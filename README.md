# xp

XP helps with designing and managing experiment configurations in a safe and holistic manner. At runtime, these configurations can be used to run experiments and generate treatments.

The API is broken down into 2 services:
- **Management Service**: Used to configure experiments
- **Treatment Service**: Used to obtain the treatment configuration from active experiments

### Why XP?

- **Reliable** - Inherent fault-detection rules help create experiments without conflicts.
- **Customizable** - Every service has unique requirements. XP allows for defining flexible segmentation and experiment validation rules.
- **Fast** - 99p server-side latency (excluding the network latency between the calling service and XP) averages around 1 ms.
- **Observable** - Resource utilization, treatment assignment and performance metrics are available on Prometheus

## Getting Started

### Setup MLP API

Instructions as described in README of https://github.com/gojek/mlp.

### Starting XP locally

Prior to starting XP, we'll need to ensure correct MLP API is correctly set in the config file, i.e management-service/config/example.yaml, and setting the correct `MLPConfig::URL` value.

```
// Start Management Service
make mgmt-svc

// Start Treatment Service
make treatment-svc
```
