package openf1

import (
	"testing"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "openf1" {
		t.Errorf("Scheme = %q, want openf1", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "openf1" {
		t.Errorf("Identity.Binary = %q, want openf1", info.Identity.Binary)
	}
}

func TestClassifyAlwaysErrors(t *testing.T) {
	_, _, err := Domain{}.Classify("any-input")
	if err == nil {
		t.Error("Classify: want error for query-driven domain, got nil")
	}
}

func TestLocateAlwaysErrors(t *testing.T) {
	_, err := Domain{}.Locate("session", "9158")
	if err == nil {
		t.Error("Locate: want error for query-driven domain, got nil")
	}
}
