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
package imprt

import (
	"context"
	"fmt"
)

// pageSize is the number of items to request per API page.
const pageSize = 100

func intPtr(v int) *int { return &v }

// pageResult holds one page of results from a Nautobot list endpoint.
type pageResult[T any] struct {
	Items []T
	Done  bool // true when there are no more pages
}

// fetchPageFunc fetches a single page at the given offset.
type fetchPageFunc[T any] func(ctx context.Context, offset int) (pageResult[T], error)

// paginate fetches all pages from a Nautobot list endpoint.
func paginate[T any](ctx context.Context, noun string, fetch fetchPageFunc[T]) ([]T, error) {
	var all []T
	offset := 0
	for {
		page, err := fetch(ctx, offset)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", noun, err)
		}
		if len(page.Items) == 0 {
			break
		}
		all = append(all, page.Items...)
		if page.Done {
			break
		}
		offset += pageSize
	}
	return all, nil
}
