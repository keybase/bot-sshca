#!/bin/python3

"""
This file is the main process running inside of the ca-bot container for the integration tests. This allows the kssh
container to specify that the tests should run with a specific set of environment variables. This allows us to easily
run integration tests for different keybaseca config options.
"""

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
        # The output from this backup is tested in test_env_1.py
        "echo yes | bin/keybaseca backup > /shared/cakey.backup\n"
        # The output from this sign operation is tested in test_env_1.py
        "ssh-keygen -t ed25519 -f /shared/userkey -N '' && bin/keybaseca sign --public-key /shared/userkey.pub > /shared/keybaseca-sign.out\n"
        "bin/keybaseca service &"
    ) % (shlex.quote(path)))
    # Sleep so keybaseca has time to start
    time.sleep(5)
    return "OK"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port='8080')
