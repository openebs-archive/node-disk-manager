#!/bin/bash

export GOTRACEBACK=crash

# openebs base directory inside the container
OPENEBS_BASE_DIR="/var/openebs"
# ndm base directory inside the container. It will be the ndm
# directory inside openebs base directory
NDM_BASE_DIR="${OPENEBS_BASE_DIR}/ndm"

# set ulimit to 0, if the core dump is not enabled
if [ -z "$ENABLE_COREDUMP" ]; then
  ulimit -c 0
else
  # making sure mountpath inside the container is available
  if ! [ -d "${NDM_BASE_DIR}" ]; then
    echo "OpenEBS/NDM Base directory not found"
    exit 1
  fi
  # set ulimit to unlimited and create a core directory for creating coredump
  echo "[entrypoint.sh] enabling core dump."
  ulimit -c unlimited
  echo "[entrypoint.sh] creating ${NDM_BASE_DIR}/core if not exists."
  mkdir -p "${NDM_BASE_DIR}/core"
  echo "[entrypoint.sh] changing directory to ${NDM_BASE_DIR}/core"
  cd "${NDM_BASE_DIR}/core" || exit
fi

echo "[entrypoint.sh] launching ndm process."
/usr/sbin/ndm start "$@" &

#sigterm caught SIGTERM signal and forward it to child process
_sigterm() {
  echo "[entrypoint.sh] caught SIGTERM signal forwarding to pid [$child]."
  kill -TERM "$child" 2> /dev/null
  waitForChildProcessToFinish
}

#sigint caught SIGINT signal and forward it to child process
_sigint() {
  echo "[entrypoint.sh] caught SIGINT signal forwarding to pid [$child]."
  kill -INT "$child" 2> /dev/null
  waitForChildProcessToFinish
}

#waitForChildProcessToFinish waits for child process to finish
waitForChildProcessToFinish(){
    while ps -p "$child" > /dev/null; do sleep 1; done;
}

trap _sigint INT
trap _sigterm SIGTERM

child=$!
wait $child