/*
Copyright 2025 The Kubernetes Authors.

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

package lifecycle

import (
	"k8s.io/apimachinery/pkg/types"
)

// CacheInvalidator provides methods to invalidate the NodeInfo cache when
// relevant events occur in the kubelet.
type CacheInvalidator interface {
	// OnPodAdded invalidates the cache when a pod is added
	OnPodAdded(podUID types.UID)

	// OnPodRemoved invalidates the cache when a pod is removed
	OnPodRemoved(podUID types.UID)

	// OnPodAllocationChanged invalidates the cache when a pod's allocation changes (resized)
	OnPodAllocationChanged(podUID types.UID)

	// OnPodStatusChanged invalidates the cache when a pod's status resources change (resize is actuated)
	OnPodStatusChanged(podUID types.UID)

	// OnPodBecameInactive invalidates the cache when a pod becomes inactive
	OnPodBecameInactive(podUID types.UID)

	// OnNodeUpdated invalidates the cache when the node is updated
	OnNodeUpdated()
}

// cacheInvalidator implements CacheInvalidator by delegating to a NodeInfoCache
type cacheInvalidator struct {
	nodeInfoCache NodeInfoCache
}

// NewCacheInvalidator creates a new CacheInvalidator that invalidates the given NodeInfoCache
func NewCacheInvalidator(nodeInfoCache NodeInfoCache) CacheInvalidator {
	return &cacheInvalidator{
		nodeInfoCache: nodeInfoCache,
	}
}

// OnPodAdded invalidates the cache when a pod is added
func (ci *cacheInvalidator) OnPodAdded(podUID types.UID) {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to pod addition", "podUID", podUID)
	ci.nodeInfoCache.InvalidatePod(podUID)
}

// OnPodRemoved invalidates the cache when a pod is removed
func (ci *cacheInvalidator) OnPodRemoved(podUID types.UID) {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to pod removal", "podUID", podUID)
	ci.nodeInfoCache.InvalidatePod(podUID)
}

// OnPodAllocationChanged invalidates the cache when a pod's allocation changes (resized)
func (ci *cacheInvalidator) OnPodAllocationChanged(podUID types.UID) {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to pod allocation change", "podUID", podUID)
	ci.nodeInfoCache.InvalidatePod(podUID)
}

// OnPodStatusChanged invalidates the cache when a pod's status resources change (resize is actuated)
func (ci *cacheInvalidator) OnPodStatusChanged(podUID types.UID) {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to pod status change", "podUID", podUID)
	ci.nodeInfoCache.InvalidatePod(podUID)
}

// OnPodBecameInactive invalidates the cache when a pod becomes inactive
func (ci *cacheInvalidator) OnPodBecameInactive(podUID types.UID) {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to pod becoming inactive", "podUID", podUID)
	ci.nodeInfoCache.InvalidatePod(podUID)
}

// OnNodeUpdated invalidates the cache when the node is updated
func (ci *cacheInvalidator) OnNodeUpdated() {
	klog.V(5).InfoS("Invalidating NodeInfo cache due to node update")
	ci.nodeInfoCache.InvalidateNode()
}
