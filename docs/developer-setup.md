# Development Workflow

## Prerequisites

* You have Go 1.9 installed on your local host/development machine.
* You have Docker installed on your local host/development machine. Docker is required for building NDM container images and to push them into a Kubernetes cluster for testing. 
* You have `kubectl` installed. For running integration tests, you will either require to use an existing cluster or NDM will create a Minikube cluster on the local host/development machine. Don't worry if you don't have access to the Kubernetes cluster, raising a PR with the NDM repository will run integration tests for your changes against a Minikube cluster.

## Initial Setup

### Fork in the cloud

1. Visit https://github.com/openebs/node-disk-manager
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone fork to local host

Place openebs/node-disk-manager's code on your `GOPATH` using the following cloning procedure.
Create your clone:

```sh

mkdir -p $GOPATH/src/github.com/openebs
cd $GOPATH/src/github.com/openebs

# Note: Here user= your github profile name
git clone https://github.com/$user/node-disk-manager.git

# Configure remote upstream
cd $GOPATH/src/github.com/openebs/node-disk-manager
git remote add upstream https://github.com/openebs/node-disk-manager.git

# Never push to upstream master
git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
git remote -v
```
> **Note:** If your `GOPATH` has more than one (`:` separated) paths in it, then you should use *one of your go path* instead of `$GOPATH` in the commands mentioned here. This statement holds throughout this document.

Install the build dependencies.
  * By default node-disk-manager enables fetching disk attributes using udev. This requires udev develop files. For Ubuntu, `libudev-dev` package should be installed.
  * Run `make bootstrap` to install the required Go tools
  * node-disk-manager uses OpenSeaChest to fetch certain details of the disk like temperature and rotation rate. This requires cloning the `openSeaChest` repo to `openebs` directory and build it. 
    ```sh
    git clone --recursive https://github.com/openebs/openSeaChest.git
    cd openSeaChest/Make/gcc
    make release
    ```
  * Copy the generated static library files to `/usr/lib`
    ```sh
    cd ../../
    sudo cp opensea-common/Make/gcc/lib/libopensea-common.a /usr/lib
    sudo cp opensea-operations/Make/gcc/lib/libopensea-operations.a /usr/lib
    sudo cp opensea-transport/Make/gcc/lib/libopensea-transport.a /usr/lib
    ```

## Git Development Workflow

### Always sync your local repository:
Open a terminal on your local host. Change directory to the node-disk-manager fork root.

```sh
$ cd $GOPATH/src/github.com/openebs/node-disk-manager
```

 Checkout the master branch.

 ```sh
 $ git checkout master
 Switched to branch 'master'
 Your branch is up-to-date with 'origin/master'.
 ```

 Recall that origin/master is a branch on your remote GitHub repository.
 Make sure you have the upstream remote openebs/node-disk-manager by listing them.

 ```sh
 $ git remote -v
 origin	https://github.com/$user/node-disk-manager.git (fetch)
 origin	https://github.com/$user/node-disk-manager.git (push)
 upstream	https://github.com/openebs/node-disk-manager.git (fetch)
 upstream	https://github.com/openebs/node-disk-manager.git (no_push)
 ```

 If the upstream is missing, add it by using below command.

 ```sh
 $ git remote add upstream https://github.com/openebs/node-disk-manager.git
 ```
 Fetch all the changes from the upstream master branch.

 ```sh
 $ git fetch upstream master
 remote: Counting objects: 141, done.
 remote: Compressing objects: 100% (29/29), done.
 remote: Total 141 (delta 52), reused 46 (delta 46), pack-reused 66
 Receiving objects: 100% (141/141), 112.43 KiB | 0 bytes/s, done.
 Resolving deltas: 100% (79/79), done.
 From github.com:openebs/node-disk-manager
   * branch            master     -> FETCH_HEAD
 ```

 Rebase your local master with the upstream/master.

 ```sh
 $ git rebase upstream/master
 First, rewinding head to replay your work on top of it...
 Fast-forwarded master to upstream/master.
 ```
 This command applies all the commits from the upstream master to your local master.

 Check the status of your local branch.

 ```sh
 $ git status
 On branch master
 Your branch is ahead of 'origin/master' by 38 commits.
 (use "git push" to publish your local commits)
 nothing to commit, working directory clean
 ```
 Your local repository now has all the changes from the upstream remote. You need to push the changes to your own remote fork which is origin master.

 Push the rebased master to origin master.

 ```sh
 $ git push origin master
 Username for 'https://github.com': $user
 Password for 'https://$user@github.com':
 Counting objects: 223, done.
 Compressing objects: 100% (38/38), done.
 Writing objects: 100% (69/69), 8.76 KiB | 0 bytes/s, done.
 Total 69 (delta 53), reused 47 (delta 31)
 To https://github.com/$user/node-disk-manager.git
 8e107a9..5035fa1  master -> master
 ```

### Contributing to a feature or bugfix. 

Always start with creating a new branch from master to work on a new feature or bugfix. Your branch name should have the format XX-descriptive where XX is the issue number you are working on followed by some descriptive text. For example:

 ```sh
 $ git checkout master
 # Make sure the master is rebased with the latest changes as described in previous step.
 $ git checkout -b 1234-fix-developer-docs
 Switched to a new branch '1234-fix-developer-docs'
 ```
Happy Hacking!

### Building and Testing your changes

* run `make` in the top directory. It will:
  * Build the binary.
  * Build the docker image with the binary.

* Test your changes
  * `sudo -E env "PATH=$PATH" make test` execute the unit tests
  * `make integration-test` will launch minikube to run the tests. Make sure that minikube can be executed via `sudo -E minikube start --vm-driver=none`

### Keep your branch in sync

[Rebasing](https://git-scm.com/docs/git-rebase) is very import to keep your branch in sync with the changes being made by others and to avoid huge merge conflicts while raising your Pull Requests. You will always have to rebase before raising the PR. 

```sh
# While on your myfeature branch (see above)
git fetch upstream
git rebase upstream/master
```

While you rebase your changes, you must resolve any conflicts that might arise and build and test your changes using the above steps. 

## Submission

### Create a pull request

Before you raise the Pull Requests, ensure you have reviewed the checklist in the [CONTRIBUTING GUIDE](../CONTRIBUTING.md):
- Ensure that you have re-based your changes with the upstream using the steps above.
- Ensure that you have added the required unit tests for the bug fixes or new feature that you have introduced. 
- Ensure your commits history is clean with proper header and descriptions.

Go to the [openebs/node-disk-manager github](https://github.com/openebs/node-disk-manager) and follow the Open Pull Request link to raise your PR from your development branch.



