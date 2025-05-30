package main

import (
	"flag"
	"html/template"
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
const (
	rootTemplate string = `
{{- define "content" }}
<!DOCTYPE html>
<html lang="en-US">
<head>
<title>Prometheus Exporter for Azure</title>
<style>
body {
  font-family: Verdana;
}
</style>
</head>
<body>
	<h2>Prometheus Exporter for Azure</h2>
	<ul>
	<li><a href="{{ .MetricsPath }}">metrics</a></li>
	<li><a href="/healthz">healthz</a></li>
	</ul>
</body>
</html>
{{- end }}`
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
	endpoint    = flag.String("endpoint", ":8080", "The endpoint of the HTTP server")
	metricsPath = flag.String("path", "/metrics", "The path on which Prometheus metrics will be served")
)

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Println(err)
	}
}
func handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	t := template.Must(template.New("content").Parse(rootTemplate))
	if err := t.ExecuteTemplate(w, "content", struct {
		MetricsPath string
	}{
		MetricsPath: *metricsPath,
	}); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
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

	// Azure Identity (uses local `az` CLI credentials)
	creds, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Object that holds Azure resources that are caches between collectors (e.g. Resource Groups)
	cache := azure.NewCache()

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewExporterCollector(OSVersion, GoVersion, GitCommit, StartTime))
	registry.MustRegister(collector.NewAccountCollector(subscription, creds))
	registry.MustRegister(collector.NewContainerAppsCollector(subscription, creds, cache))
	registry.MustRegister(collector.NewResourceGroupsCollector(subscription, creds, cache))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleRoot))
	mux.Handle("/healthz", http.HandlerFunc(handleHealthz))
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Printf("[main] Server starting (%s)", *endpoint)
	log.Printf("[main] metrics served on: %s", *metricsPath)
	log.Fatal(http.ListenAndServe(*endpoint, mux))
}
