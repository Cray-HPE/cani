## cani alpha session init

Initialize and start a session. Will perform an import of system's inventory format.

### Synopsis

Initialize and start a session. Will perform an import of system's inventory format.

```
cani alpha session init [flags]
```

### Options

```
      --csm-api-host string            (CSM Provider) Host or FQDN for authentation and APIs
  -k, --csm-insecure-https             (CSM Provider) Allow insecure connections when using HTTPS to CSM services
      --csm-keycloak-password string   (CSM Provider) Keycloak password
      --csm-keycloak-username string   (CSM Provider) Keycloak username
      --csm-kube-config string         (CSM Provider) Path to the kube config file
  -S, --csm-simulator                  (CSM Provider) Use simulation environment URLs
      --csm-url-hsm string             (CSM Provider) Base URL for the Hardware State Manager (HSM)
      --csm-url-sls string             (CSM Provider) Base URL for the System Layout Service (SLS)
  -h, --help                           help for init
```

### Options inherited from parent commands

```
      --config string   Path to the configuration file
  -D, --debug           additional debug output
  -v, --verbose         additional verbose output
```

### SEE ALSO

* [cani alpha session](cani_alpha_session.md)	 - Interact with a session.

###### Auto generated by spf13/cobra on 7-Aug-2023
