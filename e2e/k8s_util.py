#!/usr/bin/env python

"""This module provides methods to administrate k8s tailered for our use-cases."""

import subprocess

from time import sleep
from os import popen as os_popen
from os import system as os_exec

# Assumption: kubernetes python client is installed
from kubernetes import client, config
from kubernetes.client import Configuration
from kubernetes.client.apis import core_v1_api
from kubernetes.client.apis import extensions_v1beta1_api
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
from kubernetes.client.models.v1_pod import V1Pod

from yaml import load as yaml_load

# Different phases of Namespace
NS_GOOD_PHASE = ['Active']

# Different phases of Pod
POD_WAIT_PHASES = ['Pending']
POD_GOOD_PHASES = ['Running']
POD_BAD_PHASES = ['Error']

# Different states of Pod
POD_WAIT_STATES = ['ContainerCreating', 'Pending']
POD_GOOD_STATES = ['Running']
POD_BAD_STATES = ['CrashLoopBackOff', 'ImagePullBackOff', 'RunContainerError']

class NDMTestException(Exception):
    """This is NDM test specific Exception"""
    pass

def get_all_namespaces_V1NamespaceList():
    """
    This method returns V1NamespaceList of all the namespaces.

    :return kubernetes.client.models.v1_namespace_list.V1NamespaceList: list of namespaces.
    """

    config.load_kube_config()
    api = core_v1_api.CoreV1Api()
    return api.list_namespace()

def get_all_namespaces_dict():
    """
    This method returns list of the names of all the namespaces.

    :return: dict: dictionary of namespaces where key is namespace name (str)
        and value is corresponding kubernetes.client.models.v1_namespace.V1Namespace object.
    """

    namespaces_list = get_all_namespaces_V1NamespaceList()

    namespaces = {}
    for ns in namespaces_list.items:
        namespaces[ns.metadata.name] = ns
    return namespaces

def get_ndm_pod():
    """
    This method returns Pod object of node-disk-manager.

    :return: kubernetes.client.models.v1_pod.V1Pod: node-disk-manager Pod object.
    """
    config.load_kube_config()
    api_client = client.CoreV1Api()

    # Try to get node-disk-manager for 5 times as sometime code reaches
    # when pod is not even in ContainerCreating state
    i = 0
    ndm_pod = None
    while ndm_pod is None and i < 5:
        sleep(1)

        # List pods
        # Assumption: NDM pod runs under 'dafault' namespace.
        pods = api_client.list_namespaced_pod('default')

        # Find NDM Pod
        for pod in pods.items:
            #Assumption: Pod name starts with string 'node-disk-manager'.
            if pod.metadata.name.startswith('node-disk-manager'):
                ndm_pod = pod
                break
        i += 1
    if ndm_pod is None:
        print 'Failed getting NDM-Pod.'
        exit(1)

    return ndm_pod

def get_pod_phase(pod):
    """
    This method returns phase of the pod passed.

    :param kubernetes.client.models.v1_pod.V1Pod pod: pod object for which
                                                      you want to get phase. (required)
    :return: str: phase of the pod.
    """
    return pod.status.phase

def get_container_state_in_ndm_pod(container_index=0, wait_for_sec=10):
    """
    This method returns the state of the container of supplied index.

    :param container_index: index of the container for which you want state. (default is 0)
    :param wait_for_sec: maximum time to get the container's state in seconds.
                         This method does not very strictly obey this param. (default to 10)
    :return: str: state of the container.
    """
    waited = 0
    pod = get_ndm_pod()
    while pod.status.container_statuses is None and waited < wait_for_sec:
        sleep(1)
        waited += 1
        pod = get_ndm_pod()
    if not waited < wait_for_sec:
        raise NDMTestException('Pod had no container till '\
                               + str(wait_for_sec) + ' seconds.')

    while len(pod.status.container_statuses) <= container_index\
            and waited < wait_for_sec:
        sleep(1)
        waited += 1
        pod = get_ndm_pod()
    if not waited < wait_for_sec:
        raise NDMTestException('Pod did not had ' + str(container_index+1)\
                               + ' containers till ' + str(wait_for_sec)\
                               + ' seconds.')

    return pod.status.container_statuses[container_index].state

def get_node_names():
    """
    This method returns a list of the name of all the nodes.

    :return: list: list of node names (list of str).
    """
    config.load_kube_config()
    api_client = client.CoreV1Api()

    # Assumption: Cluster has only one node. (which is true for minikube).
    return [node.metadata.name for node in api_client.list_node().items]

