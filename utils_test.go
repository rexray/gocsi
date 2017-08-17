package gocsi

import (
	"os"
	"testing"
)

func TestGetCSIEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		proto    string
		addr     string
	}{
		{
			endpoint: "unix://path/to/sock.sock",
			proto:    "unix",
			addr:     "path/to/sock.sock",
		},
		{
			endpoint: "unix:///path/to/sock.sock",
			proto:    "unix",
			addr:     "/path/to/sock.sock",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			os.Setenv("CSI_ENDPOINT", tt.endpoint)
			p, a, err := GetCSIEndpoint()
			if err != nil {
				t.Errorf("Parsing of CSI_ENDPOINT returned err: %v", err)
			}
			if p != tt.proto || a != tt.addr {
				t.Errorf("Parsing of CSI_ENDPOINT incorrect, got: (%s,%s) want: (%s,%s)",
					p, a, tt.proto, tt.addr)
			}
		})
	}
}

func TestMissingCSIEndpoint(t *testing.T) {
	os.Unsetenv("CSI_ENDPOINT")
	_, _, err := GetCSIEndpoint()
	if err == nil {
		t.Fatal("No error returned when CSI_ENDPOINT not set")
	}
	if err != ErrMissingCSIEndpoint {
		t.Fatalf("Received unexpected error when CSI_ENDPOINT not set, got: %s want: %s",
			err, ErrMissingCSIEndpoint)
	}
}

func TestInvalidCSIEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
	}{
		{
			endpoint: "tcp5://localhost:5000",
		},
		{
			endpoint: "unixpcket://path/to/sock.sock",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			os.Setenv("CSI_ENDPOINT", tt.endpoint)
			_, _, err := GetCSIEndpoint()
			if err == nil {
				st.Fatal("No error returned when CSI_ENDPOINT is invalid")
			}
			if err != ErrInvalidCSIEndpoint {
				st.Fatalf("Received unexpected error when CSI_ENDPOINT set to invalid valud, got: %s want: %s",
					err, ErrInvalidCSIEndpoint)
			}
		})
	}
}
