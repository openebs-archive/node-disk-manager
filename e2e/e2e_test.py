#!/usr/bin/env python

"""This module is written to perform end-to-end test for openebs/node-disk-manager"""

# Initialization
print 'Initializing...'
import subprocess

from os.path import isfile
from os import remove as removefile

from json import loads as load_json_str
from time import sleep
from kubernetes.client.rest import ApiException as K8SApiException

import minikube_adm

import ndm_util
from ndm_util import NDM_TEST_YAML_NAME, NDM_TEST_YAML_PATH

import k8s_util
from k8s_util import POD_WAIT_STATES, POD_GOOD_STATES, POD_GOOD_PHASES, POD_BAD_PHASES

def start_minikube(max_try=1, wait_for_sec=1):
    """
    This method starts the minikube with `--vm-driver=none` and
    `--feature-gates=MountPropagation=true` options.

    :param int max_try: maximum number of tries to be make in order to start minikube.
                        (default to 1) (minimum value 1)
    :param int wait_for_sec: ammount of time in seconds to wait where needed.
                             (default to 1) (minimum value 0)
    """

    # Ensuring minimum value for the arguments

    # Making sure that it tries at least once to start minikube
    max_try = max(1, max_try)
    wait_for_sec = max(0, wait_for_sec)

    while True:
        print 'Starting minikube...'
        print
        minikube_adm.setup()
        # Wait for at least 2 sec
        sleep(max(2, wait_for_sec))

        i = 0
        while i < max_try:
            try:
                status = minikube_adm.check_status()
                break
            except Exception as err:
                print 'Error occured while getting status. Error:', str(err)
                sleep(wait_for_sec)
            finally:
                i += 1
        else:
            print '[ERROR] Failed to get minikube status.'
            exit(2)
        print 'Minikube status:', status['minikube']
        print 'Cluster status:', status['cluster']
        print 'kubectl status:', status['kubectl']
        if status['minikube'] != 'Running':
            print 'Deleting minikube...'
            minikube_adm.teardown()
        else:
            break

def apply_ndm_yaml(max_try=1, wait_for_sec=1):
    """
    This method applies the yaml prepared to test.

    :param int max_try: maximum number of tries to be make for applying configuration.
                        (default to 1) (minimum value 1)
    :param int wait_for_sec: ammount of time in seconds to wait where needed.
                             (default to 1) (minimum value 0)
    """
    # Ensuring minimum value for the arguments

    # Making sure that it tries at least once to apply the YAML
    max_try = max(1, max_try)
    wait_for_sec = max(0, wait_for_sec)

    i = 0
    while i < max_try:
        try:
            ndm_util.yaml_apply()
            break
        except Exception as err:
            print 'Error while applying YAML. Error:', str(err)
            sleep(wait_for_sec)
        finally:
            i += 1
    else:
        print '[ERROR] Error getting log.'
        exit(3)

def wait_till_ndm_is_up(max_try=0, wait_for_sec=1):
    """
    This method busy waits until the pod is up or number of try exceeds max_try.
    Each try is make after waiting for wait_for_sec number of seconds.

    :param int max_try: maximum number of tries to be make to check if NDM-Pod is up.
                        (default to 1) (minimum value 1)
    :param int wait_for_sec: ammount of time in seconds to wait after each try.
                             (default to 1) (minimum value 0)
    """

    # Ensuring minimum value for the arguments

    # Making sure that it tries at least once to apply the YAML
    max_try = max(1, max_try)
    wait_for_sec = max(0, wait_for_sec)

    i = 0
    while i < max_try:
        try:
            ndm_pod = k8s_util.get_ndm_pod()
            ndm_pod_state = k8s_util.get_container_state_in_ndm_pod()
            if ndm_pod_state.terminated is not None:
                print 'Pod terminated unexpectedly.'
                exit(4)
            elif ndm_pod_state.waiting is not None:
                if ndm_pod_state.waiting.reason in POD_WAIT_STATES:
                    print 'Waiting as pod-state:', ndm_pod_state.waiting.reason
                    sleep(wait_for_sec)
                elif ndm_pod_state.waiting.reason not in POD_GOOD_STATES:
                    print 'Pod is in bad state:', ndm_pod_state.waiting.reason,\
                        'message:', ndm_pod_state.waiting.message
                    exit(4)
            elif ndm_pod_state.running is None:
                # At this point all states are None,
                # so just showing phase is enough
                print 'Waiting as pod-phase:', k8s_util.get_pod_phase(ndm_pod)
                sleep(wait_for_sec)
            else:
                break
        except K8SApiException as err:
            if err.status == 404:
                print 'Status code: 404 while connecting to REST API.'
                sleep(wait_for_sec)
            else:
                print 'K8SApiException occured',\
                    'while checking if Pod is up. Error:', str(err)
                raise err
        except Exception as err:
            print 'Error occured while checking if Pod is up. Error:', str(err)
            raise err
        finally:
            i += 1
    print 'Pod is up.'

