/*
Copyright © 2022 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	internalv1alpha1 "github.com/openebs/data-populator/apis/openebs.io/v1alpha1"
)

var (
	//rpGK  = schema.GroupKind{Group: GroupOpenebsIO, Kind: RpKind}
	rpGVR = schema.GroupVersionResource{Group: GroupOpenebsIO, Version: VersionV1alpha1, Resource: RpResource}

	dpGK  = schema.GroupKind{Group: GroupOpenebsIO, Kind: DpKind}
	dpGVR = schema.GroupVersionResource{Group: GroupOpenebsIO, Version: VersionV1alpha1, Resource: DpResource}
)

type controller struct {
	kubeClient    *kubernetes.Clientset
	dynamicClient dynamic.Interface
	dpLister      dynamiclister.Lister
	dpSynced      cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

func RunController(cfg *rest.Config) {
	klog.Infof("Starting data populator controller for %s", strings.ToLower(dpGK.String()))
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		close(stopCh)
		<-sigCh
		os.Exit(1) // second signal. Exit directly.
	}()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create kube client: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create dynamic client: %v", err)
	}

	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 30*time.Second)
	dpInformer := dynamicInformerFactory.ForResource(dpGVR).Informer()
	c := &controller{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		dpLister:      dynamiclister.New(dpInformer.GetIndexer(), dpGVR),
		dpSynced:      dpInformer.HasSynced,
		workqueue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	dpInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleDataPopulator,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.handleDataPopulator(newObj)
		},
		DeleteFunc: c.handleDataPopulator,
	})

	dynamicInformerFactory.Start(stopCh)
	if err := c.run(stopCh); nil != err {
		klog.Fatalf("Failed to run controller: %v", err)
	}
}

func (c *controller) handleDataPopulator(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
	}
	c.workqueue.Add(key)
}

func (c *controller) run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.dpSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runWorker, time.Second, stopCh)
	<-stopCh
	return nil
}

func (c *controller) runWorker() {
	processNext := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
			return nil
		}
		if err := c.syncPopulator(context.TODO(), key, parts[0], parts[1]); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		return nil
	}

	for {
		obj, shutdown := c.workqueue.Get()
		if shutdown {
			return
		}
		if err := processNext(obj); err != nil {
			utilruntime.HandleError(err)
		}
	}
}

func (c *controller) syncPopulator(ctx context.Context, key, namespace, name string) error {
	unstruct, err := c.dpLister.Namespace(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("data populator '%s' in work queue no longer exists", key))
			return nil
		}
		return fmt.Errorf("error getting data populator error: %s", err)
	}

	dataPopulator := internalv1alpha1.DataPopulator{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(),
		&dataPopulator); err != nil {
		return fmt.Errorf("error converting data populator `%s` in `%s` namespace error: %s",
			unstruct.GetName(), unstruct.GetNamespace(), err)
	}

	// If the status is completed or failed then don't perform any action
	if dataPopulator.Status.State == internalv1alpha1.StatusCompleted ||
		dataPopulator.Status.State == internalv1alpha1.StatusFailed {
		return nil
	}

	if dataPopulator.Status.State == "" {
		clone := dataPopulator.DeepCopy()
		clone.Status.State = internalv1alpha1.StatusInProgress
		if err := c.updateDataPopulator(clone); err != nil {
			return fmt.Errorf("error updating status of data populator `%s` in `%s` namespace, error: %s",
				dataPopulator.GetName(), dataPopulator.GetNamespace(), err)
		}
		return nil
	}

	// Create a template config of data populator
	dataPopulatorClone := dataPopulator.DeepCopy()
	dptc, err := templateFromDataPopulator(*dataPopulatorClone)
	if err != nil {
		return fmt.Errorf("error creating template config error: %s", err)
	}

	// Check whether the source pvc is already created so that rsync daemon can work properly
	_, err = c.kubeClient.CoreV1().PersistentVolumeClaims(dataPopulator.Spec.SourcePVCNamespace).
		Get(context.TODO(), dataPopulator.Spec.SourcePVC, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pvc `%s` in `%s` namespace error: %s",
			dataPopulator.Spec.SourcePVC, namespace, err)
	}

	// Create destination PVC to where the data is to be populated
	destinationPvcTemplate := dptc.getDestinationPVCTemplate()
	if err := c.ensurePVC(true, namespace, &destinationPvcTemplate); err != nil {
		return fmt.Errorf("error ensuring pvc(true) `%s` in `%s` namespace, error: %s",
			destinationPvcTemplate.GetName(), namespace, err)
	}

	// Create rsync-populator resource which will take care of populating the destination pvc
	rsyncPopulatorTemplate := dptc.getRsyncPopulatorTemplate()
	if err := c.ensurePopulator(true, namespace, &rsyncPopulatorTemplate); err != nil {
		return fmt.Errorf("error ensuring(true) populator `%s` in `%s` namespace, error: %s",
			rsyncPopulatorTemplate.GetName(), namespace, err)
	}

	destinationPVC, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), destinationPvcTemplate.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting destination pvc `%s` in `%s` namespace error: %s",
			destinationPvcTemplate.Name, namespace, err)
	}

	// Check for the destination pvc's storage class volume binding mode
	sc, err := c.kubeClient.StorageV1().StorageClasses().
		Get(context.TODO(), *destinationPvcTemplate.Spec.StorageClassName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting destination pvc's storage class `%s` error: %s",
			*destinationPVC.Spec.StorageClassName, err)
	}

	waitForFirstConsumer := false
	if sc.VolumeBindingMode != nil && *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		waitForFirstConsumer = true
	}

	selectedNode := ""
	if destinationPVC.GetAnnotations() != nil {
		selectedNode = destinationPVC.GetAnnotations()[nodeNameAnnotation]
	}

	if selectedNode == "" && waitForFirstConsumer {
		// Wait for the destination PVC to get a node name before continuing.
		// Update the status of data-populator accordingly
		if dataPopulator.Status.State != internalv1alpha1.StatusWaitingForConsumer {
			dpClone := dataPopulator.DeepCopy()
			dpClone.Status.State = internalv1alpha1.StatusWaitingForConsumer
			if err := c.updateDataPopulator(dpClone); err != nil {
				return fmt.Errorf("error updating status of data populator `%s` in `%s` namespace, error: %s",
					dataPopulator.GetName(), dataPopulator.GetNamespace(), err)
			}
		}
		return nil
	}

	// Check for the finalizer which is added by the rsync-populator which is there till
	// the data population is not completed. This will help us to know whether population of
	// data is still needed or not.
	// Ref: https://github.com/kubernetes-csi/lib-volume-populator/blob/e9508a3a026888d47da5fce7d7ae2856c7810e21/populator-machinery/controller.go#L492
	want := false
	if finalizers := destinationPVC.GetFinalizers(); finalizers != nil {
		for _, f := range finalizers {
			if f == populatorFinalizer {
				want = true
				break
			}
		}
	}

	if want {
		if dataPopulator.Status.State != internalv1alpha1.StatusInProgress {
			// change the status of data-populator
			clone := dataPopulator.DeepCopy()
			clone.Status.State = internalv1alpha1.StatusInProgress
			if err := c.updateDataPopulator(clone); err != nil {
				return fmt.Errorf("error updating status of data populator `%s` in `%s` namespace, error: %s",
					dataPopulator.GetName(), dataPopulator.GetNamespace(), err)
			}
		}

		// Create all the resources needed for the rsync daemon to be up and running
		if err := c.ensureRsyncDaemon(true, dptc, dptc.sourcePVCNamespace); err != nil {
			return err
		}
		return nil
	}

	// Delete all the rsync daemon resources when the finalizer set by the rsync-populator is gone.
	// The finalizer is removed only when the source data has been fully populated into into the desired destination.
	if err := c.ensureRsyncDaemon(false, dptc, dptc.sourcePVCNamespace); err != nil {
		return err
	}

	// Update the data-populator status to mark as completed
	if dataPopulator.Status.State != internalv1alpha1.StatusCompleted {
		clone := dataPopulator.DeepCopy()
		clone.Status.State = internalv1alpha1.StatusCompleted
		if err := c.updateDataPopulator(clone); err != nil {
			return fmt.Errorf("error updating status of data populator `%s` in `%s` namespace, error: %s",
				dataPopulator.GetName(), dataPopulator.GetNamespace(), err)
		}
	}

	return nil
}

// ensureRsyncDaemon ensures the desired state of all the rsync daemon resources
func (c *controller) ensureRsyncDaemon(want bool, dptc *templateConfig, namespace string) error {
	cmTemplate := dptc.getCmTemplate()
	if err := c.ensureConfigMap(want, namespace, &cmTemplate); err != nil {
		return fmt.Errorf("error ensuring(true) configmap `%s` in `%s` namespace, error: %s",
			cmTemplate.GetName(), namespace, err)
	}

	podTemplate := dptc.getPodTemplate()
	if err := c.ensurePod(want, namespace, &podTemplate); err != nil {
		return fmt.Errorf("error ensuring(true) pod `%s` in `%s` namespace, error: %s",
			podTemplate.GetName(), namespace, err)
	}

	svcTemplate := dptc.getSvcTemplate()
	if err := c.ensureService(want, namespace, &svcTemplate); err != nil {
		return fmt.Errorf("error ensuring(true) service `%s` in `%s` namespace, error: %s",
			svcTemplate.GetName(), namespace, err)
	}
	return nil
}

// updateDataPopulator updates a data populator object
func (c *controller) updateDataPopulator(dp *internalv1alpha1.DataPopulator) error {
	dpClone := dp.DeepCopy()
	dpMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dpClone)
	if err != nil {
		return err
	}

	dpUnstruct := &unstructured.Unstructured{
		Object: dpMap,
	}

	_, err = c.dynamicClient.Resource(dpGVR).Namespace(dpClone.GetNamespace()).
		Update(context.TODO(), dpUnstruct, metav1.UpdateOptions{})
	return err
}

/*
if found and not created by the data-populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensurePVC(want bool, namespace string, pvc *corev1.PersistentVolumeClaim) error {
	pvcClone := pvc.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), pvcClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("pvc `%s` found but not created by this operator", obj.GetName())
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}

	if want && !found {
		_, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
			Create(context.TODO(), pvcClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
			Delete(context.TODO(), pvcClone.Name, metav1.DeleteOptions{})
		return err
	}
	return nil
}

/*
if found and not created by the data-populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensurePopulator(want bool, namespace string, populator *internalv1alpha1.RsyncPopulator) error {
	found := true
	populatorClone := populator.DeepCopy()
	populatorMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&populatorClone)
	if err != nil {
		return err
	}
	populatorUnstruct := &unstructured.Unstructured{
		Object: populatorMap,
	}
	obj, err := c.dynamicClient.Resource(rpGVR).Namespace(namespace).
		Get(context.TODO(), populatorClone.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("rsync populator resource `%s` found but not created by this operator", obj.GetName())
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.dynamicClient.Resource(rpGVR).Namespace(namespace).
			Create(context.TODO(), populatorUnstruct, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.dynamicClient.Resource(rpGVR).Namespace(namespace).
			Delete(context.TODO(), populatorClone.GetName(), metav1.DeleteOptions{})
		return err
	}
	return nil
}

/*
if found and not created by the data-populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensureConfigMap(want bool, namespace string, cm *corev1.ConfigMap) error {
	cmClone := cm.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().ConfigMaps(namespace).
		Get(context.TODO(), cmClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("configmap `%s` found but not created by this operator", obj.GetName())
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().ConfigMaps(namespace).
			Create(context.TODO(), cmClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().ConfigMaps(namespace).
			Delete(context.TODO(), cmClone.Name, metav1.DeleteOptions{})
		return err
	}
	return nil
}

/*
if found and not created by the data-populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensurePod(want bool, namespace string, pod *corev1.Pod) error {
	podClone := pod.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().Pods(namespace).
		Get(context.TODO(), podClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("pod `%s` found but not created by this operator", obj.GetName())
	}
	if want && found {
		if obj.Status.Phase == corev1.PodFailed || obj.Status.Phase == corev1.PodSucceeded {
			err = c.ensurePod(false, namespace, podClone)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().Pods(namespace).
			Create(context.TODO(), podClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().Pods(namespace).
			Delete(context.TODO(), podClone.Name, metav1.DeleteOptions{})
		return err
	}
	return nil
}

/*
if found and not created by the data-populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensureService(want bool, namespace string, svc *corev1.Service) error {
	svcClone := svc.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().Services(namespace).
		Get(context.TODO(), svcClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("service `%s` found but not created by this operator", obj.GetName())
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().Services(namespace).
			Create(context.TODO(), svcClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().Services(namespace).
			Delete(context.TODO(), svcClone.Name, metav1.DeleteOptions{})
		return err
	}
	return nil
}
