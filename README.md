# BBSim Sadis Server

This project is designed to aggregate Sadis entries from multiple BBSim instances running on the same kubernetes cluster.

This tool assumes that:
- The sadis service is exposed on the default port `50074`
- BBSim(s) are deployed with the default label `app=bbsim`

This component is part of the the VOLTHA project, more informations at:
https://docs.voltha.org

## Deploy

```shell
helm repo add onf https://charts.opencord.org
helm install bbsim-sadis-server onf/bbsim-sadis-server
```

## Configure ONOS to use `bbsim-sadis-server`

Assuming that `bbsim-sadis-server` was installed in the `default` namespace,
you can use this configuration to point ONOS to it:

```json
{
  "sadis" : {
    "integration" : {
      "url" : "http://bbsim-sadis-server.default.svc:58080/subscribers/%s",
      "cache" : {
        "enabled" : true,
        "maxsize" : 50,
        "ttl" : "PT1m"
      }
    }
  },
  "bandwidthprofile" : {
    "integration" : {
      "url" : "http://bbsim-sadis-server.default.svc:58080/profiles/%s",
      "cache" : {
        "enabled" : true,
        "maxsize" : 50,
        "ttl" : "PT1m"
      }
    }
  }
}
```

For more inforation about the `sadis` application you can refer to: https://github.com/opencord/sadis