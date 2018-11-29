#!/bin/bash

ulimit -c unlimited
echo "/var/ndm/core.%e.%p.%h.%t" > /proc/sys/kernel/core_pattern
env GOTRACEBACK=crash /usr/sbin/ndm start
