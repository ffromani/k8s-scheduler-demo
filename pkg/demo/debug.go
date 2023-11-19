package demo

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubernetes/pkg/api/v1/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func frameworkResourceToLoggable(req *framework.Resource) []interface{} {
	items := []interface{}{
		"cpu", humanCPU(req.MilliCPU),
		"memory", humanMemory(req.Memory),
	}

	resNames := []string{}
	for resName := range req.ScalarResources {
		resNames = append(resNames, string(resName))
	}
	sort.Strings(resNames)

	for _, resName := range resNames {
		quan := req.ScalarResources[corev1.ResourceName(resName)]
		if resourcehelper.IsHugePageResourceName(corev1.ResourceName(resName)) {
			items = append(items, resName, humanMemory(quan))
		} else {
			items = append(items, resName, strconv.FormatInt(quan, 10))
		}
	}
	return items
}

type humanMemory int64

func (hi humanMemory) String() string {
	return fmt.Sprintf("%d (%s)", hi, humanize.IBytes(uint64(hi)))
}

type humanCPU int64

func (hc humanCPU) String() string {
	return fmt.Sprintf("%d (%d)", hc, hc/1000)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// logic taken from fit.go, changing the return value to use a plain *framework.Resource
// https://github.com/kubernetes/kubernetes/blob/v1.24.0/pkg/scheduler/framework/plugins/noderesources/fit.go#L133-L175

// computePodResourceRequest returns a framework.Resource that covers the largest
// width in each resource dimension. Because init-containers run sequentially, we collect
// the max in each dimension iteratively. In contrast, we sum the resource vectors for
// regular containers since they run simultaneously.
//
// # The resources defined for Overhead should be added to the calculated Resource request sum
//
// Example:
//
// Pod:
//
//	InitContainers
//	  IC1:
//	    CPU: 2
//	    Memory: 1G
//	  IC2:
//	    CPU: 2
//	    Memory: 3G
//	Containers
//	  C1:
//	    CPU: 2
//	    Memory: 1G
//	  C2:
//	    CPU: 1
//	    Memory: 1G
//
// Result: CPU: 3, Memory: 3G
func computePodResourceRequest(pod *corev1.Pod) *framework.Resource {
	result := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	if pod.Spec.Overhead != nil {
		result.Add(pod.Spec.Overhead)
	}
	return result
}

// see again fit.go for the skeleton code. Here we intentionally only log
func checkRequest(podRequest *framework.Resource, nodeInfo *framework.NodeInfo) {
	if podRequest.MilliCPU == 0 && podRequest.Memory == 0 && podRequest.EphemeralStorage == 0 && len(podRequest.ScalarResources) == 0 {
		klog.InfoS("target resource requests none")
		return
	}
	klog.InfoS("target resource requests", frameworkResourceToLoggable(podRequest)...)

	nodeName := nodeInfo.Node().Name // shortcut

	availCPU := (nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU)
	klog.InfoS("node resources", "node", nodeName, "resource", "CPU", "request", humanCPU(podRequest.MilliCPU), "available", humanCPU(availCPU))

	availMemory := (nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory)
	klog.InfoS("node resources", "node", nodeName, "resource", "memory", "request", humanMemory(podRequest.Memory), "available", humanMemory(availMemory))

	for rName, rQuant := range podRequest.ScalarResources {
		availQuant := (nodeInfo.Allocatable.ScalarResources[rName] - nodeInfo.Requested.ScalarResources[rName])
		klog.InfoS("node resources", "node", nodeName, "resource", rName, "request", rQuant, "available", availQuant)
	}
}
