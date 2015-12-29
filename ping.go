package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
	fping "github.com/tatsushid/go-fastping"
)

type PingPlugin struct {
	Hosts    []string
	Tempfile string
}

func (pp PingPlugin) FetchMetrics() (map[string]float64, error) {
	stat := make(map[string]float64)

	pinger := fping.NewPinger()
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		rttMicroSec := float64(rtt.Nanoseconds()) / 1000.0 / 1000.0
		stat[escapeHostName(addr.String())] = rttMicroSec
	}

	for _, host := range pp.Hosts {
		ra, err := net.ResolveIPAddr("ip4:icmp", host)
		if err != nil {
			return nil, err
		}

		pinger.AddIPAddr(ra)
	}

	err := pinger.Run()
	if err != nil {
		return nil, err
	}

	pinger.RunLoop()

	return stat, nil
}

func (pp PingPlugin) GraphDefinition() map[string](mp.Graphs) {
	metrics := []mp.Metrics{}
	for _, host := range pp.Hosts {
		metrics = append(metrics, mp.Metrics{
			Name:    escapeHostName(host),
			Label:   host,
			Diff:    false,
			Stacked: true,
		})
	}

	return map[string](mp.Graphs){
		"ping.rtt": mp.Graphs{
			Label:   "Ping Round Trip Times",
			Unit:    "float",
			Metrics: metrics,
		},
	}
}

func escapeHostName(host string) string {
	return strings.Replace(host, ".", "_", -1)
}

func main() {
	optHost := flag.String("host", "localhost", "Hostname")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var pp PingPlugin
	pp.Hosts = strings.Split(*optHost, ",")

	helper := mp.NewMackerelPlugin(pp)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-ping-%s", *optHost)
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
		helper.OutputDefinitions()
	} else {
		helper.OutputValues()
	}
}
