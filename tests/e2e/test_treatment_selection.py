import copy
import datetime

import pytest
import s2cell
from data import BASE_EXPERIMENT, ID_S2_POINT
from utils import eventually, parse_datetime, wait_for
from xp_client import XPClient


@pytest.fixture()
def xp_project(xp_client: XPClient):
    p = xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="order_id",
        segmenters={
            "names": ["s2_ids", "days_of_week"],
            "variables": {
                "s2_ids": ["latitude", "longitude"],
                "days_of_week": ["day_of_week"],
            },
        },
    )
    for exp in xp_client.list_experiments(p["project_id"]):
        if exp["status"] == "active":
            xp_client.disable_experiment(p["project_id"], exp["id"])

    return p


def generate_experiment_spec(
    days_of_week,
    type_="A/B",
    interval=None,
    treatments=None,
    s2id_level=14,
    tier="default",
):
    s2id = s2cell.lat_lon_to_cell_id(ID_S2_POINT[0], ID_S2_POINT[1], s2id_level)
    spec = copy.deepcopy(BASE_EXPERIMENT)
    spec["start_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(seconds=2)
    ).isoformat() + "Z"
    spec["end_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(days=1)
    ).isoformat() + "Z"
    spec["segment"]["s2_ids"] = [s2id]
    spec["segment"]["days_of_week"] = days_of_week
    spec["type"] = type_
    spec["tier"] = tier
    spec["interval"] = interval
    spec["treatments"] = treatments or spec["treatments"]
    return spec


def test_experiment_selection(xp_client: XPClient, xp_project):
    """Test experiment selection for various types of experiment in the project.
    Note: S2ID levels used in the experiments should be within the limits used
    by the Management and Treatment service configs.
    """
    # Most granular S2ID, unmatched service type
    exp_spec_1 = generate_experiment_spec([1])
    # Most granular S2ID, optional service type
    exp_spec_2 = generate_experiment_spec([])
    # Least granular S2ID, override tier
    exp_spec_3 = generate_experiment_spec([1, 2], s2id_level=10, tier="override")
    # More granular S2ID, default tier
    exp_spec_4 = generate_experiment_spec([2], s2id_level=12)  # More granular S2ID
    # More granular S2ID, override tier
    exp_spec_5 = generate_experiment_spec([2], s2id_level=12, tier="override")

    _ = xp_client.create_experiment(xp_project["project_id"], exp_spec_1)
    _ = xp_client.create_experiment(xp_project["project_id"], exp_spec_2)
    _ = xp_client.create_experiment(xp_project["project_id"], exp_spec_3)
    _ = xp_client.create_experiment(xp_project["project_id"], exp_spec_4)
    exp_5 = xp_client.create_experiment(xp_project["project_id"], exp_spec_5)

    wait_for(datetime.datetime.fromisoformat(exp_5["start_time"][:23]))

    treatment = eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": ID_S2_POINT[0],
                "longitude": ID_S2_POINT[1],
                "day_of_week": 2,
                "order_id": 1,
            },
            pass_key=xp_project["passkey"],
        )
    )

    assert treatment == {
        "experiment_id": exp_5["id"],
        "experiment_name": exp_5["name"],
        "treatment": exp_5["treatments"][0],
        "metadata": {
            "experiment_version": 1,
            "experiment_type": "A/B",
        },
    }


def test_treatment_distribution_for_ab_experiment(xp_client: XPClient, xp_project):
    exp_spec = generate_experiment_spec(
        [1],
        treatments=[
            {"name": "Treatment-1", "traffic": 30, "configuration": {}},
            {"name": "Treatment-2", "traffic": 70, "configuration": {}},
        ],
    )

    experiment = xp_client.create_experiment(xp_project["project_id"], exp_spec)
    wait_for(parse_datetime(experiment["start_time"]))

    total_requests = 2000
    counts = {"Treatment-1": 0, "Treatment-2": 0}

    for rand_key in range(total_requests):
        treatment = xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": ID_S2_POINT[0],
                "longitude": ID_S2_POINT[1],
                "day_of_week": 1,
                "order_id": rand_key,
            },
            pass_key=xp_project["passkey"],
        )

        counts[treatment["treatment"]["name"]] += 1

    epsilon = 0.03
    assert abs(counts["Treatment-1"] - 0.3 * total_requests) < epsilon * total_requests
    assert abs(counts["Treatment-2"] - 0.7 * total_requests) < epsilon * total_requests


def test_treatment_distribution_for_switchback_experiment(
    xp_client: XPClient, xp_project
):
    exp_spec = generate_experiment_spec(
        [1],
        type_="Switchback",
        interval=1,
        treatments=[
            {"name": "Treatment-1", "configuration": {}},
            {"name": "Treatment-2", "configuration": {}},
        ],
    )

    experiment = xp_client.create_experiment(xp_project["project_id"], exp_spec)
    wait_for(parse_datetime(experiment["start_time"]))

    for rand_key in range(100):
        treatment = xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": ID_S2_POINT[0],
                "longitude": ID_S2_POINT[1],
                "day_of_week": 1,
                "order_id": rand_key,
            },
            pass_key=xp_project["passkey"],
        )

        assert treatment["treatment"]["name"] == "Treatment-1"

    wait_for(parse_datetime(experiment["start_time"]) + datetime.timedelta(minutes=1))

    for rand_key in range(100):
        treatment = xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": ID_S2_POINT[0],
                "longitude": ID_S2_POINT[1],
                "day_of_week": 1,
                "order_id": rand_key,
            },
            pass_key=xp_project["passkey"],
        )

        assert treatment["treatment"]["name"] == "Treatment-2"
