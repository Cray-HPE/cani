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
package csminv

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// serveCmd represents the base command when called without any subcommands
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Listen for API requests.",
	Long:  `Listen for API requests.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		serve()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// serveCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.csminv.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB

var (
	webPort    = 3000
	webHost    = "127.0.0.1"
	listenPort = fmt.Sprintf(":%d", webPort)
	webPath    = fmt.Sprintf("http://%s%s", webHost, listenPort)
)

// serve listens for API requests.
func serve() {
	// Use the http.NewServeMux() function to create an empty servemux.
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(indexHandler))
	mux.HandleFunc("/upload", uploadHandler)

	// Next we use the mux.Handle() function to register this with our new
	// servemux, so it acts as the handler for all incoming requests with the URL
	// path /foo.
	csih := CsiConfig{}
	mux.Handle("/extract/csi", csih)

	canuh := CanuConfig{}
	mux.Handle("/extract/canu", canuh)

	slsh := SlsConfig{}
	mux.Handle("/extract/sls", slsh)

	ih := Inventory{}
	mux.Handle("/inventory", ih)

	log.Printf("Listening %s...\n", webPath)

	// Then we create a new server and start listening for incoming requests
	// with the http.ListenAndServe() function, passing in our servemux for it to
	// match requests against as the second parameter.
	http.ListenAndServe(listenPort, mux)
}

// indexHandler handles the index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>File upload demo</title>
  </head>
  <body>
    <form
      id="form"
      enctype="multipart/form-data"
      action="/upload"
      method="POST"
    >
      <input class="input file-input" type="file" name="file" multiple />
      <button class="button" type="submit">Submit</button>
    </form>
  </body>
</html>`))
}

// ServeHTTP handles the inventory page
func (ih Inventory) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert inventory here"))
}

// ServeHTTP handles the CSI config page
func (csih CsiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert CSI config here"))
}

// ServeHTTP handles the CANU config page
func (canuh CanuConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert CANU config here"))
}

// ServeHTTP handles the SLS config page
func (slsh SlsConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Insert SLS config here"))
}

// uploadHandler handles the upload page
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.ContentLength > MAX_UPLOAD_SIZE {
		http.Error(w, "The uploaded image is too big. Please use an image less than 1MB in size", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "1MB or less", http.StatusBadRequest)
		return
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MAX_UPLOAD_SIZE {
			http.Error(w, fmt.Sprintf("The uploaded file is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		// Create a buffer for the file contents
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Seek the start of the file
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the file
		f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()
		log.Println("Upload successful:", fileHeader.Filename)
	}
	fmt.Fprintf(w, "Upload successful")
}

// Progress is used to track the progress of a file upload.
// It implements the io.Writer interface so it can be passed
// to an io.TeeReader()
type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface.
// Instead of writing somewhere, it simply aggregates
// the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
// each time Write is called
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}

	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}
