DOCKERHUB_USERNAME ?= abhishek09dh
LATEST_TAG ?= v0.1

.PHONY: crd-gen
crd-gen:
	rm -rf deploy/crds
	controller-gen crd:trivialVersions=true,preserveUnknownFields=false paths="./apis/openebs.io/..." output:crd:artifacts:config=deploy/crds/

.PHONY: rsync-populator-binary
rsync-populator-binary:
	mkdir -p bin
	rm -rf bin/rsync-populator
	CGO_ENABLED=0 go build -o bin/rsync-populator app/populator/rsync/*

.PHONY: data-populator-binary
data-populator-binary:
	mkdir -p bin
	rm -rf bin/data-populator
	CGO_ENABLED=0 go build -o bin/data-populator ./app/populator/data/

.PHONY: rsync-populator-image
rsync-populator-image: rsync-populator-binary
	docker build -t $(DOCKERHUB_USERNAME)/rsync-populator:$(LATEST_TAG) -f buildscripts/populator/rsync/Dockerfile .

.PHONY: data-populator-image
data-populator-image: data-populator-binary
	docker build -t $(DOCKERHUB_USERNAME)/data-populator:$(LATEST_TAG) -f buildscripts/populator/data/Dockerfile .

.PHONY: rsync-daemon-image
rsync-daemon-image:
	docker build -t $(DOCKERHUB_USERNAME)/rsync-daemon:$(LATEST_TAG) -f buildscripts/rsync/server/Dockerfile .

.PHONY: rsync-client-image
rsync-client-image:
	docker build -t $(DOCKERHUB_USERNAME)/rsync-client:$(LATEST_TAG) -f buildscripts/rsync/client/Dockerfile .
