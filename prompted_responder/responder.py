#!/usr/bin/env python3


import argparse
import logging
import requests
import signal
import sys
import time

from llama_cpp import Llama

def signal_handler(sig, frame):
    """Exit program successfully"""
    logging.info("exiting")
    sys.exit(0)



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

    signal.signal(signal.SIGINT, signal_handler)

    while True:

        time.sleep(1)

        # Fetch status and determine whether we should ignore the response

        resp = requests.get(URLS["status"]).json()

        if "error" in resp:
            logging.warn("There was an error:", resp["error"])
            continue

        if resp.get("response"):
            logging.warn("Response was non-empty, so no need to compute a response")
            continue

        if not resp.get("promopt"):
            logging.warn("prompt was empty skipping")

        prompt = resp["prompt"].strip()

        if not prompt.startswith("Q:"):
            response = "prompt must start with Q:"

        elif not prompt.endswith("A:"):
            response = "prompt must end with A:"

        else:
            try:
                output = llm(
                    resp["prompt"],
                    max_tokens=32,
                    stop=["Q:", "\n"],
                    echo=True,
                )
                response = output["choices"][0]["text"]
            except Excaption as e:
                logging.error("Received error!", e)
                response = "some other error happened"

        # Compute response for the prompt and send it up

        d = {
            "secret": args.secret,
            "prompt": resp["prompt"],
            "response": response,
        }

        res = requests.post(URLS["respond"], data=d)
