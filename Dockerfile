#
# This Dockerfile builds node-disk-manager
# 

#Set the base image
FROM ubuntu:16.04

#Copy binary to /usr/sbin/ndm
COPY bin/amd64/ndm /usr/sbin/ndm

#Expose port 9090
EXPOSE 9090

#Set the default command
ENTRYPOINT ["ndm", "start"]
