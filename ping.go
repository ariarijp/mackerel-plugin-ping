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

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	fping "github.com/tatsushid/go-fastping"
)

type PingPlugin struct {
	Hosts    []string
	Labels   []string
	Tempfile string
}

func (pp PingPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})

	var prevRtt time.Duration

	pinger := fping.NewPinger()
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		rttMicroSec := float64(rtt.Nanoseconds()-prevRtt.Nanoseconds()) / 1000.0 / 1000.0
		fmt.Println(rttMicroSec)
		stat[escapeHostName(addr.String())] = rttMicroSec

		prevRtt = rtt
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
	for i := 0; i < len(pp.Hosts); i++ {
		metrics = append(metrics, mp.Metrics{
			Name:    escapeHostName(pp.Hosts[i]),
			Label:   pp.Labels[i],
			Diff:    false,
			Stacked: false,
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

func escapeHostNames(hosts string) string {
	return strings.Replace(escapeHostName(hosts), ",", "-", -1)
}

func validate(ipAddr string) bool {
	r := regexp.MustCompile("^\\d+\\.\\d+\\.\\d+\\.\\d+$")
	return r.MatchString(ipAddr)
}

func parseHostsString(optHost string) ([]string, []string, error) {
	hosts := strings.Split(optHost, ",")
	ips, labels := make([]string, len(hosts)), make([]string, len(hosts))

	for i := 0; i < len(hosts); i++ {
		v := strings.SplitN(hosts[i], ":", 2)
		if !validate(v[0]) {
			msg := fmt.Sprintf("error: %v must be ipv4 address format\n", v[0])
			return nil, nil, errors.New(msg)
		}
		ips[i] = v[0]
		if len(v) == 2 {
			labels[i] = v[1]
		} else {
			labels[i] = ips[i]
		}
	}

	return ips, labels, nil
}

func main() {
	optHost := flag.String("host", "127.0.0.1:localhost", "IP Address[:Metric label],IP[:Label],...")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	hosts, labels, err := parseHostsString(*optHost)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	var pp PingPlugin
	pp.Hosts = hosts
	pp.Labels = labels

	helper := mp.NewMackerelPlugin(pp)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-ping-%s", escapeHostName(strings.Join(hosts[:], "-")))
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
		helper.OutputDefinitions()
	} else {
		helper.OutputValues()
	}
}