def label_node(node_name, key, value):
    """
    This method label the node with the given key and value.

    :param srt node_name: Name of the node. (required)
    :param str key: Key of the label. (required)
    :param str value: Value of the label. (required)
    :return: kubernetes.client.models.v1_node.V1Node: Node which is labeled.
    """
    config.load_kube_config()
    api_client = client.CoreV1Api()
    body = {
        "metadata": {
            "labels": {
                key: value
            }
        }
    }
    return api_client.patch_node(node_name, body)

def yaml_apply(yaml_path):
    """
    This method apply the yaml specified by the argument.

    :param str yaml_path: Path of the yaml file that is to be applied.
    """

    # Applying through API call
    try:
        with open(yaml_path) as conf_file:
            apply_ds_from_manifest_yaml_str(conf_file.read())
        return
    except Exception as err:
        print 'Error occured while applying yaml through API. Error:', str(err)

    # When Applying through API call does not work apply through subprocess
    try:
        subprocess.check_call(["kubectl", "apply", "-f", yaml_path])
        return
    except subprocess.CalledProcessError as err:
        print 'Subprocess error occured while applying the prepared YAML:',\
            err.returncode
    except Exception as err:
        print 'An error occured while applying the prepared YAML through subprocess.',\
            'Error:', str(err)

    # When applying through subprocess does not work apply through os.system
    try:
        os_exec('kubectl apply -f ' + yaml_path)
    except Exception as err:
        print 'An error occured while applying the prepared YAML through os module.',\
            'Error:', str(err)
        raise err

def apply_ds_from_manifest_yaml_str(manifest):
    """
    This method apply daemonset yaml from yaml string.

    :param str manifest: manifest describe the daemonset configuration.
    """

    return apply_ds_from_manifest_dict(yaml_load(manifest))

def apply_ds_from_manifest_dict(manifest):
    """
    This method apply daemonset yaml from manifest dict.

    :param dict manifest: manifest describe the daemonset configuration.
    """

    config.load_kube_config()
    api = extensions_v1beta1_api.ExtensionsV1beta1Api()
    return api.create_namespaced_daemon_set(body=manifest, namespace='default')

def exec_to_pod(command, pod_name, namespace='default'):
    """
    This method uninterractively exec to the pod with the command specified.

    :param list command: list of the str which specify the command. (required)
    :param str/kubernetes.client.models.v1_pod.V1Pod pod_name: Pod name
                                           or V1Pod obj of the pod. (required)
    :param str namespace: namespace of the Pod. (default to 'default')
    :return: str: Output of the command.
    """
    config.load_kube_config()

    # If someone send kubernetes.client.models.v1_pod.V1Pod object in pod_name
    if isinstance(pod_name, V1Pod):
        pod_name = pod_name.metadata.name

    # config.load_kube_config()
    conf = Configuration()
    assert_hostname = conf.assert_hostname
    conf.assert_hostname = False
    Configuration.set_default(conf)
    api = core_v1_api.CoreV1Api()

    try:
        result = stream(api.connect_get_namespaced_pod_exec, pod_name,\
                        namespace, command=command,\
                        stderr=True, stdin=False, stdout=True, tty=False)
    except ApiException as err:
        print 'Execing to NDM-Pod using kubernetes API failed:', str(err)
        try:
            result = subprocess.check_output(['kubectl', 'exec',
                                              '-n', namespace, pod_name]
                                             + ['--'] + command)
        except subprocess.CalledProcessError as err:
            print 'Subprocess error occured',\
                    'while executing command on pod:', err.returncode
        except Exception as err:
            print 'An error occured while executing command on pod:',\
                    str(err)

            try:
                result = os_popen('kubectl exec -n ' + namespace\
                                  + ' ' + pod_name + ' -- '\
                                  + ' '.join(command)).read()
            except Exception as err:
                raise err
    finally:
        # Undoing the previous changes to configuration
        conf.assert_hostname = assert_hostname
        Configuration.set_default(conf)

    return result

def get_log(pod_name, namespace):
    """
    This method returns the log of the pod.

    :param str pod_name: Name of the pod. (required)
    :param str namespace: Namespace of the pod. (required)
    :return: str: Log of the pod specified.
    """
    config.load_kube_config()
    api_client = client.CoreV1Api()
    return api_client.read_namespaced_pod_log(pod_name, namespace)
