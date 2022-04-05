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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	internalv1alpha1 "github.com/Ab-hishek/data-populator/apis/openebs.io/v1alpha1"
)

type templateConfig struct {
	sourcePVCName      string
	sourcePVCNamespace string
	destinationPVCSpec corev1.PersistentVolumeClaimSpec
	imageName          string
	rsyncPassword      string
	rsyncUsername      string
}

func templateFromDataPopulator(cr internalv1alpha1.DataPopulator) (*templateConfig, error) {
	tc := &templateConfig{
		sourcePVCName:      cr.Spec.SourcePVC,
		sourcePVCNamespace: cr.Spec.SourcePVCNamespace,
		destinationPVCSpec: cr.Spec.DestinationPVC,
		imageName:          rsyncServerImage,
		rsyncUsername:      rsyncUsername,
		rsyncPassword:      rsyncPassword,
	}
	return tc, nil
}

// getDestinationPVCTemplate returns destination pvc object
// To the destination pvc object add the following:
// 1. add created by label
// 2. add datasource so that it works with rsync populator
func (tc *templateConfig) getDestinationPVCTemplate() corev1.PersistentVolumeClaim {
	destinationPvc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: tc.sourcePVCName + "-populated",
			Labels: map[string]string{
				createdByLabel: componentName,
			},
		},
		Spec: tc.destinationPVCSpec,
	}

	// Set Populator data-source details
	destinationPvc.Spec.DataSourceRef = &corev1.TypedLocalObjectReference{
		Kind: RpKind,
		APIGroup: func() *string {
			name := GroupOpenebsIO
			return &name
		}(),
		Name: RsyncNamePrefix + tc.sourcePVCName,
	}

	return destinationPvc
}

func (tc *templateConfig) getRsyncPopulatorTemplate() internalv1alpha1.RsyncPopulator {
	populator := internalv1alpha1.RsyncPopulator{
		TypeMeta: metav1.TypeMeta{
			Kind:       RpKind,
			APIVersion: GroupOpenebsIO + "/" + VersionV1alpha1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: RsyncNamePrefix + tc.sourcePVCName,
			Labels: map[string]string{
				roleLabel:      populatorName,
				createdByLabel: componentName,
				managedByLabel: componentName,
			},
		},
		Spec: internalv1alpha1.RsyncPopulatorSpec{
			Username: tc.rsyncUsername,
			Password: tc.rsyncPassword,
			Path:     SourcePvcMountPath,
			URL:      RsyncNamePrefix + tc.sourcePVCName + "." + tc.sourcePVCNamespace + ":873",
		},
	}
	return populator
}

func (tc *templateConfig) getCmTemplate() corev1.ConfigMap {
	var rsyncdconfig = `
# /etc/rsyncd.conf

# Minimal configuration file for rsync daemon
# See rsync(1) and rsyncd.conf(5) man pages for help

# This line is required by the /etc/init.d/rsyncd script
pid file = /var/run/rsyncd.pid

uid = 0
gid = 0
use chroot = yes
reverse lookup = no
[data]
    hosts deny = *
    hosts allow = 0.0.0.0/0
    read only = false
    path = ` + SourcePvcMountPath + `
    auth users = , ` + tc.rsyncUsername + `:rw
    secrets file = /etc/rsyncd.secrets
    timeout = 600
    transfer logging = true
`
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: RsyncNamePrefix + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				roleLabel:      roleLabelValue,
			},
		},
		Data: map[string]string{
			"rsyncd.conf": rsyncdconfig,
		},
	}
	return cm
}

func (tc *templateConfig) getPodTemplate() corev1.Pod {
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: RsyncNamePrefix + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				appLabel:       RsyncNamePrefix + tc.sourcePVCName,
				roleLabel:      roleLabelValue,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "rsync-daemon",
					Image:           tc.imageName,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  "RSYNC_PASSWORD",
							Value: tc.rsyncPassword,
						},
						{
							Name:  "RSYNC_USERNAME",
							Value: tc.rsyncUsername,
						},
					},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 873,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: SourcePvcMountPath,
						},
						{
							Name:      "config",
							MountPath: "/etc/rsyncd.conf",
							SubPath:   "rsyncd.conf",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: tc.sourcePVCName,
						},
					},
				},
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: RsyncNamePrefix + tc.sourcePVCName,
							},
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	return pod
}

func (tc *templateConfig) getSvcTemplate() corev1.Service {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: RsyncNamePrefix + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				roleLabel:      roleLabelValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "rsync-daemon",
					Port:     873,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				appLabel:  RsyncNamePrefix + tc.sourcePVCName,
				roleLabel: roleLabelValue,
			},
		},
	}
	return svc
}
