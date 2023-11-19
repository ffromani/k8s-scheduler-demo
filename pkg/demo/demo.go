/*
 * Copyright 2022 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package demo

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type Demo struct{}

var _ framework.FilterPlugin = &Demo{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name string = "Demo"
)

// Name returns name of the plugin. It is used in logs, etc.
func (dm *Demo) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.V(6).InfoS("Creating new Demo plugin")
	return &Demo{}, nil
}

func (dm *Demo) EventsToRegister() []framework.ClusterEvent {
	// this can actually be empty - this plugin never fails, but we keep the same
	// (simple and safe) events noderesourcesfit registered
	return []framework.ClusterEvent{
		{Resource: framework.Pod, ActionType: framework.Delete},
		{Resource: framework.Node, ActionType: framework.Add | framework.UpdateNodeAllocatable},
	}
}

func (dm *Demo) Filter(ctx context.Context, cycleState *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	node := nodeInfo.Node()
	if node == nil {
		// should never happen
		return framework.NewStatus(framework.Error, "node not found")
	}

	checkRequest(computePodResourceRequest(pod), nodeInfo)
	return nil // must never fail
}
