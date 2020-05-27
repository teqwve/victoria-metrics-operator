# Network policies

[Network policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) allow you easily restrict the ingress traffic between pods using [k8s labels](https://kubernetes.io/docs/user-guide/labels/).
To keep your cluster safer, it's strongly recommended to enable network policies into prometheus namespace.

# Example

This example will close all inbound communication on the namespace monitoring, and allow only necessary traffic.
**This example has only been tested with the calico provider.**

First, follow the instructions to [add Calico to an existing Kubernetes cluster](http://docs.projectcalico.org/v1.5/getting-started/kubernetes/installation/).

Next, use the following configuration to deny all the ingress (inbound) traffic.
```yaml
cat << EOF | kubectl apply -f -
 apiVersion: networking.k8s.io/v1
 kind: NetworkPolicy
 metadata:
   name: default-deny-all
   namespace: monitoring
 spec:
   podSelector:
     matchLabels:
EOF
```
