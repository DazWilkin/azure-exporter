package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/DazWilkin/azure-exporter/azure"
	"github.com/DazWilkin/azure-exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	envSubscription  string = "SUBSCRIPTION"
	envResourceGroup string = "RESOURCE_GROUP"
)

var (
	// GitCommit is the git commit value and is expected to be set during build
	GitCommit string
	// GoVersion is the Golang runtime version
	GoVersion = runtime.Version()
	// OSVersion is the OS version (uname --kernel-release) and is expected to be set during build
	OSVersion string
	// StartTime is the start time of the exporter represented as a UNIX epoch
	StartTime = time.Now().Unix()
)
var (
	endpoint    = flag.String("endpoint", ":9999", "The endpoint of the HTTP server")
	metricsPath = flag.String("path", "/metrics", "The path on which Prometheus metrics will be served")
)

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
func handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprint(w, "<h2>Azure Resources Exporter</h2>")
	fmt.Fprint(w, "<ul>")
	fmt.Fprintf(w, "<li><a href=\"%s\">metrics</a></li>", *metricsPath)
	fmt.Fprintf(w, "<li><a href=\"/healthz\">healthz</a></li>")
	fmt.Fprint(w, "</ul>")
}
func main() {
	flag.Parse()

	// Build variables
	if GitCommit == "" {
		log.Println("[main] GitCommit value unchanged: expected to be set during build")
	}
	if OSVersion == "" {
		log.Println("[main] OSVersion value unchanged: expected to be set during build")
	}

	// Environment variables
	subscription, ok := os.LookupEnv(envSubscription)
	if !ok {
		log.Fatalf("Expected environment to contain `%s`", envSubscription)
	}
	resourcegroup, ok := os.LookupEnv(envResourceGroup)
	if !ok {
		log.Fatalf("Expected environment to contain `%s`", envResourceGroup)
	}

	// Azure Identity (uses local `az` CLI credentials)
	creds, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Object that holds Azure-specific resources (e.g. Resource Groups)
	account := azure.NewAccount()

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewExporterCollector(OSVersion, GoVersion, GitCommit, StartTime))
	registry.MustRegister(collector.NewContainerAppsCollector(account, subscription, resourcegroup, creds))
	registry.MustRegister(collector.NewResourceGroupsCollector(account))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleRoot))
	mux.Handle("/healthz", http.HandlerFunc(handleHealthz))
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Printf("[main] Server starting (%s)", *endpoint)
	log.Printf("[main] metrics served on: %s", *metricsPath)
	log.Fatal(http.ListenAndServe(*endpoint, mux))
}
