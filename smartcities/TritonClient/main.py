#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

from http.server import HTTPServer, BaseHTTPRequestHandler
import random
import string
import subprocess
import sys
import io

host = ('', 8302)
tmpDir = "/tmp/"
image_client = "/home/TritonClient/image_client.py"
get_router = {
        "/version": "get_version", # curl -v 127.0.0.1:8302/version
        }
post_router = {
        "/image": "post_image", # curl -v -X POST -T ./cup.png 127.0.0.1:8302/image
        }

class Resquest(BaseHTTPRequestHandler):
    def do_GET(self):
        func = get_router.get(self.path, None)
        if func:
            getattr(self, func)()
        else:
            self.send_error(404)

    def do_POST(self):
        func = post_router.get(self.path, None)
        if func:
            getattr(self, func)()
        else:
            self.send_error(404)

    def get_version(self):
        data = "1.0\n"
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(data.encode())

    def post_image(self):
        data = self.rfile.read(int(self.headers['Content-Length']))
        if len(data) > 0:
            # save to tmp file
            name = ''.join(random.sample(string.ascii_letters + string.digits, 16))
            path = tmpDir + name
            with open(path, 'wb') as f:
                f.write(data)
                # call image_client.py to get info
                mystdout = io.StringIO()
                _argv = sys.argv
                _stdout = sys.stdout
                try:
                    sys.argv = [image_client, '-m', 'densenet_onnx', '-c', '1', '-s', 'INCEPTION', '-u', 'host.docker.internal:8000', path]
                    sys.stdout = mystdout
                    with open(image_client) as fp:
                        exec(fp.read(), globals())
                finally:
                    sys.argv = _argv
                    sys.stdout = _stdout
                rtn = mystdout.getvalue()
                # send back request
                self.send_response(200)
                self.send_header("Content-type", "text/plain")
                self.end_headers()
                self.wfile.write(rtn.encode())
                return
            self.send_error(500)
        else:
            self.send_error(501)

if __name__ == "__main__":
    server = HTTPServer(host, Resquest)
    print("Starting server, %s:%s" % host)
    server.serve_forever()
