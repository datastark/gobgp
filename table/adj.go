// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package table

import (
	"github.com/osrg/gobgp/packet"
	"reflect"
)

type AdjRib struct {
	accepted map[bgp.RouteFamily]int
	table    map[bgp.RouteFamily]map[string]*Path
}

func NewAdjRib(rfList []bgp.RouteFamily) *AdjRib {
	table := make(map[bgp.RouteFamily]map[string]*Path)
	for _, rf := range rfList {
		table[rf] = make(map[string]*Path)
	}
	return &AdjRib{
		table:    table,
		accepted: make(map[bgp.RouteFamily]int),
	}
}

func (adj *AdjRib) Update(pathList []*Path) {
	for _, path := range pathList {
		if path == nil {
			continue
		}
		rf := path.GetRouteFamily()
		key := path.getPrefix()
		old, found := adj.table[rf][key]
		if path.IsWithdraw {
			if found {
				delete(adj.table[rf], key)
				if !old.Filtered {
					adj.accepted[rf]--
				}
			}
		} else {
			if found {
				if old.Filtered && !path.Filtered {
					adj.accepted[rf]++
				} else if !old.Filtered && path.Filtered {
					adj.accepted[rf]--
				}
			} else {
				if !path.Filtered {
					adj.accepted[rf]++
				}
			}
			if found && reflect.DeepEqual(old.GetPathAttrs(), path.GetPathAttrs()) {
				path.setTimestamp(old.GetTimestamp())
			}
			adj.table[rf][key] = path
		}
	}
}

func (adj *AdjRib) PathList(rfList []bgp.RouteFamily, accepted bool) []*Path {
	pathList := make([]*Path, 0, adj.Count(rfList))
	for _, rf := range rfList {
		for _, rr := range adj.table[rf] {
			if accepted && rr.Filtered {
				continue
			}
			pathList = append(pathList, rr)
		}
	}
	return pathList
}

func (adj *AdjRib) Count(rfList []bgp.RouteFamily) int {
	count := 0
	for _, rf := range rfList {
		if table, ok := adj.table[rf]; ok {
			count += len(table)
		}
	}
	return count
}

func (adj *AdjRib) Accepted(rfList []bgp.RouteFamily) int {
	count := 0
	for _, rf := range rfList {
		if n, ok := adj.accepted[rf]; ok {
			count += n
		}
	}
	return count
}

func (adj *AdjRib) Drop(rfList []bgp.RouteFamily) {
	for _, rf := range rfList {
		if _, ok := adj.table[rf]; ok {
			adj.table[rf] = make(map[string]*Path)
			adj.accepted[rf] = 0
		}
	}
}