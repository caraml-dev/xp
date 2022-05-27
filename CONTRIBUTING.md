# Development Guide: Contributing to XP

Thank you for your interest in contributing to XP. This document provides some suggestions and guidelines on how you can get involved.

## Become a contributor

You can contribute to XP in several ways:

- Contribute to feature development for the XP codebase
- Report bugs
- Create articles and documentation for users and contributors
- Help others answer questions about XP

### Report bugs

Report a bug by creating an issue. Provide as much information as possible
on how to reproduce the bug.

Before submitting the bug report, please make sure there are no existing issues
with a similar bug report. You can search the existing issues for similar issues.

### Suggest features

If you have an idea to improve XP, submit a feature request. It will be good
to describe the use cases and how it will benefit XP users in your feature
request.

## Making a pull request

You can submit pull requests to fix bugs, add new features or improve our documentation.

Here are some considerations you should keep in mind when making changes:

- While making changes
  - Make your changes in a [forked repo](#forking-the-repo) (instead of making a branch on the main XP repo)
  - [Rebase from master](#incorporating-upstream-changes-from-master) instead of using `git pull` on your PR branch
  - Install [pre-commit hooks](#pre-commit-hooks) to ensure all the default linters / formatters are run when you push.
- When making the PR
  - Make a pull request from the forked repo you made
  - Ensure you leave a release note for any user facing changes in the PR. There is a field automatically generated in the PR request. You can write `NONE` in that field if there are no user facing changes.
  - Please run tests locally before submitting a PR:
    - For Go, the [unit tests](#go-tests).
    - For Python, the [e2e tests](#e2e-tests).

### Forking the repo

Fork the XP Github repo and clone your fork locally. Then make changes to a local branch to the fork.

See [Creating a pull request from a fork](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork)

### Pre-commit Hooks

Setup [`pre-commit`](https://pre-commit.com/) to automatically lint and format the codebase on commit:

1. Ensure that you have Python (3.7 and above) with `pip`, installed.
2. Install `pre-commit` with `pip` &amp; install pre-push hooks

    ```sh
    pip install pre-commit
    pre-commit install --hook-type pre-commit --hook-type pre-push
    ```

3. On push, the pre-commit hook will run. This runs `make format` and `make lint`.

## XP Management/Treatment Service using Go

Both Management & Treatment services are written using Go, and the following describes how to setup your development environment.

### Environment Setup

- Install Golang, [`protoc` with the Golang &amp; grpc plugins](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)

#### API Specifications

The OpenAPI specs for both services are captured in the `api/` folder. If these specs are updated, the developer is required to regenerate the API types and interfaces using the command `make generate-api`.

#### Compiling Protos

If there are proto changes required, you can recompile and generate them using `make compile-protos`.

### Code Style & Linting

We are using [golangci-lint](https://github.com/golangci/golangci-lint), and we can run the following commands for formatting.

```sh
# Formatting code
make fmt

# Checking for linting issues
make lint
```

### Go tests

For **Unit** tests, we follow the convention of keeping it beside the main source file.

For **Integration** tests, they are available for Treatment Service currently where we mock certain functionality of Management Service under `treatment-service/testhelper/mockmanagement` and utilize them in `treatment-service/integration-test`.

1. Run Management Service tests via `make test-management-service`.
2. Run Treatment Service tests via `make test-treatment-service`.

## XP E2E tests using Python

### Environment Setup

Setting up your development environment for E2E tests:

1. Ensure that you have `make`, Python (3.7 and above) with `pip`, installed.
2. _Recommended:_ Create a virtual environment to isolate development dependencies to be installed

    ```sh
    # Create & activate a virtual environment
    python -m venv venv/
    source venv/bin/activate
    ```

3. Install test dependencies

    ```sh
    pip install -r tests/requirements.txt
    ```

### Code Style & Linting

XP E2E tests:

- Conforms to [Black code style](https://black.readthedocs.io/en/stable/the_black_code_style.html)
- Has type annotations as enforced by `mypy`
- Has imports sorted by `isort`
- Is lintable by `flake8`

To ensure your Python code conforms to XP Python code standards:

- Autoformat your code to conform to the code style:

```sh
make format-python
```

- Lint your Python code before submitting it for review:

```sh
make lint-python
```

### E2E Tests

This constitutes building Management Service and Treatment Service binaries and starting them for tests.

1. Build Go services' binaries via `make build`.
2. Setup dependencies and run tests.
    - Docker-compose setup

        ```sh
        # Starts Postgres, PubSub Emulator and runs tests
        make e2e
        ```

    - Individual services setup
        - Start Postgres, Pubsub Emulator.
        - Run `cd tests/e2e; python -m pytest -s -v`
