#
# This Dockerfile builds node-disk-manager
# 

#Set the base image
FROM @BASEIMAGE@

ARG ARCH

#Copy binary to /usr/sbin/ndm
COPY bin/${ARCH}/ndm /usr/sbin/ndm
COPY build/ndm-daemonset/entrypoint.sh /usr/local/bin/entrypoint.sh

#Expose port 9090
EXPOSE 9090

#Set the default command
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
