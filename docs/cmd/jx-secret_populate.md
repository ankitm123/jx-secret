## jx-secret populate

Populates any missing secret values which can be automatically generated

### Usage

```
jx-secret populate
```

### Synopsis

Populates any missing secret values which can be automatically generated"

### Examples

  jx-secret populate

### Options

```
  -d, --dir string      the directory to look for the .jx/gitops/secret-schema.yaml file (default ".")
  -h, --help            help for populate
      --no-wait         disables waiting for the secret store (e.g. vault) to be available
  -n, --ns string       the namespace to filter the ExternalSecret resources
  -w, --wait duration   the maximum time period to wait for the vault pod to be ready if using the vault backendType (default 5m0s)
```

### SEE ALSO

* [jx-secret](jx-secret.md)	 - External Secrets utility commands

###### Auto generated by spf13/cobra on 12-Aug-2020