def validate_log(max_try=0, wait_for_sec=1):
    """
    This method validates node-disk-manager log.

    :param int max_try: maximum number of tries to extract log of node-disk-manager.
                        (default to 1) (minimum value 1)
    :param int wait_for_sec: ammount of time in seconds to wait after each try.
                             (default to 1) (minimum value 0)
    """

    # Ensuring minimum value for the arguments

    # Making sure that it tries at least once to apply the YAML
    max_try = max(1, max_try)
    wait_for_sec = max(0, wait_for_sec)

    ndm_pod = k8s_util.get_ndm_pod()
    if ndm_pod.status.phase in POD_GOOD_PHASES:
        # Try to get logs for max_try times
        i = 0
        while i < max_try:
            try:
                ndm_log = k8s_util.get_log(ndm_pod.metadata.name, 'default')
                break
            except K8SApiException as err:
                print 'Error occured while connecting to REST API. Error:',\
                    str(err)
                sleep(wait_for_sec)
            finally:
                i += 1
        else:
            print '[ERROR] Error getting log.'
            exit(5)

        print
        print 'Validating log...'
        if ndm_util.validate_ndm_log(ndm_log) is True:
            print 'Log is OK.'
    elif ndm_pod.status.phase in POD_BAD_PHASES:
        print 'Pod is in bad phase:', ndm_pod.status.phase
        exit(5)
    else:
        print "I am uncertain about pod's this phase:", ndm_pod.status.phase
        exit(5)

def validate_lsblk_output():
    """This method validates lsblk output (inside and outside the pod)."""

    ndm_pod = k8s_util.get_ndm_pod()
    try:
        # exec_to_pod handles the scenario
        # where we pass V1Pod object instead of str, so we handle V1Pod here
        lsblk_in_pod = k8s_util.exec_to_pod(['lsblk', '-J'], ndm_pod)
        # string returned from above is not a valid json.
        # Some values converted to python equivalent like 'null' become 'None'

        # So, need to revert those changes before parsing through library
        lsblk_in_pod = load_json_str(lsblk_in_pod.replace("u'", '"')\
                                  .replace("'", '"').replace('None', 'null'))

        try:
            lsblk_in_host = subprocess.check_output(['lsblk', '-J'])
            lsblk_in_host = load_json_str(lsblk_in_host.replace("u'", '"')\
                                       .replace("'", '"')\
                                       .replace('None', 'null'))

            if ndm_util.match_lsblk_output(lsblk_in_host, lsblk_in_pod):
                print 'lsblk output OK.'
            else:
                print 'lsblk output mismatch.'
                exit(6)
        except Exception as err:
            print 'Error executing `lsblk` in host. Error:', str(err)
            exit(6)
    except Exception as err:
        print 'Error executing `lsblk` inside pod. Error:', str(err)
        exit(6)

def validate_ndm_output():
    """This method validates lsblk output (inside and outside the pod)."""
    ndm_pod = k8s_util.get_ndm_pod()
    try:
        ndm_in_pod = k8s_util.exec_to_pod(['ndm', 'device', 'list'], ndm_pod)

        try:
            ndm_in_host = subprocess.check_output(['../bin/amd64/ndm',
                                                   'device', 'list'])

            print 'In Pod:', ndm_in_pod
            print 'In Host:', ndm_in_host

            if ndm_util.match_ndm_output(ndm_in_host, ndm_in_pod):
                print 'ndm output OK.'
            else:
                print 'ndm output mismatch.'
                exit(7)

        except Exception as err:
            print 'Error executing `ndm` in host. Error:', str(err)
            exit(7)
    except Exception as err:
        print 'Error executing `ndm` inside pod. Error:', str(err)
        exit(7)

def clean():
    """
    This method is intended to clean the residue of the testing.
    It should be run at the very end of the test.
    CAUTION: it calls `minikube_adm.clear_containers`
    which removes all Docker Containers in your machine.
    """

    # Check minikube status and delete if minikube is running
    print 'Checking minikube status...'
    try:
        status = minikube_adm.check_status()
        if status['minikube'] == "Running" or status['minikube'] == "Stopped":
            print 'Deleting minikube...'
            try:
                minikube_adm.teardown()
            except Exception as err:
                print str(err)
        else:
            print 'Machine not present.'
    except Exception as err:
        print str(err)

    # Remove all docker containers
    print
    print 'Removing docker containers...'
    try:
        minikube_adm.clear_containers()
    except Exception as err:
        print str(err)

    print
    print 'Removing', NDM_TEST_YAML_NAME, '...'
    if isfile(NDM_TEST_YAML_PATH+NDM_TEST_YAML_NAME):
        removefile(NDM_TEST_YAML_PATH+NDM_TEST_YAML_NAME)
        print 'Removed.'
    else:
        print 'Not present.'

if __name__ == '__main__':
    # Step: 1 (Commenting it out for now)
    # print 'Making Project...'
    # print
    # make()

    # Step: 2
    print
    print 'Preparing', NDM_TEST_YAML_NAME, '...'
    ndm_util.yaml_prepare()

    # Step: 3
    print
    start_minikube(max_try=5, wait_for_sec=1)

    # Step: 4
    print
    print 'Applying the prepared YAML...'
    print
    apply_ndm_yaml(max_try=5, wait_for_sec=1)

    # Step: 4.5
    print
    print 'Checking if NDM Pod is up...'
    wait_till_ndm_is_up(max_try=120, wait_for_sec=1)

    # Step: 5
    validate_log()

    # Step: 6-1
    print
    print 'Validating lsblk output...'
    validate_lsblk_output()

    # Step: 6-2
    print
    print 'Validating ndm output...'
    validate_ndm_output()

    print 'Success.'

    # print
    # print 'Cleaning...'
    # clean()
    # print 'Done Cleaning.'
