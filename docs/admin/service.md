# Introduction

The component responsible for generating the Heartbeats is a shell script
named *GE_Heartbeat.sh*. This script is deployed inside the Docker container or
Virtual Machine the Generic Enabler is running in. It is then set to run
periodically, every hour for example, using cron tasks. Each run of the script
corresponds to a heartbeat.

## Installing the script

Deploying the script with a GE in a Docker container is relatively straight
forward, as it simply requires to extend the Dockerfile that builds the GE image
to include retrieving the script, installing its few dependencies (mainly jq for
JSON command-line processing), and set a cron task to periodically run the
script, as well as a task that checks for updated versions of the script (using
wget with the -N flag) and downloads the updated script if needed.

For Ubuntu/Debian, add the following to the Dockerfile

```text
ENV GENERIC_ENABLER_ID="Orion"
ENV GENERIC_ENABLER_VERSION="1.0"

RUN apt-get update && apt-get install -y wget curl jq cron

RUN wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>

RUN echo 'HEARTBEAT_SERVICE_HOST="http://heartbeat:8080/beat"\nGENERIC_ENABLER_ID="Orion"\nGENERIC_ENABLER_VERSION="1.0"\n* * * * * cd / && ./<SCRIPT>' | crontab
RUN echo '30 0 * * * wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>' | crontab

CMD cron -f
```

For CentOS, add the following to the Dockerfile:

```text
ENV GENERIC_ENABLER_ID="Orion"
ENV GENERIC_ENABLER_VERSION="1.0"

RUN yum update && yum install -y wget curl cronie
RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 &&
chmod +x ./jq-linux64 && cp jq-linux64 /usr/bin/jq

RUN wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>

RUN echo 'HEARTBEAT_SERVICE_HOST="http://heartbeat:8080/beat"\nGENERIC_ENABLER_ID="Orion"\nGENERIC_ENABLER_VERSION="1.0"\n* * * * * cd / && ./<SCRIPT>' | crontab
RUN echo '30 0 * * * wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>' | crontab

CMD crond -n
```

Deploying the script with a GE running in a VM rather than a container is
similar in concept, requiring the script to be downloaded and cron to be set up
upon launching the VM. For already running VMs or Containers the script can be
downloaded and set up with cron in the same way.

## How the script works

Functionally, what the script does is construct a JSON object, which contains a
timestamp in nanoseconds (since unix epoch), the ID and Version of the Generic
Enabler (e.g Orion 1.0), as well as the identifier of the Docker container or VM
the GE is running in. For Docker this means the ID of the container as defined
by Docker, while for VMs it corresponds to the Instance ID obtained from the
Openstack metadata. The GE ID and Version come from two environment variables:
*GENERIC_ENABLER_ID* and *GENERIC_ENABLER_VERSION*. This JSON object is
then sent in a POST request to the GE Heartbeat Proxy service, whose location
is known with an environment variable called *HEARTBEAT_SERVICE_HOST*.

It’s worth noting that the Heartbeat Script does not directly check that the
Generic Enabler in the container/VM is running. The script works under the
assumption that as long as the container/VM is up and running then so is the GE.
