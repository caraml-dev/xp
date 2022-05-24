from xp_client import XPClient


def test_validate_valid_treatment_with_url(xp_client: XPClient):
    data = {
        "validation_url": "http://www.mocky.io/v2/5e4cacd4310000e1cad8c073",
        "data": {"field1": "abc", "field2": "def", "field3": {"field4": 0.1}},
    }

    result = xp_client.validate_treatment(data)

    assert result.status_code == 200
    assert result.text == ""


def test_validate_valid_treatment_with_schema(xp_client: XPClient):
    data = {
        "treatment_schema": {
            "rules": [
                {"name": "test-rule-1", "predicate": '{{- (eq .field1 "abc") -}}'}
            ]
        },
        "data": {"field1": "abc", "field2": "def", "field3": {"field4": 0.1}},
    }

    result = xp_client.validate_treatment(data)

    assert result.status_code == 200
    assert result.text == ""


def test_validate_invalid_treatment_with_schema(xp_client: XPClient):
    data = {
        "treatment_schema": {
            "rules": [
                {"name": "test-rule-1", "predicate": '{{- (eq .field1 "def") -}}'}
            ]
        },
        "data": {"field1": "abc", "field2": "def", "field3": {"field4": 0.1}},
    }

    result = xp_client.validate_treatment(data)

    assert result.status_code == 500
    assert result.json()["error"] == "Go template rule test-rule-1 returns false"
    assert result.json()["message"] == "Go template rule test-rule-1 returns false"
