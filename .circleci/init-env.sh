#!/bin/bash

# Create a valid tests/env.sh file from values currently in the environment. This is kind of ugly, but we need an actual
# env.sh file because the test code which runs inside of docker needs to be able to source it.

echo "export BOT_USERNAME='$BOT_USERNAME'"
echo "export BOT_PAPERKEY='$BOT_PAPERKEY'"
echo "export KSSH_USERNAME='$KSSH_USERNAME'"
echo "export KSSH_PAPERKEY='$KSSH_PAPERKEY'"
echo "export SUBTEAM='$SUBTEAM'"
echo "export SUBTEAM_SECONDARY='$SUBTEAM_SECONDARY'"
