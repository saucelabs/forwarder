#!/usr/bin/env python

import os
import time
import signal
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer

# Headers taken from `failed.pcapng`
headers = [
    ("Content-Encoding", "gzip"),
    (
        "Set-Cookie",
        'glide_user=""; Expires=Thu, 01-Jan-1970 00:00:10 GMT; Path=/; HttpOnly',
    ),
    (
        "Set-Cookie",
        'glide_user_session=""; Expires=Thu, 01-Jan-1970 00:00:10 GMT; Path=/; HttpOnly',
    ),
    ("X-Is-Logged-In", "false"),
    ("X-Transaction-ID", "7cba67127310"),
    ("Pragma", "no-store,no-cache"),
    ("Cache-control", "no-cache,no-store,must-revalidate,max-age=-1"),
    ("Expires", "0"),
    ("Content-Type", "application/json;charset=UTF-8"),
    ("Transfer-Encoding", "chunked"),
    ("Date", "Thu, 23 Apr 2020 10:17:02 GMT"),
    ("Server", "ServiceNow"),
]

stats_headers = [
    ("Set-Cookie", "JSESSIONID=014F24618731E4C4CED40B8288E4CE62; Path=/; HttpOnly"),
    ("Pragma", "no-store,no-cache"),
    ("Cache-control", "no-cache,no-store,must-revalidate,max-age=-1"),
    ("Expires", "0"),
    ("Content-Type", "text/html"),
    ("Transfer-Encoding", "chunked"),
    ("Date", "Thu, 23 Apr 2020 10:16:28 GMT"),
    ("Server", "ServiceNow"),
]


class RequestHandler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"
    server_version = "ServiceNow"
    sys_version = ""

    # Chunked encoding of:
    #   '{"android":{"min_version":"4.0.0"},"ios":{"min_version":"4.0.0"}}'
    data = bytearray.fromhex(
        "61 0d 0a 1f 8b 08 00 00 00 00 00 00 00 0d \
        0a 33 34 0d 0a ab 56 4a cc 4b 29 ca cf 4c \
        51 b2 aa 56 ca cd cc 8b 2f 4b 2d 2a ce cc \
        cf 53 b2 52 32 d1 33 d0 33 50 aa d5 51 ca \
        cc 2f c6 29 5b 0b 00 37 57 ea 1d 41 00 00 \
        00 0d 0a 30 0d 0a 0d 0a"
    )

    def version_string(self):
        return self.server_version

    def do_HEAD(self):
        print("In HEAD")
        self.send_response(200)

        for header in stats_headers:
            self.send_header(header[0], header[1])
        self.end_headers()

    def do_GET(self):

        print("Start request")

        print(self.requestline)
        print(self.command)
        print(self.path)
        print(self.headers)

        bad_start = False
        # bad_chunk = False
        # bad_mid = False

        if bad_start:
            self.data[0] = 0

        # if bad_chunk:
        #     self.data[15] = 0

        # if bad_mid:
        #     self.data[16] = 0

        self.send_response_only(200)

        for header in headers:
            self.send_header(header[0], header[1])

        self._headers_buffer.append(b"\r\n")
        self._headers_buffer.append(self.data[0:15])

        # Write the headers + the first chunk of data to get a 520 byte TCP
        # packet
        self.wfile.write(b"".join(self._headers_buffer))

        self._headers_buffer = []

        # self.wfile.flush()
        # time.sleep(0.1)

        # Write the remainder
        self.wfile.write(self.data[15:])
        print("Done")

def signal_handler(signum, frame):
    print("Received signal to terminate. Exiting gracefully...")
    sys.exit(0)

if __name__ == "__main__":
    port = 8307
    print("Listening on 0.0.0.0:%s" % port)
    server = HTTPServer(("", port), RequestHandler)

    signal.signal(signal.SIGTERM, signal_handler)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        print("Shutting down server.")
        server.server_close()
