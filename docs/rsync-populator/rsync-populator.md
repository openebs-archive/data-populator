# Rsync Populator

Rsync Populator is a volume populator that helps to create volume from any rsync source. Rsync client is used as volume populator plugin. `RsyncPopulator` CR contains the information of source URL and how to access credentials for the source.

## Prerequisites

1. Kubernetes version 1.22 or above
2. `AnyVolumeDataSource` feature gate is enabled on the cluster

## Quickstart

The following things are required to use data populators:
1. Install a CRD for the rsync populator
2. Install the rsync populator controller itself
3. Scaling down the application if the source volume is being consumed by it.

## Steps to use Rsync Populator

1. Install rsync populator CRD

    ```console
    kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/crds/rsyncpopulator-crd.yaml
    ```

2.  Install rsync populator controller
    ```console
    kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/yamls/rsync-populator.yaml
    ```
    **NOTE:** `openebs-data-population` namespace is reserved for populator and no pvc with `dataSourceRef` should be created in this namespace as the controller ignores PVCs in its own working namespace.
  
3. Preparing a volume which will act as the source for rsync populator.
    - Create a sample pvc. Please feel free to edit the storageclass as per your need.
        ```console
        kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/yamls/sample-pvc.yaml
       ```  
    - Create an application to consume the above volume
        ```console
        kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/yamls/sample-app.yaml
        ```
    - Write some data into the volume
        ```console
        abhishek@abhishek-Mayadata:~$ kubectl exec -it sample-app-75675f-7ci7o sh
        / # cd /data/
        /data # ls -l
        total 0
        /data # echo "hello!" > file
        /data # cat file
        hello!
        /data # exit
        ```
      
4. After writing the data, scale down the application
    ```console
   kubectl scale deployment sample-app --replicas=0
   ```

5. Bring up a rsync source and attach the above volume to it.
    ```console
   kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/yamls/sample-rsync-daemon.yaml
   ```
   
6. Create an instance of the RsyncPopulator CR, with all the rsync source details
    ```console
    apiVersion: openebs.io/v1alpha1
    kind: RsyncPopulator
    metadata:
      name: rsync-populator
    spec:
      # rsync daemon credential used by rsync client to
      # connect to it
      username: user
    
      # password allows you to run authenticated rsync
      # connections to an rsync daemon without user intervention
      password: pass
    
      # rsync clinet needs to contact a remote server
      # runnning a rsync daemon. Client will be use this
      # to connect and get data from daemon.
      # url can be a dns or ip:port.
      url: rsync-daemon.default:873
    
      # source data path on the rsync daemon(remote server) 
      # from which the client will sync data into the
      # destination volume
      path: /data
   ```
   
7. Create a destination pvc in the same namespace as the above RsyncPopulator CR(necessary for the volume populator to work properly) where you want the older data to be cloned
    ```console
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: sample-pvc-populated
    spec:
     #storageClassName: openebs-hostpath
      dataSource:
        apiGroup: openebs.io
        kind: RsyncPopulator
        name: rsync-populator
      accessModes:
      - ReadWriteOnce
      volumeMode: Filesystem
      resources:
        requests:
          storage: 2Gi
   ```

8. Edit the deployment spec to point to the above new pvc and deploy it again.

9. After deploying the application, check whether the older data is present or not in the new pvc.
      ```console
      abhishek@abhishek-Mayadata:~$ kubectl exec -it sample-app-156418-70iae sh
      / # cd /data
      /data # ls
      file
      /data # cat file
      hello!
      /data # exit
      ```
