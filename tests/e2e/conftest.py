import pytest
from xp_client import XPClient


def pytest_addoption(parser):
    parser.addoption(
        "--management-url", action="store", help="management service url", default=""
    )
    parser.addoption(
        "--treatment-url", action="store", help="treatment service url", default=""
    )
    parser.addoption(
        "--management-bin",
        action="store",
        help="path to management service binary, if url is not provided",
        default="",
    )
    parser.addoption(
        "--treatment-bin",
        action="store",
        help="path to treatment service binary, if url is not provided",
        default="",
    )
    parser.addoption("--env", action="store", help="path to env", default="local")


from fixtures.mockups.mlp_service import *  # noqa
from fixtures.services import *  # noqa


@pytest.fixture
def xp_client(xp_management, xp_treatment):
    return XPClient(f'{xp_management.rstrip("/")}/v1', f'{xp_treatment.rstrip("/")}/v1')
