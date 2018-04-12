#!/usr/bin/env python

"""This module provides methods used for minikube administration."""

from os import environ as os_environ
from os.path import isdir
from os.path import join as path_join
from os import environ as os_environ
from time import sleep
import subprocess

def setup():
    """
    This method starts minikube with `--vm-driver=none` and
    `--feature-gates=MountPropagation=true` options.
    """
    # Start minikube
    # Assumption: This function's caller should be a Super user.
    try:
        print 'Running minikube command'###########
        subprocess.check_call(["sudo", "minikube", "start", "--vm-driver=none",
                               "--feature-gates=MountPropagation=true"])
        print 'Run minikube command'###########
    except subprocess.CalledProcessError as err:
        print 'Subprocess error occured while starting minikube:',\
            err.returncode
        raise err
    except Exception as err:
        print 'Unknown error occured while starting minikube.'
        raise err

    print "os_environ['CHANGE_MINIKUBE_NONE_USER'] =", os_environ['CHANGE_MINIKUBE_NONE_USER']
    if os_environ.get('CHANGE_MINIKUBE_NONE_USER') == 'true':
        # Below commands shall automatically run in this case.
        print 'Returning from setup.'
        return

    # Run the commands required when run minikube as --vm-driver=none
    # Assumption: Environment variables `USER` and `HOME` is well defined.
    commands = [
        "sudo mv /root/.kube " + os_environ["HOME"] + "/.kube",
        "sudo chown -R " + os_environ["USER"] + " " + os_environ["HOME"] + "/.kube",
        "sudo chgrp -R " + os_environ["USER"] + " " + os_environ["HOME"] + "/.kube",
        "sudo mv /root/.minikube " + os_environ["HOME"] + "/.minikube",
        "sudo chown -R " + os_environ["USER"] + " " + os_environ["HOME"] + "/.minikube",
        "sudo chgrp -R " + os_environ["USER"] + " " + os_environ["HOME"] + "/.minikube"
    ]

    # Wait for `.kube` to be created
    print 'Waiting for `.kube` to be created...'
    while True:
        if isdir(path_join(os_environ["HOME"], '.kube')):
            print path_join(os_environ["HOME"], '.kube'), 'created.'
            break
        elif isdir('/root/.kube'):
            print '/root/.kube created.'
            break
        sleep(1)

    # Wait for `.minikube` to be created
    print 'Waiting for `.minikube` to be created...'
    while True:
        if isdir(path_join(os_environ["HOME"], '.minikube')):
            print path_join(os_environ["HOME"], '.minikube'), 'created.'
            break
        elif isdir('/root/.minikube'):
            print '/root/.minikube created.'
            break
        sleep(1)

    for command in commands:
        print 'Running', command
        returncode = subprocess.call(command.split())
        print 'Return code:', returncode
        print

def check_status():
    """
    This method checks minikube status and parse it to a dict.

    :return: dict: minikube status parsed into dict.
    """
    # Caller of this function should have proper rights
    # to check minikube status
    try:
        status_str = subprocess.check_output(["sudo", "minikube", "status"]).strip()
    except subprocess.CalledProcessError as err:
        print 'Subprocess error occured while checking minikube status:',\
            err.returncode
        raise err
    except Exception as err:
        print 'Unknown error occured while checking minikube status.'
        raise err

    status = {}
    for line in status_str.split('\n'):
        key, val = line.split(':', 1)
        status[key.strip()] = val.strip()
    return status

def teardown():
    """This method deletes minikube."""
    # Caller of this function should have proper rights to delete minikube
    try:
        subprocess.check_output(["sudo", "minikube", "delete"])
    except subprocess.CalledProcessError as err:
        print 'Subprocess error occured while deleting minikube:',\
            err.returncode
        raise err
    except Exception as err:
        print 'Unknown error occured while deleting minikube.'
        raise err

def clear_containers():
    """This method removes all the docker containers present on the machine."""
    # CAUTION: This function call deletes all docker containers
    try:
        containers = subprocess.check_output(["sudo", "docker", "ps", "-aq"])
        if containers != '':
            containers = containers.split()
            for container in containers:
                try:
                    subprocess.check_call(["sudo", "docker", "rm", "-f", container])
                except subprocess.CalledProcessError as err:
                    print 'Subprocess error occured',\
                        'while deleting docker containers:', err.returncode
                    raise err
                except Exception as err:
                    print 'Unknown error occured',\
                        'while deleting docker containers.'
                    raise err
    except Exception as err:
        raise err
