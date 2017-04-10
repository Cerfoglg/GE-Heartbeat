#!/bin/bash

# This script sends a heartbeat (as a metric) to the Monasca API, relaying it through a proxy microservice
# To use it:

# - Fill in the variables below with the location of the proxy.
# - Copy or download the script in the Docker container or VM running the service
# - Set up crontab to execute the script inside the VM/Container every given interval (Eg. every hour)

# Getting date and time since epoch in ms
TIMESTAMP=$(($(date +'%s * 1000 + %-N / 1000000')))
DATE=$(date +"%m-%d-%y_%T")

# Env details
HEARTBEAT_SERVICE_HOST="http://localhost:8080/beat"
GENERIC_ENABLER_ID="Orange"
GENERIC_ENABLER_VERSION="1.0"

# Check if env variables are set
if [ -z ${GENERIC_ENABLER_ID+x} ]; then
  echo $DATE" Error: GENERIC_ENABLER_ID not set"
  echo $DATE" Error: GENERIC_ENABLER_ID not set" >> GE_heartbeat_log
  exit 1
fi

if [ -z ${GENERIC_ENABLER_VERSION+x} ]; then
  echo $DATE" Warning: GENERIC_ENABLER_VERSION not set, set to no_version"
  GENERIC_ENABLER_VERSION="no_version"
fi

# Generating unique ID
if [ -f /proc/self/cgroup ]; then
    ID=$(cat /proc/self/cgroup | grep "cpu:/" | sed 's/\([0-9]\):cpu:\/docker\///g')
    echo "Container ID obtained"
else
    echo "No docker container ID, assuming VM"
    if [ -f GE_instance_uuid ]; then
        echo "Using previously retrieved ID"
        ID=$(echo $(cat GE_instance_uuid))
    else
        echo "Getting new ID"
        #ID=$(curl -sS http://169.254.169.254/latest/meta-data/instance-id)
        #if [ $? == 0 ]; then
        #   echo $ID
        #else
        #  echo "Could not get VM ID"
        #  echo $DATE" Error: Could not get VM ID" >> GE_heartbeat_log
        #  exit 1
        #fi
        ID=$TIMESTAMP":"$RANDOM
        echo $ID >> GE_instance_uuid
    fi
fi

#Sending to proxy
curl -X POST -d '{"ID": "'$ID'", "Enabler_ID": "'$GENERIC_ENABLER_ID'", "Enabler_Version": "'$GENERIC_ENABLER_VERSION'", "Timestamp": "'$TIMESTAMP'"}' $HEARTBEAT_SERVICE_HOST
if [ $? == 0 ]; then
  echo "Sent heartbeat"
  echo $DATE": "$ID$"_"$GENERIC_ENABLER_ID"_"$GENERIC_ENABLER_VERSION >> GE_heartbeat_log
else
  echo "Error: Failed to send Heartbeat"
  echo $DATE" Error: Failed to send Heartbeat" >> GE_heartbeat_log
  exit 1
fi

exit 0
