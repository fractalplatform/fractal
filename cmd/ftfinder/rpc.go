package main

import (
	"github.com/fractalplatform/fractal/p2p/enode"
)

type Backend interface {
	SeedNodes() []*enode.Node
}

type FinderRPC struct {
	b Backend
}

// SeedNodes returns all seed nodes.
func (fr *FinderRPC) SeedNodes() []string {
	nodes := fr.b.SeedNodes()
	ns := make([]string, len(nodes))
	for i, node := range nodes {
		ns[i] = node.String()
	}
	return ns
}
