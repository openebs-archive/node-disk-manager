# Development Workflow

## Prerequisites

* You have Go 1.19 or above installed on your local host/development machine.
* You have Docker installed on your local host/development machine. Docker is required for building NDM container images and to push them into a Kubernetes cluster for testing. 
* You have `kubectl` installed. For running integration tests, you will require an existing single node cluster. Don't worry if you don't have access to the Kubernetes cluster, raising a PR with the NDM repository will run integration tests for your changes against a Minikube cluster.


## Install the build dependencies.
  * By default node-disk-manager enables fetching disk attributes using udev. This requires udev develop files. For Ubuntu, `libudev-dev` package should be installed.
  * Run `make bootstrap` to install the required Go tools
  * node-disk-manager uses OpenSeaChest to fetch certain details of the disk like temperature and rotation rate. This requires cloning the `openSeaChest` repo **outside the node-disk-manager repo**
    ```sh
    git clone --recursive --branch Release-19.06.02 https://github.com/openebs/openSeaChest.git
    cd openSeaChest/Make/gcc
    make release
    ```
  * Copy the generated static library files to `/usr/lib`
    ```sh
    cd ../../
    sudo cp opensea-common/Make/gcc/lib/libopensea-common.a /usr/lib
    sudo cp opensea-operations/Make/gcc/lib/libopensea-operations.a /usr/lib
    sudo cp opensea-transport/Make/gcc/lib/libopensea-transport.a /usr/lib
    ```

## Building and Testing your changes

* run `make` in the node-disk-manager directory. It will:
  * Build the binary.
  * Build the docker image with the binary.

* To build multi-arch images, you need to have `buildx` installed. Follow these [instructions](https://github.com/docker/buildx#installing) to install buildx.

  Supported architectures are: amd64, arm64, armv7, ppc64le. 

  * NDM multi-arch images are built using QEMU mode(**NOTE**: This can take a lot of time since it runs in an emulated environment). If you're on linux, run :

    ``` docker run --rm --privileged linuxkit/binfmt:v0.8 ```

  * Set the value of environment variable of `IMAGE_ORG` based on your user name at Dockerhub or any other registry which supports storing multi-arch images (This is required because multi-arch images cannot be stored locally, so `buildx` pushes to the registry after building):  

    ``` export IMAGE_ORG=name ```
  * Depending on what you're working on, run:
    * For node-disk-manager:

      ``` make docker.buildx.ndm ```
    * For node-disk-operator:

      ``` make docker.buildx.ndo ```
    * For node-disk-exporter:

      ``` make docker.buildx.exporter ```

* Test your changes
  * `sudo -E env "PATH=$PATH" make test` execute the unit tests
  * Integration tests are written in ginkgo and run against a minikube cluster. Minikube cluster should be running so as to execute the tests. To install minikube follow the doc [here](https://kubernetes.io/docs/tasks/tools/install-minikube/). 
  `make integration-test` will run the integration tests on the minikube cluster.

#### Push Image
By default, Github Action pushes the docker image to `openebs/node-disk-manager`, with *ci* tags.
You can push to your custom registry and modify the ndm-operator.yaml file for your testing. 
