package main

import (
	"flag"
	"os"

	populator_machinery "github.com/kubernetes-csi/lib-volume-populator/populator-machinery"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	internalv1alpha1 "github.com/Ab-hishek/data-populator/apis/openebs.io/v1alpha1"
)

const (
	prefix     = "openebs.io"
	mountPath  = "/mnt"
	devicePath = "/dev/block"

	groupName  = "openebs.io"
	apiVersion = "v1alpha1"
	kind       = "RsyncPopulator"
	resource   = "rsyncpopulators"
)

var (
	gk  = schema.GroupKind{Group: groupName, Kind: kind}
	gvr = schema.GroupVersionResource{Group: groupName, Version: apiVersion, Resource: resource}
)

func main() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		panic(err)
	}

	var (
		imageName string
	)
	flag.StringVar(&imageName, "image-name", "", "Image to use for populating")
	flag.Parse()

	namespace := os.Getenv("POD_NAMESPACE")

	populator_machinery.RunController("", "", imageName,
		namespace, prefix, gk, gvr, mountPath, devicePath, getPopulatorArgs)
}

func getPopulatorArgs(rawBlock bool, u *unstructured.Unstructured) ([]string, error) {
	populator := internalv1alpha1.RsyncPopulator{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), &populator)
	if err != nil {
		return nil, err
	}

	args := []string{
		"bash",
		"-c",
		"export RSYNC_PASSWORD=" + populator.Spec.Password + "; rsync -rv rsync://" + populator.Spec.Username + "@" + populator.Spec.URL + populator.Spec.Path + " " + mountPath,
	}
	return args, nil
}
