package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGraphDefinition(t *testing.T) {
	var pp PingPlugin
	pp.Hosts = []string{"127.0.0.1"}

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
