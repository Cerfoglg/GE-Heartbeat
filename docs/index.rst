**GE Heartbeat Documentation**
===============================

.. toctree::
   :maxdepth: 2
   :caption: Contents:

GE Heartbeat Script
---------------------
The main component responsible for generating the Heartbeats is a shell script 
named *GE_Heartbeat.sh*. This script is deployed inside the Docker container or 
Virtual Machine the Generic Enabler is running in. It is meant to be set to run 
periodically, every hour for example, using cron tasks. Each run of the script 
corresponds to a heartbeat of the GE. 

Deploying the script with a GE in a Docker container is relatively straight forward, 
as it simply requires to extend the Dockerfile that builds the GE image to include 
retrieving the script, installing its few dependencies (mainly jq for JSON command-line processing), 
and set a cron task to periodically run the script, as well as a task that checks for 
updated versions of the script (using wget with the -N flag). 

For Ubuntu/Debian::

	# Installing some dependencies
	RUN apt-get update && apt-get install -y wget curl jq cron
	
	# Download the script and make it executable
	RUN wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>
	
	# Add crontab to run the script every hour
	RUN echo '0 * * * * cd / && ./<SCRIPT>' | crontab
	RUN echo '30 0 * * * wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>' | crontab
	
	# Set env variable for GE Id and Version (Eg. Orion, 1.0)
	ENV GENERIC_ENABLER_ID="Orion"
	ENV GENERIC_ENABLER_VERSION="1.0"
	
	# Start cron if not started already
	CMD cron -f

For CentOS::

	# Installing some dependencies
	RUN yum update && yum install -y wget curl cronie
	RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 && chmod +x ./jq-linux64 && cp jq-linux64 /usr/bin/jq
	
	# Download the script and make it executable
	RUN wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>
	
	# Add crontab to run the script every hour
	RUN echo '0 * * * * cd / && ./<SCRIPT>' | crontab
	RUN echo '30 0 * * * wget -N <SCRIPT> -P / && chmod +x /<SCRIPT>' | crontab
	
	# Set env variable for GE Id and Version (Eg. Orion, 1.0)
	ENV GENERIC_ENABLER_ID="Orion"
	ENV GENERIC_ENABLER_VERSION="1.0"
	
	# Start cron if not started already
	CMD crond -n

Deploying the script with a GE in a VM is similar in concept, requiring the script 
to be downloaded and cron to be set up upon launching the VM.

Functionally, what the script does is construct a JSON object, which contains a 
timestamp in nanoseconds (since unix epoch), the ID and Version of the Generic Enabler 
(e.g Orion 1.0), as well as the identifier of the Docker container or VM the GE is 
running in. For Docker this means the ID of the container as defined by Docker, 
while for VMs it corresponds to the Instance ID obtained from the Openstack metadata. 
The GE ID and Version come from two environment variables: **GENERIC_ENABLER_ID** and 
**GENERIC_ENABLER_VERSION**. This JSON object is then sent in a POST request to the 
GE Heartbeat Proxy service, whose location is known with an environment variable 
called **HEARTBEAT_SERVICE_HOST**.

It’s worth noting that the Heartbeat Script does not directly check that the 
Generic Enabler in the container/VM is running. The script works under the assumption 
that as long as the container/VM is up and running then so is the GE.

GE Heartbeat Service
---------------------
The GE Heartbeat script creates a JSON to represent a Heartbeat and posts it to a
Golang microservice, simply referred to as the GE Heartbeat Service. It offers a 
REST API that receives Heartbeats and sends them to Monasca in the form of metrics. 
The main purpose of the service is to handle authentication: obtaining a token from 
Keystone, using a username and password for login details, and posting the heartbeat 
metric to the Monasca API, authenticating using the received token. 
 
The service can be easily deployed with Docker, using the image built with the provided 
Dockerfile. The service can be configured using the configuration.yml file, or through 
environment variables. 

Monasca
---------------------
Metrics in Monasca are defined by a name and optional dimensions. In our case, 
the Heartbeat metrics are all named GE_Heartbeat, and their dimensions contain 
the GE’s Name, Version, and the ID of the container/VM running it. Each heartbeat
corresponds to a metric measurement, which contains the timestamp of the heartbeat,
and a value, which in our case is always 1. 

Monasca allows for metrics, as well as measurements, to be queried from its REST API,
allowing for filtering by dimensions (and thus by GE Name or Version), as well as by
times of measurements. This in turns allows for querying of GEs that were active in 
a given period of time by simply querying the API for metrics, filtered by a start
and end time. For the full specification of the Monasca API and all its endpoints, 
see https://github.com/openstack/monasca-api/blob/master/docs/monasca-api-spec.md

Grafana
---------------------
To have a better way to visualise the heartbeats (the metrics), Grafana can be used in 
combination with Monasca. By using Monasca as a data source Grafana allows for graphs to 
be created that can display, given a period of time, the number of GEs running. 
Since the measurements of the heartbeat metrics always have value 1, it’s easy to 
aggregate multiple metrics by summing those measurements for each heartbeat interval, 
obtaining to number of active GEs at a given time. Depending on the queries, it’s easy 
to obtain the number of running GEs for a given GE, or a given Version of it, and draw 
a graph displaying the number of active instances over time, including max counts over 
the graph. For an example dashboard for Grafana to import, see the provided example.