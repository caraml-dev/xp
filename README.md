# xp

[![License](https://img.shields.io/badge/License-Apache%202.0-blue)](https://github.com/feast-dev/feast/blob/master/LICENSE)

## Overview

XP helps with designing and managing experiment configurations in a safe and holistic manner. At runtime, these configurations can be used to run experiments and generate treatments.

The API is broken down into 2 services:

- **Management Service**: Used to configure experiments
- **Treatment Service**: Used to obtain the treatment configuration from active experiments

## Why XP?

- **Reliable** - Inherent fault-detection rules help create experiments without conflicts.
- **Customizable** - Every service has unique requirements. XP allows for defining flexible segmentation and experiment validation rules.
- **Fast** - 99p server-side latency (excluding the network latency between the calling service and XP) averages around 1 ms.
- **Observable** - Resource utilization, treatment assignment and performance metrics are available on Prometheus

## Development Environment

### Quick Start

#### a. Setup MLP API

Instructions as described in README of https://github.com/gojek/mlp.

#### b. Starting XP

Prior to starting XP, we'll need to ensure correct MLP API is correctly set in the config file, i.e management-service/config/example.yaml, and setting the correct `MLPConfig::URL` value.

```bash
# Start Management Service
make mgmt-svc

# Start Treatment Service
make treatment-svc

# Exploring Swagger-UI
make swagger-ui
```

`make mgmt-svc` runs the following:

1. `make local-db`
    - Setup a local DB for storing experiment configurations
2. `make local-authz-server`
    - Setup AuthZ server that is accessible at http://localhost:4466/.

To test authorization for Management Service locally, make the following changes before starting Management Service:

- A sample policy exists at keto/policies/example_policy.json. It can be modified.
- Set AuthorizationConfig.Enabled=true in the config file that's being used
- Issue requests to the app with the header User-Email: test-user@gojek.com

### API Specifications

The OpenAPI specs for both services are captured in the `api/` folder. If these specs are updated, the developer is required to regenerate the API types and interfaces using the command `make generate-api`.

### Tests

For **Unit** tests, we follow the convention of keeping it beside the main source file.

For **Integration** tests, they are available for Treatment Service currently where we mock certain functionality of Management Service under `treatment-service/testhelper/mockmanagement` and utilize them in `treatment-service/integration-test`.

**End-to-End** tests can be found in tests/e2e, where we build Management Service and Treatment Service binaries to be used in Pytest.

### Code style guidelines

We are using [golangci-lint](https://github.com/golangci/golangci-lint), and we can run the following commands for formatting.

```bash
# Formatting code
make fmt

# Checking for linting issues
make lint
```
