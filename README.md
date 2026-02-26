# GKE Authentication Plugin

> Note: This is a hard fork of [this repo](https://github.com/traviswt/gke-auth-plugin). This version adds support for service account impersonation and uses `client.authentication.k8s.io/v1` instead of `v1beta1`

This plugin provides a standalone way to generate an ExecCredential for use by k8s.io/client-go applications or
human users.

Google already provides a [gke-gcloud-auth-plugin](https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke); however,
that plugin depends on the gcloud CLI, which is written in Python. This dependency graph is large, if you want to authenticate and interact with a GKE cluster from a go application.

This plugin is for use outside of a cluster; when running in the cluster, mount a service account and use that token to interact with the Kubernetes API.

## Run

```shell
# generate ExecCredential
gke-auth-plugin

# generate credentials for service account
gke-auth-plugin --impersonate_service_account=${GOOGLE_SERVICE_ACCOUNT_EMAIL}

# version
gke-auth-plugin version
```

You can straight up replace the gke-gcloud-auth-plugin with this binary, or place on your path and update your kubeconfig exec command to run gke-auth-plugin.

### Example Exec Section of Kubeconfig

```yaml
users:
- name: user_id
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: gke-auth-plugin
      provideClusterInfo: true
      interactiveMode: Never
```
