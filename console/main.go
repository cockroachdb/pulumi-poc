package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const pulumiDir = "/Users/josh/go/src/github.com/cockroachdb/pulumi-poc/template"

func readHandler(w http.ResponseWriter, r *http.Request) {
	// Haven't figured out how to get direct access to this API. With that said,
	// not totally clear how much we'd use Pulumi config, since can use the
	// managed console DB instead.
	c := exec.Command("pulumi", "config", "get", "whitelist")
	c.Dir = pulumiDir
	out, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%v is allowed!\n", string(out))
}

func prepareHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	cidr := buf.String()

	c := exec.Command("pulumi", "config", "set", "whitelist", cidr)
	c.Dir = pulumiDir
	err := c.Run()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%s added!\n", cidr)
}

func diffHandler(w http.ResponseWriter, r *http.Request) {
	// Call API directly.
	plan, err := DiffOrUpdate(true /*dryrun*/)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "Here's the plan!\n")
	fmt.Fprint(w, plan)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	// Call API directly.
	_, err := DiffOrUpdate(false /*dryrun*/)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "Cloud provider updated!\n")
}

func main() {
	// Fetching the current project requires reading a YAML file. Ideally, we'd
	// remove this dependency.
	os.Chdir(pulumiDir)

	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/prepare", prepareHandler)
	http.HandleFunc("/diff", diffHandler)
	http.HandleFunc("/write", writeHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
