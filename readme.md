# Service Router Configurator #

Service Router Configurator creates haproxy configurations for Kubernetes services.

## Quick Start ##

Service Router Configurator depends on `kubectl` on the host system.

The following can be used to run and will use the default `kubectl` context to connect to the kubernetes cluster.

Environment variables:

- ETCDSERVERS: list of etcd servers to connect to

```bash
#!/bin/bash

while true
do
    etcdctl --endpoints "$ETCDSERVERS" watch -r /registry/services/specs > /dev/null \
        && echo "$(date): Kubernetes service updated" \
        && /usr/local/bin/service-router-configurator --etcd-host "$ETCDSERVERS" --etcd-path /service-router/haproxy-config apply
done
```

This will save off the dynamically generated configuration to `/service-router/haproxy-config` in etcd and can be consumed by haproxy nodes.  Haproxy will need to be configured to use `/etc/haproxy/dynamic.cfg` as a configuration file in the below example:

```bash
! /bin/bash

while true
do
    etcdctl --endpoints "$ETCDSERVERS" watch /service-router/haproxy-config > /dev/null \
        && echo "$(date): HAproxy config updated" \
        && etcdctl --endpoints "$ETCDSERVERS" get /service-router/haproxy-config > /etc/haproxy/dynamic.cfg \
        && /usr/local/sbin/haproxy -f /etc/haproxy/haproxy.cfg -f /etc/haproxy/dynamic.cfg -c -q \
        && echo "$(date): HAproxy config check passed; reloading" \
        && systemctl restart haproxy
done
```
