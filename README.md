# BBSim Sadis Server

This project is designed to aggregate Sadis entries from multiple BBSim instances running on the same kubernetes cluster.
This tool assumes that:
- The sadis service is exposed on the default port `50074`
- BBSim(s) are deployed with the default label `app=bbsim`

## Deploy

```shell
kubectl create configmap kube-config --from-file=kube_config=$KUBECONFIG
kubectl apply -f deployments/bbsim-sadis-server.yaml
```