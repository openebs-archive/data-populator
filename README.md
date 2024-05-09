# data-populator
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/data-populator)](https://goreportcard.com/report/github.com/openebs/data-populator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/mum4k/termdash/blob/master/LICENSE)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fdata-populator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fdata-populator?ref=badge_shield)

Kubernetes controller for loading seed data into a kubernetes persistent volume from another existing kubernetes volume. 

## Scope 

Data populator can be used in the following scenarios:

- Cluster node re-cycle: A kubernetes node needs to be pulled down for either upgrade or maintenance purposes. In this case, data saved into the local storage of the node(to be brought down) should be migrated to another node in the cluster.
- Load the seed into K8s volumes: The data can be pre-populated from an existing PV that will help with scaling the application with static content(without using read-write many).

## Project Status

**Alpha**: Only filesystem volumes are supported by data-populator as of the current release.

## Usuage

### Prerequisites

Before installing data-populator make sure your kubernetes cluster meets the following requirements:

1. Kubernetes version 1.23 or above
2. `AnyVolumeDataSource` feature gate is enabled on the cluster

### Install

Please refer to our [Quickstart](/docs/data-populator/data-populator.md)

## Contributing

Head over to the [CONTRIBUTING.md](./CONTRIBUTING.md).

## Community

- [Join OpenEBS community on Kubernetes Slack](https://kubernetes.slack.com)
    - Already signed up? Head to our discussions at [#openebs](https://kubernetes.slack.com/messages/openebs/)
    - Want to join our contributor community meetings, [check this out](https://github.com/openebs/openebs/blob/HEAD/community/README.md).
- Join our OpenEBS CNCF Mailing lists
    - For OpenEBS project updates, subscribe to [OpenEBS Announcements](https://lists.cncf.io/g/cncf-openebs-announcements)
    - For interacting with other OpenEBS users, subscribe to [OpenEBS Users](https://lists.cncf.io/g/cncf-openebs-users)

## Code of conduct

Participation in the OpenEBS community is governed by the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/HEAD/code-of-conduct.md).

##License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fdata-populator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fdata-populator?ref=badge_shield)
