# Quickstart

Rsync Populator is a volume populator that helps to create a volume from any rsync source. Data populator internally makes use of Rsync Populator to load data into a volume. When a DataPopulator CR is created it sets up rsync source on the source PVC and creates a RsyncPopulator CR and a new PVC pointing to that rsync populator as a data source.

The following things are required to use data populators:
1. Install a CRD for the specific populators
2. Install the populator controllers itself
3. Scaling down the application if the source volume is being consumed by it.

## Copying data from one volume into another

1. Install data populator operator

    ```console
    kubectl apply -f https://raw.githubusercontent.com/openebs/data-populator/master/deploy/data-populator-operator.yaml
    ```
    **NOTE:** `openebs-data-population` namespace is reserved for populator and no pvc with `dataSourceRef` should be created in this namespace as the controller ignores PVCs in its own working namespace.

2. Preparing a volume which will act as the source for data populator.
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
      
3. After writing the data, scale down the application
    ```console
   kubectl scale deployment sample-app --replicas=0
   ```

4. Create an instance of the DataPopulator CR
    ```console
    apiVersion: openebs.io/v1alpha1
    kind: DataPopulator
    metadata:
      name: sample-data-populator
    spec:
      # Name of the pvc to copy data from
      sourcePVC: sample-pvc
    
      # Namespace in which the source pvc is present
      sourcePVCNamespace: default
    
      # specification of pvc into which the source data
      # is copied. It has has all the `PersistentVolumeClaim`
      # spec attributes which will be used to create the 
      # destination PVC
      destinationPVC:
        storageClassName: cstor-csi
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
   ```
   
   **NOTE:** Destination PVC will be created in the same namespace as the data populator instance. Also `destinationPVC` field in the above CR .
   
5. Wait for the data populator to come to `WaitingForConsumer` or `Completed` state
    ```console
    abhishek@abhishek-Mayadata:~$ kubectl get datapopulator.openebs.io/sample-data-populator -o=jsonpath="{.status.state}{'\n'}"
    Completed
   ```
   
6. Edit the deployment spec to point to the new pvc and deploy it again. You can get the name of the new/destination pvc by using the below command.
    ```console
   abhishek@abhishek-Mayadata:~$ kubectl get pvc -A
   NAMESPACE   NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS         AGE
   default     sample-pvc              Bound    pvc-bba632ae-98f2-4ee5-abad-b2673caf4835   2Gi        RWO            openebs-hostpath     14d
   default     sample-pvc-populated    Bound    pvc-ab91512c-8514-4952-b558-d711cd21af56   2Gi        RWO            openebs-hostpath-1   20h
   ```

7. After deploying the application, check whether the older data is present or not in the new pvc.
    ```console
    abhishek@abhishek-Mayadata:~$ kubectl exec -it sample-app-156418-70iae sh
    / # cd /data
    /data # ls
    file
    /data # cat file
    hello!
    /data # exit
   ```
