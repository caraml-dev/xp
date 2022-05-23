import os
import socket
import subprocess
import time
from os import PathLike
from os.path import isfile
from pathlib import Path
from typing import List, Optional, Tuple, Union

import pytest
import requests
import yaml
from _pytest.fixtures import FixtureRequest
from dotenv import load_dotenv
from google.api_core.exceptions import AlreadyExists
from google.cloud import pubsub_v1

MANAGEMENT_SERVICE_NAME = "management-service"
TREATMENT_SERVICE_NAME = "treatment-service"

SERVICE_TO_BINARY_NAME = {
    MANAGEMENT_SERVICE_NAME: "xp-management",
    TREATMENT_SERVICE_NAME: "xp-treatment",
}

PROJECT_ROOT_DIR = Path(__file__).parents[3]
TEST_DIR = Path.joinpath(PROJECT_ROOT_DIR, "tests", "e2e")


def _service_dir(service_name) -> Path:
    return Path.joinpath(PROJECT_ROOT_DIR, service_name)


def _default_bin_path(service_name) -> Path:
    return Path.joinpath(
        _service_dir(service_name), "bin", SERVICE_TO_BINARY_NAME[service_name]
    )


def _wait_port_open(host, port, max_wait=60):
    print(f"Waiting for port {port}")
    start = time.time()

    while True:
        try:
            socket.create_connection((host, port), timeout=1)
        except OSError:
            if time.time() - start > max_wait:
                raise

            time.sleep(1)
        else:
            return


def _start_binary(
    binary: Union[PathLike, str],
    options: List[str] = None,
    working_dir: Optional[PathLike] = None,
):
    if not isfile(binary):
        raise ValueError(f"The binary file '{binary}' doesn't exist")

    cmd = [binary]
    if options:
        cmd.extend(options)

    return subprocess.Popen(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=working_dir
    )


def create_pubsub_topic(project_id: str, topic: str):
    publisher = pubsub_v1.PublisherClient()
    topic_name = f"projects/{project_id}/topics/{topic}"
    try:
        publisher.create_topic(name=topic_name)
    except AlreadyExists:
        pass


def _service_config(env: str, service_name: str) -> PathLike:
    return Path.joinpath(TEST_DIR, "config", f"{service_name}.{env}.yaml")


def _env_config(env: str) -> PathLike:
    return Path.joinpath(TEST_DIR, "config", f"{env}.env")


def _create_default_xp_project(xp_management_url: str):
    default_project = {
        "username": "user1",
        "randomization_key": "order_id",
        "segmenters": ["s2_ids"],
    }
    response = requests.post(
        f"{xp_management_url}/v1/projects/1/settings", json=default_project
    )
    if response.status_code != 200:
        print(response.status_code)
        print(response.content)
        raise ValueError("unable to create default project")


def _start_management_service(
    bin_path: PathLike, env: str, mlp_service_url
) -> Tuple[str, subprocess.Popen]:
    load_dotenv(_env_config(env))
    config_path = _service_config(env, MANAGEMENT_SERVICE_NAME)

    with open(config_path, "r") as config_file:
        config = yaml.safe_load(config_file)
    pubsub_config = config.get("PubSubConfig", {})
    pubsub_project = pubsub_config.get("Project", "dev")
    pubsub_topic = pubsub_config.get("TopicName", "xp-update")
    create_pubsub_topic(pubsub_project, pubsub_topic)

    os.environ["MLPCONFIG::URL"] = mlp_service_url

    process = _start_binary(
        bin_path,
        ["serve", "--config", config_path],
        working_dir=_service_dir(MANAGEMENT_SERVICE_NAME),
    )

    try:
        port = config.get("Port", 3000)
        _wait_port_open("localhost", port, 15)
    except OSError:
        outs, errs = process.communicate(timeout=5)
        print(outs)
        print(errs)
        raise ValueError("unable to run management service binary")

    xp_management_url = f"http://localhost:{port}"

    return xp_management_url, process


def _start_treatment_service(
    bin_path: PathLike, env: str
) -> Tuple[str, subprocess.Popen]:
    load_dotenv(_env_config(env))

    config_path = _service_config(env, TREATMENT_SERVICE_NAME)
    process = _start_binary(bin_path, ["serve", "--config", config_path])

    with open(config_path, "r") as config_file:
        config = yaml.safe_load(config_file)

    try:
        port = config.get("Port", 8080)
        _wait_port_open("localhost", port, 15)
    except OSError:
        outs, errs = process.communicate(timeout=5)
        print(outs)
        print(errs)
        raise ValueError("unable to run treatment service binary")

    xp_treatment_url = f"http://localhost:{port}"
    return xp_treatment_url, process


@pytest.fixture(scope="session")
def xp_management(pytestconfig, request: FixtureRequest):
    xp_management_url = pytestconfig.getoption("management_url")
    process = None
    if xp_management_url == "":
        bin_path = pytestconfig.getoption("management_bin")
        if bin_path == "":
            bin_path = _default_bin_path(MANAGEMENT_SERVICE_NAME)
        env = pytestconfig.getoption("env")
        xp_management_url, process = _start_management_service(
            bin_path, env, mlp_service_url=request.getfixturevalue("mlp_service")
        )
    yield xp_management_url
    if process:
        process.terminate()


@pytest.fixture(scope="session")
def xp_treatment(pytestconfig, xp_management):
    xp_treatment_url = pytestconfig.getoption("treatment_url")
    process = None
    if xp_treatment_url == "":
        bin_path = pytestconfig.getoption("treatment_bin")
        if bin_path == "":
            bin_path = _default_bin_path(TREATMENT_SERVICE_NAME)
        env = pytestconfig.getoption("env")
        xp_treatment_url, process = _start_treatment_service(bin_path, env)
    yield xp_treatment_url
    if process:
        process.terminate()
