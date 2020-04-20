# Contributing to Node Disk Manager(NDM)

NDM uses the standard GitHub pull requests process to review and accept contributions.  There are several areas that could use your help. For starters, you could help in improving the sections in this document by either creating a new issue describing the improvement or submitting a pull request to this repository. The issues are maintained at [openebs/openebs](https://github.com/openebs/openebs/issues?q=is%3Aissue+is%3Aopen+label%3Andm) repository.

* If you are a first-time contributor, please see [Steps to Contribute](#steps-to-contribute).
* If you want to file an issue for a bug or feature request, please see [Filing a issue](#filing-an-issue)
* If you have documentation improvement ideas, go ahead and create a pull request. See [Pull Request checklist](#pull-request-checklist)
* If you would like to make code contributions, please start with [Setting up the Development Environment](#setting-up-your-development-environment).
* If you would like to work on something more involved, please connect with the OpenEBS Contributors. See [OpenEBS Community](https://github.com/openebs/openebs/tree/master/community)

## Steps to Contribute

NDM is an Apache 2.0 Licensed project and all your commits should be signed with Developer Certificate of Origin. See [Sign your work](#sign-your-work). 

* Find an issue to work on or create a new issue. The issues are maintained at [openebs/openebs](https://github.com/openebs/openebs/issues?q=is%3Aissue+is%3Aopen+label%3Andm). You can pick up from a list of [good-first-issues](https://github.com/openebs/node-disk-manager/labels/good%20first%20issue).
* Claim your issue by commenting your intent to work on it to avoid duplication of efforts. 
* Fork the repository on GitHub.
* Create a branch from where you want to base your work (usually master).
* Make your changes. If you are working on code contributions, please see [Setting up the Development Environment](#setting-up-your-development-environment).
* Relevant coding style guidelines are the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](http://peter.bourgon.org/go-in-production/#formatting-and-style).
* Commit your changes by making sure the commit messages convey the need and notes about the commit. The commit message format followed for OpenEBS projects can be found [here](https://github.com/openebs/openebs/blob/master/contribute/git-commit-message.md).
* Push your changes to the branch in your fork of the repository.
* Submit a pull request to the original repository. See [Pull Request checklist](#pull-request-checklist)

## Filing an issue
### Before filing an issue

If you are unsure whether you have found a bug, please consider asking in the [Slack](https://kubernetes.slack.com/messages/openebs) first. If
the behavior you are seeing is confirmed as a bug or issue, it can easily be re-raised in the [issue tracker](https://github.com/openebs/openebs/issues).

### Filing issues

When filing an issue, make sure to answer these seven questions:

1. What version of OpenEBS are you using?
2. Type of disks used and the environment
3. What did you expect to see?
4. What did you see instead?

#### For maintainers
* We are using labelling for issue to track it more effectively. Following are valid labels for the issue.
   - **Bug** - if issue is a **bug to existing feature**
   - **Enhancement** - if issue is a **feature request**
   - **Maintenance**  - if issue is not related to production code. **build, document or test related issues falls into this category**
   - **Question** - if issue is about **querying information about how product or build works, or internals of product**.
   - **Documentation** - if issue is about **tracking the documentation work for the feature**. This label should not be applied to the issue about bug in documentations.
   - **Good First Issue** - if issues is easy to get started with. Please make sure that issue should be ideal for beginners to dive into the code base.
   - **Duplicate** - if issue is **duplicate of another issue**
   - **Help Wanted** - if issue **requires extra attention** from more users/contributors

* We are using following labels for issue work-flow:
   - **Backlog** - if issues has **not been planned for current release cycle**
   - **Release-Note/Closed** - if issue is still open / being worked in a release
   - **Release-Note/Open** - if issue has been **resolved in a release**
   - **Wont Fix** - if the issue will not be worked on
   
**If you want to introduce a new label then you need to raise a PR to update this document with the new label details.**

## Pull Request Checklist
* Rebase to the current master branch before submitting your pull request.
* Commits should be as small as possible. Each commit should follow the checklist below:
  - For code changes, add tests relevant to the fixed bug or new feature.
  - Pass the compilation and tests - includes spell checks, formatting, etc.
  - Commit header (first line) should convey what changed
  - Commit body should include details such as why the changes are required and how the proposed changes help
  - DCO Signed, please refer [signing commit](code-standard.md/sign-your-commits) 
* If your PR is about fixing a issue or new feature, make sure you add a change-log. Refer [Adding a Change log](code-standard.md/adding-a-changelog)
* PR title must follow convention: `<type>(<scope>): <subject>`.

  For example:
  ```
   feat(partition) : add support for partitions
   ^--^ ^-----^   ^-----------------------^
     |     |         |
     |     |         +-> PR subject, summary of the changes
     |     |
     |     +-> scope of the PR, i.e. component of the project this PR is intend to update
     |
     +-> type of the PR.
  ```

    Most common types are:
    * `feat`        - for new features, not a new feature for build script
    * `fix`         - for bug fixes or improvements, not a fix for build script
    * `chore`       - changes not related to production code
    * `docs`        - changes related to documentation
    * `style`       - formatting, missing semi colons, linting fix etc; no significant production code changes
    * `test`        - adding missing tests, refactoring tests; no production code change
    * `refactor`    - refactoring production code, eg. renaming a variable or function name, there should not be any significant production code changes
    * `cherry-pick` - if PR is merged in master branch and raised to release branch(like v0.4.x)

* If your PR is not getting reviewed, or you need a specific person to review it, please reach out to the OpenEBS Contributors. See [OpenEBS Community](https://github.com/openebs/openebs/tree/master/community)

## Code Reviews
All submissions, including submissions by project members, require review. We use GitHub pull requests for this purpose. Consult [GitHub Help](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests) for more information on using pull requests.

* If your PR is not getting reviewed or you need a specific person to review it, please reach out to the OpenEBS Contributors. See [OpenEBS Community](https://github.com/openebs/openebs/tree/master/community)

* If PR is fixing any issues from [github-issues](https://github.com/openebs/openebs/issues) then you need to mention the issue number with link in PR description. like : _fixes https://github.com/openebs/openebs/issues/2961_

* If PR is for bug-fix and release branch(like v0.4.x) is created then cherry-pick for the same PR needs to be created against the release branch. Maintainer of the Project needs to make sure that all the bug fix after RC release are cherry-picked to release branch.

### For maintainers
* We are using labelling for PR to track it more effectively. Following are valid labels for the PR.
   - **Bug** - if PR is a **bug to existing feature**
   - **Enhancement** - if PR is a **feature request**
   - **Maintenance**  - if PR is not related to production code. **build, document or test related PR falls into this category**
   - **Documentation** - if PR is about **tracking the documentation work for the feature**. This label should not be applied to the PR fixing bug in documentations.

* We are using following label for PR work-flow:
   - **pr/hold-review** - if PR is currently being developed, and is not yet ready for review
   - **pr/hold-merge** - if PR is waiting on some dependencies, and should not be merged even if approved
   - **pr/documentation-pending** - if the changes in PR are yet to be documented
   - **pr/documentation-complete** - if the changes / features in the PR are already included in the documentation
   - **pr/release-note-alpha** - the changes in this PR should be included in the alpha feature list in release note
   - **pr/release-note** - the changes in this PR should be included in the release note
   - **pr/upgrade-automated** - upgrade is automated for changes in this PR
   - **pr/upgrade-pending** - the automatic upgrade for the changes in this PR are yet to be done.

* Maintainer needs to make sure that appropriate milestone and project tracker is assigned to the PR.

**If you want to introduce a new label then you need to raise a PR to update this document with the new label details.**

## Setting up your Development Environment

This project is implemented using Go and uses the standard golang tools for development and build. In addition, this project heavily relies on Docker and Kubernetes. It is expected that the contributors:
- are familiar with working with Go
- are familiar with Docker containers
- are familiar with Kubernetes and have access to a Kubernetes cluster or Minikube to test the changes.

For setting up a Development environment on your local host, see the detailed instructions [here](./docs/developer-setup.md).

The NDM design document is available [here](./docs/design.md).

