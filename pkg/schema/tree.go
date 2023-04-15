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

package schema

type GroupTree struct {
	groups []Group
}

func NewGroupTree(groups []Group) *GroupTree {
	tree := &GroupTree{groups: groups}
	return tree
}

func (t *GroupTree) GetParent(myId string) []Group {
	newArr := make([]Group, 0)
	pid := ""
	for _, g := range t.groups {
		if g.Identity == myId {
			pid = g.ParentId
			break
		}
	}
	if len(pid) > 0 {
		for _, g := range t.groups {
			if g.Identity == pid {
				newArr = append(newArr, g)
				break
			}
		}
	}
	return newArr
}

func (t *GroupTree) GetParents(myId string, withSelf bool) []Group {
	newArr := make([]Group, 0)
	pid := ""
	for _, g := range t.groups {
		if g.Identity == myId {
			if withSelf {
				newArr = append(newArr, g)
			}
			pid = g.ParentId
			break
		}
	}
	if len(pid) > 0 {
		recursiveResult := t.GetParents(pid, true)
		newArr = append(newArr, recursiveResult...)
	}
	return newArr
}

func (t *GroupTree) GetParentsIds(myId string, withSelf bool) []string {
	parents := t.GetParents(myId, withSelf)
	parentIds := make([]string, 0, len(parents))
	for _, parent := range parents {
		parentIds = append(parentIds, parent.Identity)
	}
	return parentIds
}
