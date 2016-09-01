package main

import (
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
	Hosts       []string
	Labels      []string
	Tempfile    string
	Count       int
	WaitTime    int
	AcceptCount int
}

func (pp PingPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})
	total := make(map[string]float64)
	count := make(map[string]int)

	pinger := fping.NewPinger()
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		rttMilliSec := float64(rtt.Nanoseconds()) / 1000.0 / 1000.0
		total[escapeHostName(addr.String())] += rttMilliSec
		count[escapeHostName(addr.String())] += 1
	}

	for _, host := range pp.Hosts {
		pinger.AddIP(host)
	}

	pinger.MaxRTT = time.Millisecond * time.Duration(pp.WaitTime)

	for i := 0; i < pp.Count; i++ {
		err := pinger.Run()
		if err != nil {
			return nil, err
		}
	}
	for k, v := range total {
		if count[k] >= (pp.Count - pp.AcceptCount) {
			stat[k] = v / float64(count[k])
		}
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

func validate(ipAddr string) bool {
	r := regexp.MustCompile("^\\d+\\.\\d+\\.\\d+\\.\\d+$")
	return r.MatchString(ipAddr)
}

func parseHostsString(optHost string, strict ...string) ([]string, []string, error) {
	hosts := strings.Split(optHost, ",")
	ips, labels := make([]string, len(hosts)), make([]string, len(hosts))

	for i := 0; i < len(hosts); i++ {
		v := strings.SplitN(hosts[i], ":", 2)
		if !validate(v[0]) {
			ip, err := net.ResolveIPAddr("ip4", v[0])
			if err != nil {
				if strict[0] != "" {
					return nil, nil, err
				}
				continue
			}
			ips[i] = ip.String()
		} else {
			ips[i] = v[0]
		}

		if len(v) == 2 {
			labels[i] = v[1]
		} else {
			labels[i] = v[0]
		}
	}

	return ips, labels, nil
}

func main() {
	optHost := flag.String("host", "127.0.0.1:localhost", "IP Address[:Metric label],IP[:Label],...")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	optCount := flag.Int("count", 1, "Sending (and receiving) count ping packets.")
	optWaitTime := flag.Int("waittime", 1000, "Wait time, Max RTT(ms)")
	optAcceptCount := flag.Int("acceptmiss", 0, "Accept out of wait time count ping packets.")
	flag.Parse()

	hosts, labels, err := parseHostsString(*optHost, os.Getenv("MACKEREL_AGENT_PLUGIN_META"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}

	var pp PingPlugin
	pp.Hosts = hosts
	pp.Labels = labels
	pp.Count = *optCount
	pp.WaitTime = *optWaitTime
	pp.AcceptCount = *optAcceptCount

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
