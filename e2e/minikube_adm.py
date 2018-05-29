#!/usr/bin/env python

"""This module provides methods used for minikube administration."""

import subprocess

def setup():
    """
    This method starts minikube with `--vm-driver=none` and
    `--feature-gates=MountPropagation=true` options.
    """
    # Start minikube
    # Assumption: This function's caller should be a Super user.
    try:
        subprocess.check_call(["minikube", "start", "--vm-driver=none",
                               "--feature-gates=MountPropagation=true"])
    except subprocess.CalledProcessError as err:
        print 'Subprocess error occured while starting minikube:',\
            err.returncode
        raise err
    except Exception as err:
        print 'Unknown error occured while starting minikube.'
        raise err

    # Run the commands required when run minikube as --vm-driver=none
    # Assumption: Environment variables `USER` and `HOME` is well defined.
    commands = [
        "mv /root/.kube $HOME/.kube",
        "chown -R $USER $HOME/.kube",
        "chgrp -R $USER $HOME/.kube",
        "mv /root/.minikube $HOME/.minikube",
        "chown -R $USER $HOME/.minikube",
        "chgrp -R $USER $HOME/.minikube"
    ]

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
        status_str = subprocess.check_output(["minikube", "status"]).strip()
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
        subprocess.check_output(["minikube", "delete"])
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
        containers = subprocess.check_output(['docker', 'ps', '-aq'])
        if containers != '':
            containers = containers.split()
            for container in containers:
                try:
                    subprocess.check_call(['docker', 'rm', '-f', container])
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
