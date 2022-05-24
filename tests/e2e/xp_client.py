import json
import typing
from functools import partial

import requests


class XPClientError(Exception):
    pass


class NotFound(XPClientError):
    pass


class ServerError(XPClientError):
    pass


TEST_PROJECT_ID = 999


class XPClient:
    def __init__(
        self,
        management_url="http://localhost:3000/v1",
        treatment_url="http://localhost:8080/v1",
    ):
        self._management_url = management_url.rstrip("/")
        self._treatment_url = treatment_url.rstrip("/")

    def create_or_update_project(
        self, user_name: str, randomization_key: str, segmenters: typing.List[str]
    ):
        existing_projects = requests.get(f"{self._management_url}/projects").json()[
            "data"
        ]
        existing_projects = [p for p in existing_projects if p["username"] == user_name]

        if existing_projects:
            method = partial(
                requests.put,
                f"{self._management_url}/projects/{existing_projects[0]['id']}/settings",
            )
        else:
            method = partial(
                requests.post,
                f"{self._management_url}/projects/{TEST_PROJECT_ID}/settings",
            )

        resp = method(
            data=json.dumps(
                dict(
                    username=user_name,
                    randomization_key=randomization_key,
                    segmenters=segmenters,
                )
            ),
            headers={"Content-Type": "application/json"},
        )
        if resp.status_code != 200:
            return resp.json()
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def create_experiment(
        self, project_id: int, experiment: typing.Dict[str, typing.Any]
    ):
        resp = requests.post(
            f"{self._management_url}/projects/{project_id}/experiments",
            data=json.dumps(experiment),
            headers={"Content-Type": "application/json"},
        )
        try:
            assert resp.status_code == 200, resp.content
        except AssertionError:
            if (
                resp.status_code != 400
                or resp.json()["message"] != "Segment Orthogonality check failed"
            ):
                raise

            # experiment already exist -> updating
            experiments = requests.get(
                f"{self._management_url}/projects/{project_id}/experiments"
            ).json()["data"]
            experiment_id = [
                e["id"]
                for e in experiments
                if e["name"] == experiment["name"] and e["status"] == "active"
            ][0]
            return self.update_experiment(project_id, experiment_id, experiment)

        return resp.json()["data"]

    def update_experiment(
        self, project_id, experiment_id, experiment: typing.Dict[str, typing.Any]
    ):
        resp = requests.put(
            f"{self._management_url}/projects/{project_id}/experiments/{experiment_id}",
            data=json.dumps(experiment),
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def list_experiments(self, project_id):
        return (
            requests.get(
                f"{self._management_url}/projects/{project_id}/experiments"
            ).json()["data"]
            or []
        )

    def list_experiment_history(self, project_id, experiment_id):
        return (
            requests.get(
                f"{self._management_url}/projects/{project_id}/experiments/{experiment_id}/history"
            ).json()["data"]
            or []
        )

    def disable_experiment(self, project_id, experiment_id):
        resp = requests.put(
            f"{self._management_url}/projects/{project_id}/experiments/{experiment_id}/disable"
        )
        assert resp.status_code == 204, resp.content

    def enable_experiment(self, project_id, experiment_id):
        resp = requests.put(
            f"{self._management_url}/projects/{project_id}/experiments/{experiment_id}/enable"
        )
        assert resp.status_code == 204, resp.content

    def create_segment(self, project_id: int, segment: typing.Dict[str, typing.Any]):
        resp = requests.post(
            f"{self._management_url}/projects/{project_id}/segments",
            data=json.dumps(segment),
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def update_segment(
        self, project_id, segment_id, segment: typing.Dict[str, typing.Any]
    ):
        resp = requests.put(
            f"{self._management_url}/projects/{project_id}/segments/{segment_id}",
            data=json.dumps(segment),
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def get_segment(self, project_id, segment_id):
        resp = requests.get(
            f"{self._management_url}/projects/{project_id}/segments/{segment_id}",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def create_treatment(
        self, project_id: int, treatment: typing.Dict[str, typing.Any]
    ):
        resp = requests.post(
            f"{self._management_url}/projects/{project_id}/treatments",
            data=json.dumps(treatment),
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def update_treatment(
        self, project_id, treatment_id, treatment: typing.Dict[str, typing.Any]
    ):
        resp = requests.put(
            f"{self._management_url}/projects/{project_id}/treatments/{treatment_id}",
            data=json.dumps(treatment),
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def get_treatment(self, project_id, treatment_id):
        resp = requests.get(
            f"{self._management_url}/projects/{project_id}/treatments/{treatment_id}",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200, resp.content
        return resp.json()["data"]

    def list_treatment_history(self, project_id, treatment_id):
        return (
            requests.get(
                f"{self._management_url}/projects/{project_id}/treatments/{treatment_id}/history"
            ).json()["data"]
            or []
        )

    def get_segmenters(self, project_id: int):
        return (
            requests.get(
                f"{self._management_url}/projects/{project_id}/segmenters"
            ).json()["data"]
            or []
        )

    def list_segment_history(self, project_id, segment_id):
        return (
            requests.get(
                f"{self._management_url}/projects/{project_id}/segments/{segment_id}/history"
            ).json()["data"]
            or []
        )

    def fetch_treatment(
        self, project_id: int, req: typing.Dict[str, typing.Any], pass_key: str
    ):
        resp = requests.post(
            f"{self._treatment_url}/projects/{project_id}/fetch-treatment",
            data=json.dumps(req),
            headers={"Content-Type": "application/json", "pass-key": pass_key},
        )
        if resp.status_code == 500:
            raise ServerError(resp.content)

        assert resp.status_code == 200, resp.content
        if not resp.json().get("data"):
            raise NotFound

        return resp.json()["data"]

    def validate_treatment(self, req: typing.Dict[str, typing.Any]):
        return requests.post(
            f"{self._management_url}/validate",
            data=json.dumps(req),
            headers={"Content-Type": "application/json"},
        )
