import datetime
import time
import uuid


def wait_for(ts: datetime.datetime):
    while datetime.datetime.utcnow() < ts:
        time.sleep(0.1)


def eventually(fn, interval=0.1, timeout=5):
    start = time.time()
    while True:
        try:
            return fn()
        except:  # noqa
            if time.time() - start > timeout:
                raise

            time.sleep(interval)


def parse_datetime(s):
    # since datetime format returned by management service is unstable
    # (microseconds can have variable length) we take only stable part from datetime string
    return datetime.datetime.fromisoformat(s[: len("YYYY-mm-ddTHH:MM:SS.fff")])


def generate_segment(segment_config):
    # require uuid since segment name is unique
    segment = {
        "name": str(uuid.uuid4()),
        "segment": segment_config,
    }

    return segment


def generate_treatment():
    # require uuid since treatment name is unique
    treatment = {
        "name": str(uuid.uuid4()),
        "configuration": {"foo": "bar"},
    }

    return treatment
