---
name: Bug Report
about: Tell us about a problem you are experiencing
labels: Bug

---

**What steps did you take and what happened:**
[A clear and concise description of what the bug is, and the sequence of operations performed / commands you ran.]


**What did you expect to happen:**


**The output of the following commands will help us better understand what's going on**:
[Pasting long output into a [GitHub gist](https://gist.github.com) or other pastebin is fine.]

* `kubectl get pods -n openebs`
* `kubectl get blockdevices -n openebs -o yaml`
* `kubectl get blockdeviceclaims -n openebs -o yaml`
* `kubectl logs <ndm daemon pod name> -n openebs`
* `lsblk` from nodes where ndm daemonset is running 

**Anything else you would like to add:**
[Miscellaneous information that will assist in solving the issue.]


**Environment:**
- OpenEBS version
- Kubernetes version (use `kubectl version`):
- Kubernetes installer & version:
- Cloud provider or hardware configuration:
- Type of disks connected to the nodes (eg: Virtual Disks, GCE/EBS Volumes, Physical drives etc)
- OS (e.g. from `/etc/os-release`):
