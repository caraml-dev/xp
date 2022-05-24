import copy
import datetime
from functools import partial
from typing import Tuple

import pytest
import s2cell
from data import (
    BASE_EXPERIMENT,
    DAY_OF_WEEK_MONDAY,
    DAY_OF_WEEK_TUESDAY,
    ID_S2_POINT,
    SG_S2_POINT,
    SG_S2_POINT2,
)
from utils import (
    eventually,
    generate_segment,
    generate_treatment,
    parse_datetime,
    wait_for,
)
from xp_client import NotFound, XPClient


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
    s2_point: Tuple, day_of_week: int, start_after_seconds: int = 2
):
    s2id = s2cell.lat_lon_to_cell_id(s2_point[0], s2_point[1], 14)
    experiment_spec = copy.deepcopy(BASE_EXPERIMENT)
    # start_time needs to be in the future
    experiment_spec["start_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(seconds=start_after_seconds)
    ).isoformat() + "Z"
    experiment_spec["end_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(days=1)
    ).isoformat() + "Z"
    experiment_spec["segment"]["s2_ids"] = [s2id]
    experiment_spec["segment"]["days_of_week"] = [day_of_week]
    return experiment_spec


def test_simple_experiment_creation(xp_client: XPClient, xp_project):
    experiment_spec = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    experiment = xp_client.create_experiment(xp_project["project_id"], experiment_spec)

    # we need to wait until experiment became active
    wait_for(parse_datetime(experiment["start_time"]))

    treatment = xp_client.fetch_treatment(
        xp_project["project_id"],
        {
            "latitude": SG_S2_POINT[0],
            "longitude": SG_S2_POINT[1],
            "order_id": 1,
            "day_of_week": DAY_OF_WEEK_MONDAY,
        },
        pass_key=xp_project["passkey"],
    )

    assert treatment == {
        "experiment_id": experiment["id"],
        "experiment_name": experiment["name"],
        "treatment": experiment["treatments"][0],
    }


def test_experiment_disabling(xp_client: XPClient, xp_project):
    experiment_spec = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    experiment = xp_client.create_experiment(xp_project["project_id"], experiment_spec)
    assert experiment["status"] == "active"

    # wait for propagation
    eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "order_id": 1,
                "day_of_week": DAY_OF_WEEK_MONDAY,
            },
            pass_key=xp_project["passkey"],
        )
    )

    xp_client.disable_experiment(xp_project["project_id"], experiment["id"])
    history = xp_client.list_experiment_history(
        xp_project["project_id"], experiment["id"]
    )
    assert history[0]["status"] == "active"

    eventually(
        partial(
            pytest.raises,
            NotFound,
            lambda: xp_client.fetch_treatment(
                xp_project["project_id"],
                {
                    "latitude": SG_S2_POINT[0],
                    "longitude": SG_S2_POINT[1],
                    "order_id": 1,
                    "day_of_week": DAY_OF_WEEK_MONDAY,
                },
                pass_key=xp_project["passkey"],
            ),
        )
    ), "Experiment no longer available in treatment service"

    xp_client.enable_experiment(xp_project["project_id"], experiment["id"])
    history = xp_client.list_experiment_history(
        xp_project["project_id"], experiment["id"]
    )
    assert history[0]["status"] == "inactive"

    eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "order_id": 1,
                "day_of_week": DAY_OF_WEEK_MONDAY,
            },
            pass_key=xp_project["passkey"],
        )
    )


def test_segment_updating(xp_client: XPClient, xp_project):
    segment_spec = generate_segment(segment_config={"days_of_week": [1]})
    configured_segment = xp_client.create_segment(
        xp_project["project_id"], segment_spec
    )

    new_segment = {
        "segment": {"days_of_week": [2]},
    }
    updated_segment = xp_client.update_segment(
        project_id=xp_project["project_id"],
        segment_id=configured_segment["id"],
        segment=new_segment,
    )

    segment = xp_client.get_segment(
        project_id=xp_project["project_id"], segment_id=configured_segment["id"]
    )

    assert updated_segment == segment

    history = xp_client.list_segment_history(xp_project["project_id"], segment["id"])
    assert history[0]["segment"] == {"days_of_week": [1]}


def test_treatment_updating(xp_client: XPClient, xp_project):
    treatment_spec = generate_treatment()
    configured_treatment = xp_client.create_treatment(
        xp_project["project_id"], treatment_spec
    )

    new_treatment = {
        "status": "active",
        "configuration": {"bar": "baz"},
    }
    updated_treatment = xp_client.update_treatment(
        project_id=xp_project["project_id"],
        treatment_id=configured_treatment["id"],
        treatment=new_treatment,
    )

    treatment = xp_client.get_treatment(
        project_id=xp_project["project_id"], treatment_id=configured_treatment["id"]
    )

    assert updated_treatment == treatment

    history = xp_client.list_treatment_history(
        xp_project["project_id"], treatment["id"]
    )
    assert history[0]["configuration"] == {"foo": "bar"}


