# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

COMMONENVVAR=GOOS=$(shell uname -s | tr A-Z a-z)
BUILDENVVAR=

CONTAINER_REGISTRY?="quay.io/ffromani"
CONTAINER_IMAGE?="k8s-scheduler-demo"

RELEASE_SEQUENTIAL?=01
RELEASE_VERSION?=devel-v0.0.$(shell date +%Y%m%d)$(RELEASE_SEQUENTIAL)

# VERSION is the scheduler's version
#
# The RELEASE_VERSION variable can have one of the following formats:
# v20201009-v0.18.800-46-g939c1c0 - automated build for a commit(not a tag) and also a local build
# v20200521-v0.18.800             - automated build for a tag
VERSION=$(shell echo $(RELEASE_VERSION) | awk -F - '{print $$2}')

.PHONY: all
all: image

.PHONY: sched
sched:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION) -w' -o bin/sched cmd/k8s-scheduler-demo/main.go

.PHONY: image
image:
	podman build -f Dockerfile --build-arg ARCH="amd64" --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" -t $(CONTAINER_REGISTRY)/$(CONTAINER_IMAGE):$(VERSION) .

.PHONY: clean
clean:
	rm -rf ./bin
