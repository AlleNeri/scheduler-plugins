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
ARG ARCH
ARG GO_BASE_IMAGE=golang
ARG UBUNTU_BASE_IMAGE=ubuntu
FROM $GO_BASE_IMAGE:1.21

WORKDIR /go/src/sigs.k8s.io/scheduler-plugins
COPY . .
ARG ARCH
ARG RELEASE_VERSION
RUN RELEASE_VERSION=${RELEASE_VERSION} make build-scheduler.$ARCH

FROM $UBUNTU_BASE_IMAGE:22.04

RUN apt-get update && \
	apt-get install -y python3 python3-pip python3-venv
RUN ln -s /usr/bin/python3 /usr/bin/python

ENV SCRIPT_PATH=/opt/script

COPY --from=0 /go/src/sigs.k8s.io/scheduler-plugins/bin/kube-scheduler /bin/kube-scheduler
RUN mkdir -p $SCRIPT_PATH
COPY --from=0 /go/src/sigs.k8s.io/scheduler-plugins/script/* $SCRIPT_PATH/

RUN python3 -m venv $SCRIPT_PATH/venv
# RUN $SCRIPT_PATH/venv/bin/pip install --upgrade pip
RUN $SCRIPT_PATH/venv/bin/pip install -r $SCRIPT_PATH/requirements.txt

WORKDIR /bin
CMD ["kube-scheduler"]
