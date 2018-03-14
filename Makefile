# Specify the name for the binaries
NODE_DISK_MANAGER=ndmctl

# Use this to build only the node-disk-manager.
ndm:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@CTLNAME=${NODE_DISK_MANAGER} sh -c "'$(PWD)/hack/build.sh'"

deps:
	dep ensure

clean:
	rm -rf bin
	rm -rf ${GOPATH}/bin/${NODE_DISK_MANAGER}
	rm -rf ${GOPATH}/pkg/*

.PHONY: all ndm
