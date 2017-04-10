#!/bin/bash

# This script sends a heartbeat (as a metric) to the Monasca API
# To use it:

# - fill in the variables below with the location of keystone and monasca,
#   as well as the username, password and tenant for retrieving an auth token for the API.
#   Make sure the provided user and tenant are allowed to create metrics from the Monasca API.
# - Encrypt this script using shc (https://github.com/neurobin/shc), and copy the compiled binary
#   in the Docker container or VM running the service
# - Set up crontab to execute the script inside the VM/Container every given interval (Eg. every hour)

# Getting date and time since epoch in ms
TIMESTAMP=$(($(date +'%s * 1000 + %-N / 1000000')))
DATE=$(date +"%m-%d-%y_%T")

# Keystone/Monasca details
KEYSTONE_HOST="http://172.18.0.8"
KEYSTONE_PORT="5000"
MONASCA_HOST="http://172.18.0.10"
MONASCA_PORT="8070"
USERNAME="mini-mon"
PASSWORD="password"
TENANT="mini-mon"

GENERIC_ENABLER_ID="Orion"
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

# Request Auth token
KEYSTONE_AUTH_RESPONSE=$(curl -sS -X POST $KEYSTONE_HOST:$KEYSTONE_PORT/v2.0/tokens -d '{"auth":{"passwordCredentials":{"username": "'$USERNAME'", "password":"'$PASSWORD'"}, "tenantName":"'$TENANT'"}}' -H "Content-type: application/json")
if [ $? == 0 ]; then
  KEYSTONE_AUTH_TOKEN=$(echo $KEYSTONE_AUTH_RESPONSE | jq '.access.token.id' | sed -e 's/^"//' -e 's/"$//')
  if [ $? == 0 ]; then
    echo "Obtained Token from keystone"
  else
    echo "Error: Failed to resolve token from keystone"
    echo $DATE" Error: Failed to resolve token from keystone" >> GE_heartbeat_log
    exit 1
  fi
else
  echo "Error: Failed to obtain token from keystone"
  echo $DATE" Error: Failed to obtain token from keystone" >> GE_heartbeat_log
  exit 1
fi

# Send heartbeat (create metric) to Monasca
curl -sS -X POST -d \
'{"name": "GE_Heartbeat", "dimensions": {"id": "'$ID'", "enabler_id": "'$GENERIC_ENABLER_ID'", "enabler_version": "'$GENERIC_ENABLER_VERSION'"}, "timestamp": '$TIMESTAMP', "value": 1, "value_meta":{"id": "'$ID'", "enabler_id": "'$GENERIC_ENABLER_ID'", "enabler_version": "'$GENERIC_ENABLER_VERSION'"}}' \
-H "Content-type: application/json" \
-H "Accept: application/json" \
-H "X-Auth-Token: "$KEYSTONE_AUTH_TOKEN \
$MONASCA_HOST:$MONASCA_PORT/v2.0/metrics
if [ $? == 0 ]; then
  echo "Sent heartbeat"
  echo $DATE": "$ID$"_"$GENERIC_ENABLER_ID"_"$GENERIC_ENABLER_VERSION >> GE_heartbeat_log
else
  echo "Error: Failed to send Heartbeat"
  echo $DATE" Error: Failed to send Heartbeat" >> GE_heartbeat_log
  exit 1
fi

exit 0
