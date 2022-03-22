# data-populator

Kubernetes controller for loading seed data into a kubernetes persistent volume from another existing kubernetes volume. 

## Scope 

Data populator can be used in the following scenarios:

- Cluster node re-cycle: A kubernetes node needs to be pulled down for either upgrade or maintenance purposes. In this case, data saved into the local storage of the node(to be brought down) should be migrated to another node in the cluster.
- Load the seed into K8s volumes: The data can be pre-populated from an existing PV that will help with scaling the application with static content(without using read-write many).

## Status

**Alpha**: Only filesystem volumes are supported by data-populator as of the current release.

## Usuage

### Prerequisites

Before installing data-populator make sure your kubernetes cluster meets the following requirements:

1. Kubernetes version 1.22 or above
2. `AnyVolumeDataSource` feature gate is enabled on the cluster

### Install

Please refer to our [Quickstart](/docs/data-populator/data-populator.md)

