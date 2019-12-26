#!/bin/bash

export GOTRACEBACK=crash

echo "[entrypoint.sh] enabling core dump."
ulimit -c unlimited
echo "[entrypoint.sh] creating /var/openebs/sparse if not exists."
mkdir -p /var/openebs/sparse
echo "[entrypoint.sh] changing directory to /var/openebs/sparse"
cd /var/openebs/sparse || exit
echo "[entrypoint.sh] launching ndm process."
/usr/sbin/ndm start &

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