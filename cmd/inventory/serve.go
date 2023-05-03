/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package inventory

import "net/http"

// ServeHTTP handles the inventory page
func (ih System) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert inventory here"))
}

// // ServeHTTP handles the CSI config page
// func (csih CsiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Insert CSI config here"))
// }

// // ServeHTTP handles the CANU config page
// func (canuh CanuConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Insert CANU config here"))
// }

// ServeHTTP handles the SLS config page
func (slsh SlsConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert SLS config here"))
}
