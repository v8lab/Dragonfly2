/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//go:generate mockgen -destination resource_mock.go -source resource.go -package resource

package resource

import (
	"google.golang.org/grpc"

	"d7y.io/dragonfly/v2/pkg/gc"
	"d7y.io/dragonfly/v2/scheduler/config"
)

type Resource interface {
	// SeedPeer interface.
	SeedPeer() SeedPeer

	// Host manager interface.
	HostManager() HostManager

	// Peer manager interface.
	PeerManager() PeerManager

	// Task manager interface.
	TaskManager() TaskManager

	// Stop resource serivce.
	Stop() error
}

type resource struct {
	// seedPeer interface.
	seedPeer SeedPeer

	// Seed peer client interface.
	seedPeerClient SeedPeerClient

	// Host manager interface.
	hostManager HostManager

	// Peer manager interface.
	peerManager PeerManager

	// Task manager interface.
	taskManager TaskManager
}

func New(cfg *config.Config, gc gc.GC, dynconfig config.DynconfigInterface, opts ...grpc.DialOption) (Resource, error) {
	resource := &resource{}

	// Initialize host manager interface.
	hostManager, err := newHostManager(&cfg.Scheduler.GC, gc)
	if err != nil {
		return nil, err
	}
	resource.hostManager = hostManager

	// Initialize task manager interface.
	taskManager, err := newTaskManager(&cfg.Scheduler.GC, gc)
	if err != nil {
		return nil, err
	}
	resource.taskManager = taskManager

	// Initialize peer manager interface.
	peerManager, err := newPeerManager(&cfg.Scheduler.GC, gc)
	if err != nil {
		return nil, err
	}
	resource.peerManager = peerManager

	// Initialize seed peer interface.
	if cfg.SeedPeer.Enable {
		client, err := newSeedPeerClient(dynconfig, hostManager, opts...)
		if err != nil {
			return nil, err
		}

		resource.seedPeerClient = client
		resource.seedPeer = newSeedPeer(client, peerManager, hostManager)
	}

	return resource, nil
}

func (r *resource) SeedPeer() SeedPeer {
	return r.seedPeer
}

func (r *resource) HostManager() HostManager {
	return r.hostManager
}

func (r *resource) TaskManager() TaskManager {
	return r.taskManager
}

func (r *resource) PeerManager() PeerManager {
	return r.peerManager
}

func (r *resource) Stop() error {
	return r.seedPeerClient.Close()
}
