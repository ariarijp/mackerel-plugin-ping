package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
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

func validate(ipAddr string) bool {
	r := regexp.MustCompile("^\\d+\\.\\d+\\.\\d+\\.\\d+$")
	return r.MatchString(ipAddr)
}

func parseHostsString(optHost string) ([]string, error) {
	hosts := strings.Split(optHost, ",")
	for _, host := range hosts {
		if !validate(host) {
			msg := fmt.Sprintf("error: %v must be ipv4 address format\n", host)
			return nil, errors.New(msg)
		}
	}

	return hosts, nil
}

func main() {
	optHost := flag.String("host", "127.0.0.1", "IP Address(es)")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	hosts, err := parseHostsString(*optHost)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	var pp PingPlugin
	pp.Hosts = hosts

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
