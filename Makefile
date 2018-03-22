# Specify the name for the binaries
NODE_DISK_MANAGER=ndm
DOCKER_IMAGE_NAME?= openebs/node-disk-manager
DOCKER_IMAGE_TAG?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

# Use this to build only the node-disk-manager.

build: header ndm docker

header:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@echo

ndm:
	@echo '--> Building binary...'
	@CTLNAME=${NODE_DISK_MANAGER} sh -c "'$(PWD)/hack/build.sh'"
	@echo '--> Built binary.'
	@echo

deps: header
	@echo '--> Resolving dependencies...'
	dep ensure
	@echo '--> Done resolving.'
	@echo

docker:
	@echo "--> Building docker image..."
	@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .
	@echo "--> Build docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"
	@echo

clean: header
	@echo '--> Cleaning directory...'
	rm -rf bin
	rm -rf ${GOPATH}/bin/${NODE_DISK_MANAGER}
	rm -rf ${GOPATH}/pkg/*
	@echo '--> Done cleaning.'
	@echo

.PHONY: all build