def test_project_updating(xp_client: XPClient, xp_project):
    experiment_spec = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec)

    xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="customer_id",
        segmenters={
            "names": ["s2_ids", "days_of_week"],
            "variables": {
                "s2_ids": ["latitude", "longitude"],
                "days_of_week": ["day_of_week"],
            },
        },
    )

    eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "customer_id": 1,
                "day_of_week": DAY_OF_WEEK_MONDAY,
            },
            pass_key=xp_project["passkey"],
        )
    )


def test_project_remove_segmenter_invalid_updating(xp_client: XPClient, xp_project):
    experiment_spec1 = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    experiment_spec2 = generate_experiment_spec(SG_S2_POINT2, DAY_OF_WEEK_MONDAY)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec1)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec2)

    resp = xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="order_id",
        segmenters={
            "names": ["days_of_week"],
            "variables": {"days_of_week": ["day_of_week"]},
        },
    )

    assert resp["code"] == "400"
    assert "Segment Orthogonality check failed" in resp["error"]


def test_project_remove_segmenter_valid_updating(xp_client: XPClient, xp_project):
    new_segmenters = {
        "names": ["days_of_week"],
        "variables": {"days_of_week": ["day_of_week"]},
    }

    experiment_spec1 = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    experiment_spec2 = generate_experiment_spec(ID_S2_POINT, DAY_OF_WEEK_TUESDAY)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec1)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec2)

    resp = xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="order_id",
        segmenters=new_segmenters,
    )

    assert resp["segmenters"] == new_segmenters


def test_project_add_segmenter_updating(xp_client: XPClient, xp_project):
    new_segmenters = {
        "names": ["s2_ids", "days_of_week", "hours_of_day"],
        "variables": {
            "s2_ids": ["latitude", "longitude"],
            "days_of_week": ["tz"],
            "hours_of_day": ["tz"],
        },
    }

    experiment_spec = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY)
    xp_client.create_experiment(xp_project["project_id"], experiment_spec)

    resp = xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="order_id",
        segmenters=new_segmenters,
    )

    assert resp["segmenters"] == new_segmenters


def test_experiment_updating(xp_client: XPClient, xp_project):
    experiment_spec = generate_experiment_spec(SG_S2_POINT, DAY_OF_WEEK_MONDAY, 5)
    experiment = xp_client.create_experiment(xp_project["project_id"], experiment_spec)

    # wait for propagation
    eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "order_id": 1,
                "day_of_week": DAY_OF_WEEK_MONDAY,
            },
            pass_key=xp_project["passkey"],
        )
    )

    experiment["treatments"][0]["name"] = "Treatment-2"

    experiment = xp_client.update_experiment(
        xp_project["project_id"], experiment["id"], experiment
    )
    assert experiment["treatments"][0]["name"] == "Treatment-2"

    history = xp_client.list_experiment_history(
        xp_project["project_id"], experiment["id"]
    )
    assert history[0]["treatments"][0]["name"] == "Treatment-1"

    def check():
        treatment = xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "order_id": 1,
                "day_of_week": DAY_OF_WEEK_MONDAY,
            },
            pass_key=xp_project["passkey"],
        )

        assert treatment == {
            "experiment_id": experiment["id"],
            "experiment_name": experiment["name"],
            "treatment": experiment["treatments"][0],
        }

    eventually(check)


def test_all_segmenters(xp_client: XPClient, xp_project):
    segmenters = {
        "names": ["s2_ids", "days_of_week", "hours_of_day"],
        "variables": {
            "s2_ids": ["latitude", "longitude"],
            "days_of_week": ["tz"],
            "hours_of_day": ["tz"],
        },
    }
    s2id = s2cell.lat_lon_to_cell_id(SG_S2_POINT[0], SG_S2_POINT[1], 14)
    project = xp_client.create_or_update_project(
        user_name="test-project",
        randomization_key="order_id",
        segmenters=segmenters,
    )

    exp_spec = copy.deepcopy(BASE_EXPERIMENT)
    exp_spec["start_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(seconds=1)
    ).isoformat() + "Z"
    exp_spec["end_time"] = (
        datetime.datetime.utcnow() + datetime.timedelta(days=1)
    ).isoformat() + "Z"
    exp_spec["segment"]["s2_ids"] = [s2id]
    exp_spec["segment"]["days_of_week"] = list(range(1, 8))
    exp_spec["segment"]["hours_of_day"] = list(range(0, 24))

    experiment = xp_client.create_experiment(project["project_id"], exp_spec)
    assert experiment["segment"] == exp_spec["segment"]

    treatment = eventually(
        lambda: xp_client.fetch_treatment(
            xp_project["project_id"],
            {
                "latitude": SG_S2_POINT[0],
                "longitude": SG_S2_POINT[1],
                "order_id": 1,
                "tz": "Asia/Singapore",
            },
            pass_key=xp_project["passkey"],
        )
    )

    assert treatment == {
        "experiment_id": experiment["id"],
        "experiment_name": experiment["name"],
        "treatment": experiment["treatments"][0],
    }


def test_get_segmenters(xp_client: XPClient, xp_project):
    segmenters = xp_client.get_segmenters(xp_project["project_id"])
    assert segmenters[0]["name"] == "s2_ids"
    assert segmenters[0]["treatment_request_fields"] == [
        ["s2id"],
        ["latitude", "longitude"],
    ]
    assert segmenters[0]["type"] == "INTEGER"
