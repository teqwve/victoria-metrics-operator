# VictoriaMetrics operator
![](https://github.com/VictoriaMetrics/operator/workflows/main/badge.svg)

## Documentation

- quick start [doc](/docs/quick-start.MD)
- design and description of implementation [design](/docs/design.MD)
- high availability [doc](/docs/high-availability.MD)
- network policies [doc](/docs/network-policies.MD)
- operator objects description [doc](/docs/api.MD)
- backup [doc](/docs/backup.MD)




## todo
1) tests
2) documentation

## limitations

- Alertmanager is supported at api monitoring.victoriametrics.com/v1beta1     
- alert relabel is not supported

## development

- operator-sdk verson v0.17.0 +  [https://github.com/operator-framework/operator-sdk]
- golang 1.13 +
- minikube 

start:
```bash
make run
```