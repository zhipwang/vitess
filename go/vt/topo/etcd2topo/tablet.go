/*
Copyright 2017 Google Inc.

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

package etcd2topo

import (
	"path"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/topoproto"

	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// tabletPathForAlias converts a tablet alias to the node path.
func tabletPathForAlias(alias *topodatapb.TabletAlias) string {
	return path.Join(tabletsPath, topoproto.TabletAliasString(alias), topo.TabletFile)
}

// CreateTablet implements topo.Server.
func (s *Server) CreateTablet(ctx context.Context, tablet *topodatapb.Tablet) error {
	data, err := proto.Marshal(tablet)
	if err != nil {
		return err
	}

	nodePath := tabletPathForAlias(tablet.Alias)
	_, err = s.Create(ctx, tablet.Alias.Cell, nodePath, data)
	return err
}

// UpdateTablet implements topo.Server.
func (s *Server) UpdateTablet(ctx context.Context, tablet *topodatapb.Tablet, existingVersion int64) (int64, error) {
	data, err := proto.Marshal(tablet)
	if err != nil {
		return 0, err
	}

	nodePath := tabletPathForAlias(tablet.Alias)
	version, err := s.Update(ctx, tablet.Alias.Cell, nodePath, data, VersionFromInt(existingVersion))
	if err != nil {
		return 0, err
	}
	return int64(version.(EtcdVersion)), nil
}

// DeleteTablet implements topo.Server.
func (s *Server) DeleteTablet(ctx context.Context, alias *topodatapb.TabletAlias) error {
	nodePath := tabletPathForAlias(alias)
	return s.Delete(ctx, alias.Cell, nodePath, nil)
}

// GetTablet implements topo.Server.
func (s *Server) GetTablet(ctx context.Context, alias *topodatapb.TabletAlias) (*topodatapb.Tablet, int64, error) {
	nodePath := tabletPathForAlias(alias)
	data, version, err := s.Get(ctx, alias.Cell, nodePath)
	if err != nil {
		return nil, 0, err
	}

	tablet := &topodatapb.Tablet{}
	if err := proto.Unmarshal(data, tablet); err != nil {
		return nil, 0, err
	}
	return tablet, int64(version.(EtcdVersion)), nil
}

// GetTabletsByCell implements topo.Server.
func (s *Server) GetTabletsByCell(ctx context.Context, cell string) ([]*topodatapb.TabletAlias, error) {
	// Check if the cell exists first.
	if _, err := s.clientForCell(ctx, cell); err != nil {
		return nil, err
	}

	children, err := s.ListDir(ctx, cell, tabletsPath)
	if err != nil {
		if err == topo.ErrNoNode {
			return nil, nil
		}
		return nil, err
	}

	result := make([]*topodatapb.TabletAlias, len(children))
	for i, child := range children {
		result[i], err = topoproto.ParseTabletAlias(child)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
