# Proxy Konfigurator #

Proxy Konfigurator creates haproxy configurations for Kubernetes services and uses an etcd back-end to store the configurations for consumption by load balancers.  This allows for automatic service endpoint creation when using haproxy on-premises.  It supports in-cluster and out-of-cluster configuration options for interacting with the Kubernetes API and supports certificate based authentication against the etcd cluster.

Its main functionality operates in two loops:

### Producer Loop ###

* `haproxy-kubefigurator` watches for changes to service endpoints
* `haproxy-kubefigurator` watches for changes to service endpoints

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

### Kubernetes Service Configuration

The service configures services based on the following criteria:

* Label `haproxy-kubefigurator.enabled` is set to "yes"
* Service type is a NodePort

All annotations are prefixed by the namespace `haproxy-kubefigurator.` and the name of the port in the NodePort spec.  Let's break down the following example:

```
---

kind: Service
apiVersion: v1
metadata:
  annotations:
    haproxy-kubefigurator.web-ui.hostname: CLUSTER-dashboard.ds.stackexchange.com
  labels:
    k8s-app: kubernetes-dashboard
    haproxy-kubefigurator.enabled: "yes"
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  type: NodePort
  ports:
    - name: web-ui
      port: 443
      targetPort: 8443
  selector:
    k8s-app: kubernetes-dashboard
```

The NodePort being mapped under the service spec has the name `web-ui`--that means the prefix for all configuration options for that service (annotations) will all start with `haproxy-kubefigurator.web-ui.`.  The hostname uses the `CLUSTER` alias, which simply does a text replacement with the value passed via `--cluster` flag.  This can be useful for runtime templating of service URLs.

The following annotations can be used to configure service properties:

* `backends-balance-method`: Method to balance requests across back-ends (default 'roundrobin')
* `backends-use-ssl`: "true" to use TLS between haproxy and back-end (default 'true' for HTTP services; otherwise 'false')
* `backends-verify-ssl`: 'true' to verify certificate chain between haproxy and back-end (default 'false')
* `haproxy-mode`: Listen mode for haproxy front-end (default 'http')
* `hostname`: HTTP hostname to listen on. (default '')
* `listen-ip`: IP to listen on. (default '*')
* `listen-port`: Port for the service to listen on.  Multiple HTTP endpoints can be specified for one port, and haproxy will use SNI if multiple certificates are specified. (default '443')
* `use-ssl`: "true" to use TLS (default 'true' for HTTP services; otherwise 'false')
