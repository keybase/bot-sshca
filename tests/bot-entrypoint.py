#!/bin/python3

import os
import shlex
import time

from flask import Flask, request

app = Flask(__name__)

@app.route('/load_env')
def load_env():
    filename = request.args.get('filename')
    path = os.path.join("tests/generated-env/", filename)
    os.system((
        "killall keybaseca 2>&1 > /dev/null|| true\n"
        ". %s\n"
        "bin/keybaseca --wipe-all-configs\n"
        "bin/keybaseca --wipe-logs || true\n"
        "bin/keybaseca generate --overwrite-existing-key\n"
        "echo yes | bin/keybaseca backup > /mnt/cakey.backup\n"
        "bin/keybaseca service &"
    ) % (shlex.quote(path)))
    time.sleep(2)
    return "OK"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port='8080')