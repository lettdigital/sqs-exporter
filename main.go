package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress         = flag.String("listen", ":9108", "Listen address for prometheus")
	metricsPath           = flag.String("path", "/metrics", "Path under which to expose metrics")
	updateIntervalPointer = flag.Int64("interval", 600, "Queue update interval, seconds")
	tagTeamPointer        = flag.String("tags", "", "Allows you to filter the queues based on the desired tag")
)

func main() {
	flag.Parse()

	tagTeam := *tagTeamPointer
	updateInterval := *updateIntervalPointer

	if len(os.Getenv("INTERVAL")) > 0 {
		if i, err := strconv.ParseInt(os.Getenv("INTERVAL"), 10, 64); err == nil {
			updateInterval = i
		}
	}

	if len(os.Getenv("TAGS")) > 0 {
		tagTeam = os.Getenv("TAGS")
	}

	ctx, cancel := context.WithCancel(context.Background())

	col := newCollector(ctx, time.Second*time.Duration(updateInterval), tagTeam)

	r := prometheus.NewRegistry()
	r.MustRegister(col)

	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`<html>
<head><title>AWS SQS exporter</title></head>
<body>
<h1>AWS SQS exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, *metricsPath)))
	})

	log.Printf("Starting http server, listening on %s\n", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		cancel()
		log.Fatal(err)
	}
}
