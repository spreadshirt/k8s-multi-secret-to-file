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
      index.html: |-
        <h1>{{ .headline }}</h1>
    ...
    ```

2. provide secrets inside Kubernetes Secrets

    ```yaml
    ...
    data:
      secret-headline: dmFsdWVGcm9tU2VjcmV0
    ...
    ```

3. configure an init-container using `k8s-multi-secret-to-file`, e.g.

    ```yaml
    ...
    initContainers:
      image: ghcr.io/spreadshirt/k8s-multi-secret-to-file:latest
      imagePullPolicy: Always
      name: secret-init
      volumeMounts:
        - mountPath: /etc/rendered
          name: init-share
        - mountPath: /etc/templates/index.html
          name: configmap
          subPath: index.html
    ...
    ```

4. provide secrets as environment variables to the init container. Envs must be prefixed (default: `SECRET_`)

    ```yaml
    ...
    - env:
      - name: SECRET_headline
        valueFrom:
          secretKeyRef:
            name: apache-demo
            key: secret-headline
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
      - mountPath: /etc/templates/index.html
        name: configmap
        subPath: index.html
    ...
    ```

    `/etc/templates` and `/etc/rendered` are the default paths for templates and the results, this can be configured, if necessary

7. mount the rendered config file to the application container

    ```yaml
    ...
    volumeMounts:
    - mountPath: /var/www/html/index.html
      name: init-share
      subPath: index.html
    ...
    ```

8. deploy the application and check the rendered file inside the application container 

    ```sh
    $ kubectl exec <POD_NAME> -c apache -- cat /var/www/html/index.html
    <h1>valueFromSecret</h1>
    ```