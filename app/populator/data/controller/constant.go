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
