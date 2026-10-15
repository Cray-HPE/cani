/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package export

import (
	"sort"

	"github.com/google/uuid"
)

// hierarchyOf tells topologicalOrder how to read one kind of parent-linked
// inventory item (locations under locations, FRUs under FRUs).
type hierarchyOf[T any] struct {
	id     func(*T) uuid.UUID
	parent func(*T) uuid.UUID
	name   func(*T) string
}

// topologicalOrder returns the items parents-first so sequential creation can
// resolve every parent's Nautobot ID before its children need it. Roots and
// siblings are ordered by name so the output is stable across runs. Items
// whose parent chain never reaches a root (a cycle or a dangling parent) are
// returned separately, also by name, so the caller can report them.
func topologicalOrder[T any](items map[uuid.UUID]*T, h hierarchyOf[T]) (ordered, unreachable []*T) {
	roots, children := partitionByParent(items, h)
	byName := func(ids []uuid.UUID) {
		sort.Slice(ids, func(i, j int) bool { return h.name(items[ids[i]]) < h.name(items[ids[j]]) })
	}
	byName(roots)
	for _, siblings := range children {
		byName(siblings)
	}

	seen := make(map[uuid.UUID]bool, len(items))
	for queue := roots; len(queue) > 0; queue = queue[1:] {
		id := queue[0]
		seen[id] = true
		ordered = append(ordered, items[id])
		queue = append(queue, children[id]...)
	}

	missing := unvisited(items, seen)
	byName(missing)
	for _, id := range missing {
		unreachable = append(unreachable, items[id])
	}
	return ordered, unreachable
}

// partitionByParent splits the non-nil items into roots and a parent → children
// index.
func partitionByParent[T any](items map[uuid.UUID]*T, h hierarchyOf[T]) (roots []uuid.UUID, children map[uuid.UUID][]uuid.UUID) {
	children = make(map[uuid.UUID][]uuid.UUID)
	for id, item := range items {
		if item == nil {
			continue
		}
		if parent := h.parent(item); parent == uuid.Nil {
			roots = append(roots, id)
		} else {
			children[parent] = append(children[parent], id)
		}
	}
	return roots, children
}

// unvisited lists the non-nil items the walk never reached.
func unvisited[T any](items map[uuid.UUID]*T, seen map[uuid.UUID]bool) []uuid.UUID {
	var missing []uuid.UUID
	for id, item := range items {
		if item != nil && !seen[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

// labels renders "name (id)" for each item, the form the loaders log for
// unreachable entries.
func (h hierarchyOf[T]) labels(items []*T) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, h.name(item)+" ("+h.id(item).String()+")")
	}
	return out
}
