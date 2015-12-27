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
	Host     string
	Tempfile string
}

func (pp PingPlugin) FetchMetrics() (map[string]float64, error) {
	pinger := fping.NewPinger()

	ra, err := net.ResolveIPAddr("ip4:icmp", pp.Host)
	if err != nil {
		return nil, err
	}
	pinger.AddIPAddr(ra)

	stat := make(map[string]float64)

	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		rttMicroSec := float64(rtt.Nanoseconds()) / 1000.0 / 1000.0
		stat[escapeHostName(pp.Host)] = rttMicroSec
	}

	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	return stat, nil
}

func (pp PingPlugin) GraphDefinition() map[string](mp.Graphs) {
	return map[string](mp.Graphs){
		"ping.rtt": mp.Graphs{
			Label: "Ping Round Trip Times",
			Unit:  "float",
			Metrics: [](mp.Metrics){
				mp.Metrics{
					Name:    escapeHostName(pp.Host),
					Label:   pp.Host,
					Diff:    false,
					Stacked: true,
				},
			},
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
	pp.Host = fmt.Sprintf("%s", *optHost)

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
