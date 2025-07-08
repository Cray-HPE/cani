/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package extract

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var debug bool

func Extract(cmd *cobra.Command, args []string) (bomQueues Queues, err error) {
	if cmd.Root().PersistentFlags().Changed("debug") {
		debug = true
		log.Printf("Debug mode enabled")
	}
	if cmd.Flags().Changed("bom") {
		bomQueues, err = extract(cmd, args)
		if err != nil {
			return bomQueues, err
		}
	} else {
		return bomQueues, fmt.Errorf("no BOMs provided")
	}

	return bomQueues, nil
}

// validateBoms loads the bom files and unmarshals them into the Ngsm.Boms map
func extract(cmd *cobra.Command, args []string) (bomQueues Queues, err error) {
	// multiple bom flags can be passed, so iterate over them
	// this allows more than one bom to be parsed at a time
	// a customer may have one bom from the initial order, which may later be
	// supplemented with additional orders, so this can combine them all
	bomQueues = make(Queues)

	boms, _ := cmd.Flags().GetStringArray("bom")
	for _, bom := range boms {
		log.Printf("Validating BOM before import: %s", bom)
		// get all the rows from the current bom
		rows, err := getRowsFromBom(bom)
		if err != nil {
			return bomQueues, err
		}

		// make a Queue for each bom.  these will be combined into a single Queues
		q, err := newQueues(bom)
		if err != nil {
			return bomQueues, err
		}

		// add the rows to the Queue
		for n, row := range rows {
			// parse each row and sanitize the cells
			r, err := row.Sanitize()
			if err != nil {
				return bomQueues, err
			}

			// set the row number, +1 for 0 based index, +1 for header row
			// TODO: there is likely a better way to do this
			row.SetRow(n + 2)

			// add the row to the Queue if it is valid
			err = q.QueuesRow(r)
			if err != nil {
				return bomQueues, err
			}
		}

		// add the Queue to the Queues
		bomQueues[bom] = q

		log.Println()
		log.Printf("  %d racks Queued for transformation", len(q.RacksToCreate))
		log.Printf("  %d devices Queued for transformation", len(q.DevicesToCreate))
		log.Println()
	}

	return bomQueues, nil

}
