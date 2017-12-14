package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	port   int
	dir    string
	browse bool
)

func init() {
	flag.IntVar(&port, "p", 8080, "deployment port")
	flag.BoolVar(&browse, "open", false, "open browser window")
}

func parseArgs() {
	flag.Parse()
	switch len(flag.Args()) {
	case 0:
		dir = "."
	case 1:
		dir = flag.Arg(0)
	default:
		log.Fatalf("expected directory name, got %v\n",
			strings.Join(flag.Args(), ", "))
	}
	info, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() == false {
		log.Fatalf("%v is not a directory\n", dir)
	}
}

func Open() {
	endpoint := fmt.Sprintf("http://localhost:%d", port)
	for {
		_, err := http.Get(endpoint)
		if err != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		break
	}
	// based on http://stackoverflow.com/a/39324149/7357996
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, endpoint)
	exec.Command(cmd, args...).Start()
}

// NoCache sets headers that forbid caching
// see:
// * https://stackoverflow.com/questions/33880343/go-webserver-dont-cache-files-using-timestamp
// * https://stackoverflow.com/questions/49547/how-to-control-web-page-caching-across-all-browsers
func NoCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range map[string]string{
			"Expires":       "0",
			"Cache-Control": "no-store, must-revalidate",
		} {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	parseArgs()
	port := fmt.Sprintf(":%d", port)
	if browse {
		go Open()
	}
	fmt.Printf("Running server at 127.0.0.1%s...\n", port)
	handler := http.FileServer(http.Dir(dir))
	handler = NoCache(handler)
	log.Fatal(http.ListenAndServe(port, handler))
}
