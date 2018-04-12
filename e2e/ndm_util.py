#!/usr/bin/env python

"""This module provides node-disk-manager specific general tools."""

import subprocess

from os import environ as os_environ
from os.path import join as path_join
from re import compile as regex_compile

import yaml # Assumption: PyYAML is installed

import k8s_util

NDM_DIR = path_join(os_environ['GOPATH'], 'src/github.com/openebs/node-disk-manager/')
NDM_YAML = 'node-disk-manager.yaml'
NDM_TEST_YAML_PATH = '/tmp/'
NDM_TEST_YAML_NAME = 'NDM_Test_'+NDM_YAML
NDM_BIN = path_join(NDM_DIR, 'bin/')
NDM_NS = 'default'
DOCKER_IMAGE_NAME = 'openebs/node-disk-manager-' + os_environ.get('XC_ARCH', 'amd64')
try:
    DOCKER_IMAGE_TAG = subprocess.check_output(["git",
                                                "describe",
                                                "--tags",
                                                "--always"]).strip()
except subprocess.CalledProcessError as err:
    print 'Subprocess error occured while getting HEAD name:', err.returncode
    raise err
except Exception as err:
    print 'Unknown error occured while getting HEAD name.'
    raise err

class NDMTestException(Exception):
    """This is NDM test specific Exception"""
    pass

# TODO: Check the pod current status like we do in `kubectl describe`
# Example: Check if all volumes are mounted correctly
# def validateNDMPod(ndm_pod):

def validate_ndm_log(log):
    """
    This method checks the supplied log and checks for any error in the log.

    :param str log: log of node-disk-manager-xxx Pod (required)
    :return: bool: True if log contains no error otherwise return False.
    """

    # Definition of this function should grow as program grows
    if 'Started controller' in log:
        return True
    return False

def yaml_prepare():
    """
    This method reads and parse the configuration file and changes some fields
    so that it can be applied to create node-disk-manager daemonset from recently built image.
    Then it saves that configuration to temp directory which will be cleaned by calling clean()
    """

    # Prepare the yaml
    with open(NDM_DIR+NDM_YAML) as f_ndm_yaml:
        ndm_yaml = yaml.load(f_ndm_yaml)

    # Assumption: In following two statements it is assumed that
    # this pod has only one container

    # Set image name
    print 'Image name:', DOCKER_IMAGE_NAME + ":" + DOCKER_IMAGE_TAG # Remove me
    ndm_yaml['spec']['template']['spec']['containers'][0]['image']\
        = DOCKER_IMAGE_NAME + ":" + DOCKER_IMAGE_TAG

    # set imagePullPolicy
    ndm_yaml['spec']['template']['spec']['containers'][0]['imagePullPolicy']\
        = "IfNotPresent"

    with open(path_join(NDM_TEST_YAML_PATH, NDM_TEST_YAML_NAME), 'w') as f_ndm_test_yaml:
        yaml.dump(ndm_yaml, f_ndm_test_yaml)
    
    # Remove me
    with open(path_join(NDM_TEST_YAML_PATH, NDM_TEST_YAML_NAME)) as f_ndm_test_yaml:
        print f_ndm_test_yaml.read()

def yaml_apply():
    """
    This method apply the yaml prepared for NDM with yaml_prepare()
    """

    k8s_util.yaml_apply(path_join(NDM_TEST_YAML_PATH, NDM_TEST_YAML_NAME))

def match_lsblk_output(host_op, pod_op):
    """
    This method takes `lsblk` command's output on host as well as inside pod,
    then matches them.

    :param dict host_op: output of `lsblk -J` on host parsed into a dictionary (required)
    :param dict pod_op: output of `lsblk -J` inside node-disk-manager-xxx Pod
                        parsed into a dictionary (required)
    :return: bool: True if outputs are equivalent otherwise return False.
    """

    if set(host_op.keys()) != set(pod_op.keys()):
        return False

    for k in host_op.keys():
        # Special Case for mountpoint
        if k == 'mountpoint':
            if host_op['mountpoint'] != pod_op['mountpoint']\
                    and pod_op['mountpoint']\
                    .replace('/etc/hosts/', '/')\
                    .replace('/etc/hosts', '/')\
                    != host_op['mountpoint']:
                return False

        # and if value is a list
        elif isinstance(host_op[k], list)\
            and isinstance(pod_op[k], list):
            for i in xrange(len(host_op[k])):
                # if element of this list will again be a JSON
                if isinstance(host_op[k][i], dict)\
                        and isinstance(pod_op[k][i], dict):
                    if not match_lsblk_output(host_op[k][i],\
                                          pod_op[k][i]):
                        return False
                # Or if it is other then JSON then just direct match
                else:
                    if host_op[k][i] != pod_op[k][i]:
                        return False

        # and for normal cases
        else:
            if host_op[k] != pod_op[k]:
                return False

    # If no negative cases matches then it is OK
    return True

def match_ndm_output(host_op, pod_op):
    """
    This method takes `ndm` binary's output on host as well as inside pod,
    then matches them.

    :param dict host_op: output of `ndm` on host parsed into a dictionary (required)
    :param dict pod_op: output of `ndm` inside node-disk-manager-xxx Pod
                        parsed into a dictionary (required)
    :return: bool: True if outputs are equivalent otherwise return False.
    """
    # Before comparing it removes all characters (which are mostly spaces)
    # except alphanumeric characters, . (Dot), / (Forward Slash),
    # [ (Square bracket open) and ] (Square bracket close), then compares
    # (where mountpoints is an exception as we mount '/' on '/etc/hosts')
    pattern = regex_compile('[^a-zA-Z0-9./\[\]]')
    return pattern.sub('', pod_op).replace('/etc/hosts/', '/')\
        .replace('/etc/hosts', '/') == pattern.sub('', host_op)
