// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"github.com/nautes-labs/base-operator/pkg/schema"
)

type groupMemory struct {
	data []*schema.Group
}

func NewGroupMemery(groups []*schema.Group) *groupMemory {
	g := &groupMemory{
		data: groups,
	}
	return g
}

func (g *groupMemory) Store(group *schema.Group) {
	for i, item := range g.data {
		if item.Identity == group.Identity {
			g.data[i] = group
			return
		}
	}
	g.data = append(g.data, group)
}

func (g *groupMemory) Get(identity string) *schema.Group {
	for _, item := range g.data {
		if item.Identity == identity {
			return item
		}
	}
	return nil
}

func (g *groupMemory) List() []*schema.Group {
	return g.data
}
