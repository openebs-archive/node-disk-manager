# Node Disk Manager Helm Repository

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
$ helm repo add ndm https://openebs.github.io/node-disk-manager
```

You can then run `helm search repo ndm` to see the charts.

#### Update OpenEBS NDM Repo

Once OpenEBS NDM repository has been successfully fetched into the local system, it has to be updated to get the latest version. The OpenEBS NDM repo can be updated using the following command.

```bash
helm repo update
```

#### Install using Helm 3

- Assign openebs namespace to the current context:
```bash
kubectl config set-context <current_context_name> --namespace=openebs
```

- If namespace is not created, run the following command
```bash
helm install <your-relase-name> ndm/openebs-ndm --create-namespace
```
- Else, if namespace is already created, run the following command
```bash
helm install <your-relase-name> ndm/openebs-ndm
```

