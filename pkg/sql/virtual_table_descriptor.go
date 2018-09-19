// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package sql

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
)

// virtualTableGenerator is the function signature for the virtualTableNode
// `next` property. Each time the virtualTableGenerator function is called, it
// returns a tree.Datums corresponding to the next row of the virtual schema
// table. If there is no next row (end of table is reached), then return (nil,
// nil). If there is an error, then return (nil, error).
type virtualTableDescriptorGenerator func() (tree.Datums, error)

// virtualTableDescriptorNode is a planNode that constructs its rows by repeatedly
// invoking a virtualTableDescriptorGenerator function.
type virtualTableDescriptorNode struct {
	columns    sqlbase.ResultColumns
	next       virtualTableDescriptorGenerator
	currentRow tree.Datums
}

func (p *planner) newContainerVirtualTableDescriptorNode(
	columns sqlbase.ResultColumns, capacity int, next virtualTableDescriptorGenerator,
) *virtualTableDescriptorNode {
	return &virtualTableDescriptorNode{
		columns: columns,
		next:    next,
	}
}

func (n *virtualTableDescriptorNode) Next(params runParams) (bool, error) {
	row, err := n.next()
	if err != nil {
		return false, err
	}
	n.currentRow = row
	return row != nil, nil
}

func (n *virtualTableDescriptorNode) Values() tree.Datums {
	return n.currentRow
}

func (n *virtualTableDescriptorNode) Close(ctx context.Context) {
}
