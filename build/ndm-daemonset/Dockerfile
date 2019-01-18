#
# This Dockerfile builds node-disk-manager
# 

#Set the base image
FROM ubuntu:16.04

RUN apt update

#Copy binary to /usr/sbin/ndm
ADD ndm /usr/local/bin/ndm
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

#Expose port 9090
EXPOSE 9090

#Set the default command
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
