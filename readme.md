# Proxy Konfigurator #

Proxy Konfigurator creates haproxy configurations for Kubernetes services.

## Quick Start ##

`go get -u github.com/stackexchange/haproxy-kubefigurator`

The following block can be used to watch for changes to service specs (via kubernetes - this should be moved to use kubernetes client-go watch functionality on nodes and services instead) and update a centrally stored haproxy configuration (in etcd) on change.

By default, if `--kubeconfig` is not set, the service will operate in in-cluster configuration mode; allowing full functionality with minimal configuration when running in a pod inside the cluster.

```bash
#!/bin/bash

while true
do
    # Watch etcd for k8s service spec changes
    etcdctl --endpoints https://etcd1:2379,https://etcd2:2379,https://etcd3:2379 watch -r /registry/services/specs > /dev/null \
        && echo "$(date): Kubernetes service updated" \
        && /usr/local/bin/haproxy-kubefigurator \
            --etcd-host https://etcd1:2379 \
            --etcd-host https://etcd2:2379 \
            --etcd-host https://etcd3:2379 \
            --etcd-path /service-router/haproxy-config \
            --etcd-ca-file /path/ca.crt \
            --etcd-client-cert-file /path/client.crt \
            --etcd-client-key-file /path/client.key \
            apply
done

```

Once the configuration is saved off to etcd, consumers can load in the config and update themselves.  This should be orchestrated across nodes to prevent service-level outages during updates.  Haproxy needs to be configured to use `/etc/haproxy/dynamic.cfg` as a configuration file for the following example to work:

```bash
#!/bin/bash

while true
do
    etcdctl --endpoints https://etcd1:2379,https://etcd2:2379,https://etcd3:2379 watch /stackexchange.com/haproxy-kubefigurator/config > /dev/null \
        && echo "$(date): HAproxy config updated" \
        && etcdctl --endpoints https://etcd1:2379,https://etcd2:2379,https://etcd3:2379 get /stackexchange.com/haproxy-kubefigurator/config > /etc/haproxy/dynamic.cfg \
        && /usr/local/sbin/haproxy -f /etc/haproxy/haproxy.cfg -f /etc/haproxy/dynamic.cfg -c -q \
        && echo "$(date): HAproxy config check passed; reloading" \
        && systemctl restart haproxy
done

```
