// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package rpcapi

import (
	"context"
	"fmt"

	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/rpc"
)

// PrivateP2pAPI offers and API for p2p networking.
type PrivateP2pAPI struct {
	b Backend
}

type notifyEvent struct {
	Count int
	Add   bool
	URL   *string
}

// NewPrivateP2pAPI creates a new p2p service that gives information about p2p networking.
func NewPrivateP2pAPI(b Backend) *PrivateP2pAPI {
	return &PrivateP2pAPI{b}
}

// PeerEvents creates an RPC subscription which receives peer events from the
// node's p2p.Server
func (api *PrivateP2pAPI) PeerEvents(ctx context.Context) (*rpc.Subscription, error) {
	// Create the subscription
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		ch := make(chan *router.Event)
		var pstring *string
		subNew := router.Subscribe(nil, ch, router.NewPeerNotify, pstring)
		subDel := router.Subscribe(nil, ch, router.DelPeerNotify, pstring)
		defer subNew.Unsubscribe()
		defer subDel.Unsubscribe()

		for {
			select {
			case e := <-ch:
				notifier.Notify(rpcSub.ID, &notifyEvent{
					Count: api.b.PeerCount(),
					Add:   e.Typecode == router.NewPeerNotify,
					URL:   e.Data.(*string),
				})
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// AddPeer requests connecting to a remote node, and also maintaining the new
// connection at all times, even reconnecting if it is lost.
func (api *PrivateP2pAPI) AddPeer(url string) (bool, error) {
	if err := api.b.AddPeer(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// RemovePeer disconnects from a remote node if the connection exists
func (api *PrivateP2pAPI) RemovePeer(url string) (bool, error) {
	if err := api.b.RemovePeer(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// AddTrustedPeer allows a remote node to always connect, even if slots are full
func (api *PrivateP2pAPI) AddTrustedPeer(url string) (bool, error) {
	if err := api.b.AddTrustedPeer(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// RemoveTrustedPeer removes a remote node from the trusted peer set, but it
// does not disconnect it automatically.
func (api *PrivateP2pAPI) RemoveTrustedPeer(url string) (bool, error) {
	if err := api.b.RemoveTrustedPeer(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// SeedNodes returns all seed nodes.
func (api *PrivateP2pAPI) SeedNodes() []string {
	return api.b.SeedNodes()
}

// PeerCount return number of connected peers
func (api *PrivateP2pAPI) PeerCount() int {
	return api.b.PeerCount()
}

// Peers return connected peers
func (api *PrivateP2pAPI) Peers() []string {
	return api.b.Peers()
}

// BadNodesCount returns the number of bad nodes.
func (api *PrivateP2pAPI) BadNodesCount() int {
	return api.b.BadNodesCount()
}

// BadNodes returns all bad nodes.
func (api *PrivateP2pAPI) BadNodes() []string {
	return api.b.BadNodes()
}

// AddBadNode add a bad node
func (api *PrivateP2pAPI) AddBadNode(url string) (bool, error) {
	if err := api.b.AddBadNode(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// RemoveBadNode remove a bad node
func (api *PrivateP2pAPI) RemoveBadNode(url string) (bool, error) {
	if err := api.b.RemoveBadNode(url); err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	return true, nil
}

// SelfNode return self enode url
func (api *PrivateP2pAPI) SelfNode() string {
	return api.b.SelfNode()
}
