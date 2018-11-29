#!/bin/bash

ulimit -c unlimited
env GOTRACEBACK=crash /usr/sbin/ndm start
