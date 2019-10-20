#!/usr/bin/env python3.7

import asyncio
import logging
import os
import time
from multiprocessing import Process, Value

from flask import Flask, request


import pykeybasebot.types.chat1 as chat1
from pykeybasebot import Bot

logging.basicConfig(level=logging.DEBUG)


class Handler:
    """
    A Keybase chatbot handler that reacts to two-man requests with a thumbs up
    """
    def __init__(self, shared_running_val: Value):
        self.shared_running_val = shared_running_val

    async def __call__(self, bot, event):
        if self.shared_running_val.value:
            if event.msg.content.type_name != chat1.MessageTypeStrings.TEXT.value:
                return
            channel = event.msg.channel
            msg_id = event.msg.id
            body = event.msg.content.text.body
            if "has requested access to the two-man realm" in body:
                await bot.chat.react(channel, msg_id, ":+1:")

# A shared boolean flag that tracks whether the auto-reacter is currently running
shared_running_val = Value('i', 0)

def start_bot_event_loop():
    # Start the bot running in a separate process so that it doesn't block the main process that hosts the flask
    # webserver
    username = os.environ["TWO_MAN_APPROVER_USERNAME"]
    paperkey = os.environ["TWO_MAN_APPROVER_PAPERKEY"]
    bot = Bot(
        username=username, paperkey=paperkey,
        handler=Handler(shared_running_val)
    )
    p = Process(target=lambda: asyncio.run(bot.start({})))
    p.start()

app = Flask(__name__)

@app.route('/start')
def start_autoresponder():
    global shared_running_val
    shared_running_val.value = 1
    time.sleep(1)
    return "OK"

@app.route('/stop')
def stop_autoresponder():
    global shared_running_val
    shared_running_val.value = 0
    time.sleep(1)
    return "OK"


if __name__ == '__main__':
    start_bot_event_loop()
    app.run(host='0.0.0.0', port='8080')
