#!/usr/bin/env python3


import argparse
import logging
import requests
import signal
import sys
import time


def signal_handler(sig, frame):
    sys.exit(0)



if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument("--host", default="whatever.rip")
    parser.add_argument("--secret", default="8181")
    args = parser.parse_args()

    URLS = {}
    URLS["status"] = f"http://{args.host}/status"
    URLS["respond"] = f"http://{args.host}/respond"

    signal.signal(signal.SIGINT, signal_handler)

    while True:

        time.sleep(1)

        resp = requests.get(URLS["status"]).json()

        if "error" in resp:
            logging.warn("There was an error:", resp["error"])
            continue

        if resp.get("response"):
            logging.warn("Response was non-empty, so no need to compute a response")
            continue

        d = {
            "secret": args.secret,
            "prompt": resp["prompt"],
            "response": "weom: " + resp["prompt"],
        }

        res = requests.post(URLS["respond"], data=d)
