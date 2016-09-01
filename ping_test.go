package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGraphDefinition(t *testing.T) {
	var pp PingPlugin
	pp.Hosts = []string{"127.0.0.1"}
	pp.Labels = []string{"localhost"}

	gd := pp.GraphDefinition()

	actual := gd["ping.rtt"].Label
	expected := "Ping Round Trip Times"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = gd["ping.rtt"].Unit
	expected = "float"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual_stacked := gd["ping.rtt"].Metrics[0].Stacked
	expected_stacked := false
	if actual_stacked != expected_stacked {
		t.Errorf("got %v\nwant %v", actual_stacked, expected_stacked)
	}

	actual = fmt.Sprintf("%v", reflect.TypeOf(gd["ping.rtt"].Metrics))
	expected = "[]mackerelplugin.Metrics"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestEscapeHostName(t *testing.T) {
	actual := escapeHostName("127.0.0.1")
	expected := "127_0_0_1"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("8.8.8.8")
	expected = "8_8_8_8"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("8_8_8_8")
	expected = "8_8_8_8"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestValidate(t *testing.T) {
	actual := validate("127.0.0.1")
	expected := true
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("8.8.8.8")
	expected = true
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("8.8.8.")
	expected = false
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("localhost")
	expected = false
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestParseHostsString(t *testing.T) {
	actualIPs, actualLabels, err := parseHostsString("127.0.0.1")
	expected := []string{"127.0.0.1"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8,8.8.4.4")
	expected = []string{"8.8.8.8", "8.8.4.4"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8:google-public-dns-a")
	expected = []string{"8.8.8.8"}
	expected_labels := []string{"google-public-dns-a"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8:google-public-dns-a,8.8.4.4:google-public-dns-b")
	expected = []string{"8.8.8.8", "8.8.4.4"}
	expected_labels = []string{"google-public-dns-a", "google-public-dns-b"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected_labels[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	_, _, err = parseHostsString("8.8.8.", "1")
	if err == nil {
		t.Errorf("got %v", err)
	}

	actualIPs, actualLabels, err = parseHostsString("m.root-servers.net", "1")
	expected = []string{"202.12.27.33"}
	expected_labels = []string{"m.root-servers.net"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

  actualIPs, actualLabels, err = parseHostsString("m.root-servers.net:m-root")
	expected = []string{"202.12.27.33"}
	expected_labels = []string{"m-root"}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
}
