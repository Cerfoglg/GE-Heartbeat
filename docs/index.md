# Introduction

The Generic Enabler Heartbeat solution presented here is aimed at providing
information regarding the number of running Fiware Generic Eanblers. It
accomplishes this by providing a simple script to deploy alongside the GEs,
which sends heartbeats periodically to a proxy microservice, which in turn
relays them to a Monasca installation in the form of metrics, where they can
be queried and visualised via Grafana Dashboards.
