# k8s-multi-secret-to-file
Allows to generate multiple Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) into one file which is not supported by Kubernetes.

## What can this be used for?
Whenever multiple secrets are needed inside files, e.g. when applications are using configuration files that include secrets.

## How does it work?
Assumption is, that applications configuration (files) are stored in [Configmaps](https://kubernetes.io/docs/concepts/configuration/configmap/). Instead of directly mounting the configfile to application containers, its mounted to the k8s-multi-secret-to-file init container.
All secrets needed by the application are also mounted to the init container, allowing a templating mechanism to inject secrets into application configuration. The rendered templates are available to application containers using shared volumes between init containers and regular containers inside one Pod.
