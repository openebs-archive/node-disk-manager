#
# This Dockerfile builds node-disk-operator
# 

#Set the base image
FROM @BASEIMAGE@

ARG ARCH

ADD bin/${ARCH}/ndo /usr/local/bin/ndo

ENTRYPOINT ["/usr/local/bin/ndo"]
