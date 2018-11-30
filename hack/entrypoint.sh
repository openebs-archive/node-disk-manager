#!/bin/bash

ulimit -c unlimited

echo "/var/openebs/core.%e.%p.%h.%t" > /proc/sys/kernel/core_pattern

env GOTRACEBACK=crash /usr/sbin/ndm start
