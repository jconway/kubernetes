/*
Copyright 2015 The Kubernetes Authors.

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

package qos

import (
	v1 "k8s.io/api/core/v1"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

const (
	// KubeletOOMScoreAdj is the OOM score adjustment for Kubelet
	KubeletOOMScoreAdj int = -999
	// KubeProxyOOMScoreAdj is the OOM score adjustment for kube-proxy
	KubeProxyOOMScoreAdj  int = -999
	guaranteedOOMScoreAdj int = -998
	besteffortOOMScoreAdj int = 1000
)

// GetContainerOOMScoreAdjust returns the amount by which the OOM score of all processes in the
// container should be adjusted.
// The OOM score of a process is the percentage of memory it consumes
// multiplied by 10 (barring exceptional cases) + a configurable quantity which is between -1000
// and 1000. Containers with higher OOM scores are killed if the system runs out of memory.
// See https://lwn.net/Articles/391222/ for more information.
func GetContainerOOMScoreAdjust(pod *v1.Pod, container *v1.Container, memoryCapacity int64) int {
	if types.IsCriticalPod(pod) {
		// Critical pods should be the last to get killed.
		return guaranteedOOMScoreAdj
	}

	switch v1qos.GetPodQOS(pod) {
	case v1.PodQOSGuaranteed:
		// Guaranteed containers should be the last to get killed.
		return guaranteedOOMScoreAdj
	case v1.PodQOSBestEffort:
		return besteffortOOMScoreAdj
	}

	// Burstable containers are a middle tier, between Guaranteed and Best-Effort.
	// The formula below is a heuristic. A container requesting for 10% of a system's
	// memory will have an OOM score adjust of -900.
	memoryRequest := container.Resources.Requests.Memory().Value()
	oomScoreAdjust := -1000 + (1000*memoryRequest)/memoryCapacity

	return int(oomScoreAdjust)
}
