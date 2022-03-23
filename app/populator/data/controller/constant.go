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

const (
	GroupOpenebsIO  = "openebs.io"
	VersionV1alpha1 = "v1alpha1"

	RpKind     = "RsyncPopulator"
	RpResource = "rsyncpopulators"

	DpKind     = "DataPopulator"
	DpResource = "datapopulators"

	createdByLabel = "openebs.io/created-by"
	roleLabel      = "openebs.io/role"
	managedByLabel = "openebs.io/managed-by"
	appLabel       = "openebs.io/app"
	roleLabelValue = "rsync-daemon"

	componentName = "data-populator"
	populatorName = "rsync-populator"

	SourcePvcMountPath = "/data"

	nodeNameAnnotation = "volume.kubernetes.io/selected-node"

	populatorFinalizer = "openebs.io/populate-target-protection"

	RsyncNamePrefix  = "rsync-daemon-"
	rsyncServerImage = "abhishek09dh/rsync-daemon:v0.1"
	rsyncUsername    = "openebs-user"
	rsyncPassword    = "openebs-pass"
)
