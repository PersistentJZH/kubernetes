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
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfoCache provides a cached NodeInfo for the kubelet to avoid repeated
// heavy NodeInfo construction operations during admission requests and resize retries.
type NodeInfoCache interface {
	// GetNodeInfo returns the cached NodeInfo for the given node and pods.
	// If the cache is invalid or missing, it will construct a new NodeInfo.
	GetNodeInfo(node *v1.Node, pods []*v1.Pod) *framework.NodeInfo

	// InvalidatePod marks the cache as invalid when a pod is added, removed, or modified.
	InvalidatePod(podUID types.UID)

	// InvalidateNode marks the cache as invalid when the node is updated.
	InvalidateNode()

	// InvalidateAll marks the entire cache as invalid.
	InvalidateAll()
}

// nodeInfoCache implements NodeInfoCache with thread-safe caching.
type nodeInfoCache struct {
	mu sync.RWMutex

	// Cache state
	cachedNodeInfo *framework.NodeInfo
	cachedNodeUID  types.UID
	cachedPodUIDs  map[types.UID]bool
	lastUpdate     time.Time

	// Cache invalidation settings
	maxAge time.Duration
}

// NewNodeInfoCache creates a new NodeInfoCache instance.
func NewNodeInfoCache(maxAge time.Duration) NodeInfoCache {
	return &nodeInfoCache{
		cachedPodUIDs: make(map[types.UID]bool),
		maxAge:        maxAge,
	}
}

// GetNodeInfo returns the cached NodeInfo if it's valid, otherwise constructs a new one.
func (c *nodeInfoCache) GetNodeInfo(node *v1.Node, pods []*v1.Pod) *framework.NodeInfo {
	c.mu.RLock()
	if c.isValid(node, pods) {
		defer c.mu.RUnlock()
		return c.cachedNodeInfo.Snapshot()
	}
	c.mu.RUnlock()

	// Cache miss or invalid, need to rebuild
	return c.rebuildCache(node, pods)
}

// isValid checks if the cached NodeInfo is still valid for the given node and pods.
func (c *nodeInfoCache) isValid(node *v1.Node, pods []*v1.Pod) bool {
	if c.cachedNodeInfo == nil {
		return false
	}

	// Check if node has changed
	if c.cachedNodeUID != node.UID {
		return false
	}

	// Check if cache is too old
	if time.Since(c.lastUpdate) > c.maxAge {
		return false
	}

	// Check if pod set has changed
	if len(c.cachedPodUIDs) != len(pods) {
		return false
	}

	// Check if any pods have changed
	for _, pod := range pods {
		if !c.cachedPodUIDs[pod.UID] {
			return false
		}
	}

	return true
}

// rebuildCache constructs a new NodeInfo and updates the cache.
func (c *nodeInfoCache) rebuildCache(node *v1.Node, pods []*v1.Pod) *framework.NodeInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check validity under write lock
	if c.isValid(node, pods) {
		return c.cachedNodeInfo.Snapshot()
	}

	// Construct new NodeInfo
	nodeInfo := framework.NewNodeInfo(pods...)
	nodeInfo.SetNode(node)

	// Update cache
	c.cachedNodeInfo = nodeInfo
	c.cachedNodeUID = node.UID
	c.cachedPodUIDs = make(map[types.UID]bool, len(pods))
	for _, pod := range pods {
		c.cachedPodUIDs[pod.UID] = true
	}
	c.lastUpdate = time.Now()

	klog.V(5).InfoS("NodeInfo cache rebuilt", "node", klog.KObj(node), "pods", len(pods))
	return nodeInfo.Snapshot()
}

// InvalidatePod marks the cache as invalid when a pod is added, removed, or modified.
func (c *nodeInfoCache) InvalidatePod(podUID types.UID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedNodeInfo != nil {
		delete(c.cachedPodUIDs, podUID)
		klog.V(5).InfoS("NodeInfo cache invalidated for pod", "podUID", podUID)
	}
}

// InvalidateNode marks the cache as invalid when the node is updated.
func (c *nodeInfoCache) InvalidateNode() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedNodeInfo != nil {
		c.cachedNodeInfo = nil
		c.cachedNodeUID = ""
		c.cachedPodUIDs = make(map[types.UID]bool)
		klog.V(5).InfoS("NodeInfo cache invalidated for node update")
	}
}

// InvalidateAll marks the entire cache as invalid.
func (c *nodeInfoCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cachedNodeInfo = nil
	c.cachedNodeUID = ""
	c.cachedPodUIDs = make(map[types.UID]bool)
	klog.V(5).InfoS("NodeInfo cache completely invalidated")
}
