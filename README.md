# k8s-multi-secret-to-file
Allows to template multiple Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) into one file which is not supported by Kubernetes.

## What can this be used for?
Whenever multiple secrets are needed inside files, e.g. when applications are using configuration files that include secrets.

## How does it work?
We assume that applications configuration (files) are stored in [Configmaps](https://kubernetes.io/docs/concepts/configuration/configmap/). Instead of directly mounting the configfile to application containers, it is mounted to the k8s-multi-secret-to-file init container.
All secrets needed by the application are also mounted to the init container, allowing the [templating](https://pkg.go.dev/text/template) mechanism to inject secrets into application configuration. The rendered templates are available to application containers using shared volumes between init containers and regular containers inside one Pod.

## Development
The example inside `examples/simple` can be used for development when operating locally.

## Setup
A working example can be found in `examples/k8s`. The files inside manifests directory can be deployed to a Kubernetes cluster.

1. wrap a configfile inside a configmap, and use `{{ .secretKey }}` as placeholder

    ```yaml
    ...
    data:
      secret-config: |-
        key1={{ .secret1 }}
        key2={{ .secret2 }}
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

4. provide secrets as environment variables to the init container. Envs must be prefixed (default: `SECRET_`)

    ```yaml
    ...
    - env:
      - name: SECRET_secret1
        valueFrom:
          secretKeyRef:
            name: apache-demo
            key: secret1
      - name: SECRET_secret2
        valueFrom:
          secretKeyRef:
            name: apache-demo
            key: secret2
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
      - mountPath: /etc/rendered
        name: init-share
      - mountPath: /etc/templates/path/to/secret/config
        name: configmap
        subPath: secret-config
    ...
    ```

    `/etc/templates` and `/etc/rendered` are the default paths for templates and the results, this can be configured, if necessary

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