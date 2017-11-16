# Introduction

The GE Heartbeat script creates a JSON to represent a Heartbeat and posts it to
a Golang Microservice, simply referred to as the GE Heartbeat Proxy. It offers a
REST API (on the /beat endpoint) that receives Heartbeats and sends them to
Monasca in the form of metrics. The main purpose of the service is to handle
authentication: obtaining a token from Keystone using a username and password
for login details, and posting the heartbeat metric to the Monasca API,
authenticating using the received token.

## Deploying the Service

The service can be easily deployed with Docker, using the image built with the
provided Dockerfile:

```shell
docker build -t IMAGE_NAME .
docker run -d -p 8080:8080 --name heartbeat-proxy IMAGE_NAME
```

## Configuring the Service

The service can be configured using the configuration.yml file when building the
image, or with environment variables overriding it at runtime, directly when
launching the Docker container:

```shell
docker run --rm -it -p 8080:8080 /
--env KEYSTONE_HOST="http://some.host.fiware.org" /
--env KEYSTONE_PORT="4730" /
--env MONASCA_HOST="http://some.monasca.fiware.org" /
--env MONASCA_PORT="8070" /
--env USERNAME="user" /
--env PASSWORD="pass" /
--env PROJECT="service" /
--name heartbeat-proxy IMAGE_NAME
```

## Monasca

The heartbeats sent by the proxy are stored as metrics in Monasca: a monitoring
and logging solution for OpenStack. To read more about Monasca and how to
install it in OpenStack, refer to the
[official wiki](https://wiki.openstack.org/wiki/Monasca>)

Metrics in Monasca are defined by a name and optional dimensions. In our case,
the Heartbeat metrics are all named GE_Heartbeat, and their dimensions contain
the GE’s Name, Version, and the ID of the container/VM running it. Each
heartbeat corresponds to a metric measurement, which contains the timestamp of
the heartbeat, and a value, which in our case is always 1.

Monasca allows for metrics, as well as measurements, to be queried from its REST
API (specifications
[here](https://github.com/openstack/monasca-api/blob/master/docs/monasca-api-spec.md),
allowing for filtering by dimensions (and thus by GE Name or Version), as well
as by times of measurements. This in turns allows for querying of GEs that were
active in a given period of time by simply querying the API for metrics,
filtered by a start and end time.

## Grafana

For visualisation of the heartbeat metrics from Monasca, we make use of Grafana:
a platform for time series analytics. An installation guide can be found at the
[Grafana website](https://grafana.com/).

Grafana can be used in combination with Monasca through a plugin that allows it
to be used as data source, found
[here](https://grafana.com/plugins/monasca-datasource/installation). By using
Monasca as a data source Grafana allows for graphs to be created that can
display, given a period of time, the number of GEs running.
Since the measurements of the heartbeat metrics always have value 1, it’s easy
to aggregate multiple metrics by summing those measurements for each heartbeat
interval, obtaining the number of active GEs at a given time. Depending on the
queries, it’s easy to obtain the number of running GEs for a given GE, or a
given Version of it, and draw a graph displaying the number of active instances
over time, including max counts over the graph.

Example dashboards for visualizing the Monasca metrics can be found [here](https://github.com/Cerfoglg/GE-Heartbeat/tree/master/grafana-dashboards)
