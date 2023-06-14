import http.server
import json
import socket
import threading
from contextlib import closing
from http import HTTPStatus

import pytest


def find_free_port():
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as s:
        s.bind(("", 0))
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        return s.getsockname()[1]


class MLPService(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        print("Request", self.path)
        if self.path == "/v1/projects":
            self.get_projects()
            return

        self.send_error(HTTPStatus.NOT_FOUND, "URL not found")

    def get_projects(self):
        body = json.dumps([{"id": 999, "name": "test-project"}]).encode()

        self.send_response(HTTPStatus.OK)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()

        self.wfile.write(body)


@pytest.fixture(scope="session")
def mlp_service():
    port = find_free_port()
    with http.server.HTTPServer(("", port), MLPService) as httpd:
        threading.Thread(target=httpd.serve_forever, daemon=True).start()
        yield f"http://localhost:{port}"
        httpd.shutdown()
