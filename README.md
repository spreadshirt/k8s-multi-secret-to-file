# k8s-multi-secret-to-file
Allows to template multiple Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) into one file which is not supported by Kubernetes.

## What can this be used for?
Whenever multiple secrets are needed inside files, e.g. when applications are using configuration files that include secrets.

## How does it work?
We assume that applications configuration (files) are stored in [Configmaps](https://kubernetes.io/docs/concepts/configuration/configmap/). Instead of directly mounting the configfile to application containers, it is mounted to the k8s-multi-secret-to-file init container.
All secrets needed by the application are also mounted to the init container, allowing the [templating](https://pkg.go.dev/text/template) mechanism to inject secrets into application configuration. The rendered templates are available to application containers using shared volumes between init containers and regular containers inside one Pod.

## Development
The example inside `examples/simple` can be used for development when operating locally. Example docker command for local use is:
`docker run -v $(pwd)/tests/secrets:/var/run/secrets/spreadgroup.com/multi-secret/secrets -v $(pwd)/examples/simple/templates:/var/run/secrets/spreadgroup.com/multi-secret/templates -v $(pwd)/examples/rendered:/var/run/secrets/spreadgroup.com/multi-secret/rendered docker.io/library/k8s-multi-secret-to-file:local`
Make sure the directory mounted to `/var/run/secrets/spreadgroup.com/multi-secret/rendered` already exists.
### CLI parameters
| Parameter               | Default value  | Description                                                                                                                          |
|-------------------------|:--------------:|--------------------------------------------------------------------------------------------------------------------------------------|
| continue-on-missing-key |     false      | Templating of configfiles fails if a (secret) key is missing. This flag allows to continue on missing keys.                          |
| left-delimiter          |       {{       | Left delimiter for internal templating. Change if this delimiter conflicts with your config format.                                  |
| right-delimiter         |       }}       | Right delimiter for internal templating. Change if this delimiter conflicts with your config format.                                 |
| secret-path             |  /etc/secrets  | Path were secrets are read from. Can be changed for local development or testing. Should not matter when using the container.        |
| target-base-dir         | /etc/rendered  | Path were rendered files are stored. Can be changed for local development or testing. Should not matter when using the container.    |
| template-base-dir       | /etc/templates | Path were template files are read from. Can be changed for local development or testing. Should not matter when using the container. |

## Setup
A working example can be found in `examples/k8s`. The files inside manifests directory can be deployed to a Kubernetes cluster.

1. wrap a configfile inside a configmap, and use `{{ index .Secrets "<secret name>" "<secret key>" }}` as placeholder. (using the `index` function here to allow special chars in secret names and keys)

    ```yaml
    ...
    data:
      secret-config: |-
        key1={{ index .Secrets "apache-demo" "secret1" }}
        key2={{ index .Secrets "apache-demo" "secret2" }}
    ...
    ```

2. provide secrets inside Kubernetes Secrets

    ```yaml
    ...
    data:
      secret1: dmFsdWVGcm9tU2VjcmV0
      secret2: Mm5kVmFsdWVGcm9tU2VjcmV0
    ...
    ```

3. configure an init-container using `k8s-multi-secret-to-file`, e.g.

    ```yaml
    ...
    initContainers:
      image: ghcr.io/spreadshirt/k8s-multi-secret-to-file:latest
      imagePullPolicy: Always
      name: secret-init
    ...
    ```

4. provide secrets as files to the init container. `/var/run/secrets/spreadgroup.com/multi-secret/secrets` is the default secret path inside the init-container. For each Secret, configure the volumes:
    ```yaml
    ...
    volumes:
      - name: apache-demo
        secret:
          defaultMode: 420
          secretName: apache-demo
    ...
    ```
   
   and the volumeMounts:
    ```yaml
    ...
    volumeMounts:
      - mountPath: /var/run/secrets/spreadgroup.com/multi-secret/secrets/apache-demo
        name: apache-demo
        readOnly: true
    ...
    ```

5. configure [Volumes](https://kubernetes.io/docs/concepts/storage/volumes/) to allow interaction between the init-container and the application

    ```yaml
    ...
    volumes:
    - emptyDir: { }
      name: init-share
    - configMap:
        name: apache-demo-cfg
      name: configmap
    ...
    ```

6. mount configfile and target path to init-container (IMPORTANT: don't mount the configfile directly to the application container!)

    ```yaml
    ...
    volumeMounts:
      - mountPath: /var/run/secrets/spreadgroup.com/multi-secret/rendered
        name: init-share
      - mountPath: /var/run/secrets/spreadgroup.com/multi-secret/templates/path/to/secret/config
        name: configmap
        subPath: secret-config
    ...
    ```

    `/var/run/secrets/spreadgroup.com/multi-secret/templates` and `/var/run/secrets/spreadgroup.com/multi-secret/rendered` are the default paths for templates and the results, this can be configured, if necessary

7. mount the rendered config file to the application container

    ```yaml
    ...
    volumeMounts:
    - mountPath: /path/to/secret/config
      name: init-share
      subPath: path/to/secret/config
    ...
    ```

8. deploy the application and check the rendered file inside the application container 

    ```sh
    $ kubectl exec apache-demo-8479b98dd4-82jnn -c apache -- cat /path/to/secret/config
   key1=valueFromSecret
   key2=2ndValueFromSecret
    ```
   
## Advanced features
### getValueByOrderedKeys
This extension allows you to use dynamic keys in your secrets, which could happen if you use an external Secret operator for example.

When is this useful? E.g. if you use keys inside Secrets to represent some kind of hierarchical configuration:
```yaml
value: my-value
value_aws: aws-specific-value
value_azure: azure-specific-value
```

Use the following line to access it: `value={{ getValueByOrderedKeys (index .Secrets "secretValues") "value_aws" "value" }}`
This will return the value for the first matching key in `.Secrets.secretValues`, `aws-specific-value` but would fall back to `my-value` in case it does not find a value for key `value_aws`.