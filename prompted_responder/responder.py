#!/usr/bin/env python3


import argparse
import logging
import requests
import signal
import sys
import time


try:
    from llama_cpp import Llama
    DEBUG = False
except ImportError:
    from unittest import mock
    logging.warn("Could not import Llama from llama_cpp")
    DEBUG = True
    Llama = mock.MagicMock()


def signal_handler(sig, frame):
    """Exit program successfully"""
    logging.info("exiting")
    sys.exit(0)


def predict(prompt):
    """Return a LLaMa response given a prompt"""

    if DEBUG:
        time.sleep(2)
        return f"[DEBUG] Q: {prompt} A: This is a response"

    if not prompt.startswith("Q:"):
        return "prompt must start with Q:"

    elif not prompt.endswith("A:"):
        return "prompt must end with A:"

    try:
        output = llm(
            resp["prompt"],
            max_tokens=32,
            stop=["Q:", "\n"],
            echo=True,
        )
        response = output["choices"][0]["text"]
    except Exception as e:
        logging.error("Received error!", e)
        response = "some other error happened"

    return response


def heartbeat(url, secret, state):
    """Send a heartbeat to the server"""

    d = {
        "secret": secret,
        "state": state,
    }

    return requests.post(url, data=d)


def respond(url, secret, prompt, response):
    """Send a response to the server"""

    d = {
        "secret": secret,
        "prompt": prompt,
        "response": response,
        "state": "ready",
    }

    return requests.post(url, data=d)


if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument("--host", default="whatever.rip")
    parser.add_argument("--secret", default="8181")
    args = parser.parse_args()

    llm = Llama(
        model_path="/home/matt/Models/Llama-2-7b-chat-hf/ggml-model-q4_0.gguf",
        n_gpu_layers=32,
    )

    URLS = {}
    URLS["status"] = f"http://{args.host}/status"
    URLS["respond"] = f"http://{args.host}/respond"
    URLS["heartbeat"] = f"http://{args.host}/heartbeat"

    signal.signal(signal.SIGINT, signal_handler)

    while True:

        time.sleep(1)

        # Fetch status and determine whether we should ignore the response

        resp = requests.get(URLS["status"]).json()

        if resp.get("state") != "waiting":
            logging.info("server is not waiting, so skipping")
            continue

        if "error" in resp:
            logging.warn("There was an error:", resp["error"])
            continue

        if resp.get("response"):
            logging.warn("Response was non-empty, so no need to compute a response")
            continue

        if not resp.get("prompt"):
            logging.warn("prompt was empty skipping")
            continue

        heartbeat(
            URLS["heartbeat"],
            args.secret,
            "working",
        )

        response = predict(resp["prompt"].strip())

        heartbeat(
            URLS["heartbeat"],
            args.secret,
            "ready",
        )

        respond(
            URLS["respond"],
            args.secret,
            resp["prompt"],
            response,
        ).json()
