// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// Wire shapes from upstream globus-sdk-python 4.10.0 testing data
// (transfer/v2/get_tunnel.py, get_stream_access_point.py).
const (
	activeTunnelDoc = `{"data": {"type": "Tunnel", "id": "tun-2",
	  "attributes": {"label": "live", "state": "ACTIVE", "status": "The tunnel is active",
	    "created_time": "2025-12-12T21:11:50.525278",
	    "listener_ip_address": "10.0.0.5", "listener_port": 50000,
	    "initiator_ip_address": "10.0.0.9", "initiator_port": 50001,
	    "lifetime_mins": 60, "restartable": true},
	  "relationships": {
	    "listener": {"data": {"id": "ap-l2", "type": "StreamAccessPoint"}},
	    "initiator": {"data": {"id": "ap-i2", "type": "StreamAccessPoint"}},
	    "owner": {"data": {"id": "owner-1", "type": "Identity"}}}}}`
	accessPointDoc = `{"data": {"type": "StreamAccessPoint", "id": "ap-1",
	  "attributes": {"display_name": "Buzz Dev Listener", "advertised_owner": "john@globus.org",
	    "organization": "Globus", "department": null, "contact_email": "a@b.org",
	    "tlsftp_server": "tlsftp://s-463c7.example.org:443"},
	  "relationships": {"host_endpoint": {"data": {"id": "host-ep", "type": "Endpoint"}}}}}`
)

// request records one call the mock Transfer server received.
type request struct {
	Method, Path string
	Body         map[string]interface{}
}

// useMockTransfer points getClient at a server answering every request with
// respond(req), recording requests into the returned slice.
func useMockTransfer(t *testing.T, respond func(req request) string) *[]request {
	t.Helper()
	var seen []request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{Method: r.Method, Path: r.URL.Path}
		if r.Body != nil {
			raw, _ := io.ReadAll(r.Body)
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &req.Body); err != nil {
					t.Errorf("request body is not JSON: %s", raw)
				}
			}
		}
		seen = append(seen, req)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, respond(req))
	}))
	t.Cleanup(srv.Close)
	orig := getClient
	getClient = func(ctx context.Context) (*transfer.Client, error) {
		return transfer.NewClient(ctx, &core.Config{BaseURL: srv.URL, AccessToken: "test-token", Scopes: []string{"test"}})
	}
	t.Cleanup(func() { getClient = orig })
	return &seen
}

func execute(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	viper.Set("format", "text")
	t.Cleanup(func() { viper.Set("format", "") })
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// attrs returns body.data.attributes from a JSON:API request document.
func attrs(t *testing.T, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	data, _ := body["data"].(map[string]interface{})
	a, _ := data["attributes"].(map[string]interface{})
	return a
}

func TestStreamsTunnelCreatePositionalArgs(t *testing.T) {
	seen := useMockTransfer(t, func(request) string { return activeTunnelDoc })
	out, err := execute(t, StreamsCmd(), "tunnel", "create", "ap-init", "ap-listen",
		"--label", "demo", "--lifetime-minutes", "90", "--restartable")
	if err != nil {
		t.Fatal(err)
	}
	if len(*seen) != 1 || (*seen)[0].Method != http.MethodPost || (*seen)[0].Path != "/v2/tunnels" {
		t.Fatalf("unexpected requests %+v", *seen)
	}
	body := (*seen)[0].Body
	data := body["data"].(map[string]interface{})
	rels := data["relationships"].(map[string]interface{})
	relID := func(name string) interface{} {
		return rels[name].(map[string]interface{})["data"].(map[string]interface{})["id"]
	}
	if relID("initiator") != "ap-init" || relID("listener") != "ap-listen" {
		t.Errorf("positional args mapped wrongly: %v", rels)
	}
	a := attrs(t, body)
	if a["label"] != "demo" || a["lifetime_mins"] != float64(90) || a["restartable"] != true {
		t.Errorf("unexpected attributes %v", a)
	}
	for _, want := range []string{"ID:", "tun-2", "State:", "ACTIVE", "Listener Access Point:", "ap-l2",
		"Listener IP:", "10.0.0.5", "Listener Port:", "50000", "Initiator Access Point:", "ap-i2", "50001"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestStreamsTunnelStopSendsStoppingState(t *testing.T) {
	seen := useMockTransfer(t, func(request) string { return activeTunnelDoc })
	if _, err := execute(t, StreamsCmd(), "tunnel", "stop", "tun-2"); err != nil {
		t.Fatal(err)
	}
	req := (*seen)[0]
	if req.Method != http.MethodPatch || req.Path != "/v2/tunnels/tun-2" {
		t.Fatalf("unexpected request %s %s", req.Method, req.Path)
	}
	a := attrs(t, req.Body)
	if len(a) != 1 || a["state"] != "STOPPING" {
		t.Errorf("stop must send only state=STOPPING, got %v", a)
	}
}

func TestStreamsTunnelUpdateLabel(t *testing.T) {
	seen := useMockTransfer(t, func(request) string { return activeTunnelDoc })
	if _, err := execute(t, StreamsCmd(), "tunnel", "update", "tun-2", "--label", "renamed"); err != nil {
		t.Fatal(err)
	}
	if a := attrs(t, (*seen)[0].Body); len(a) != 1 || a["label"] != "renamed" {
		t.Errorf("unexpected update attributes %v", a)
	}
}

func TestStreamsTunnelListAndDelete(t *testing.T) {
	useMockTransfer(t, func(req request) string {
		if req.Method == http.MethodDelete {
			return `{}`
		}
		return `{"data": [` + strings.TrimSuffix(strings.TrimPrefix(activeTunnelDoc, `{"data": `), `}`) + `]}`
	})
	out, err := execute(t, StreamsCmd(), "tunnel", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CreatedTime", "tun-2", "live", "ACTIVE", "2025-12-12 21:11:50"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q:\n%s", want, out)
		}
	}
	out, err = execute(t, StreamsCmd(), "tunnel", "delete", "tun-2")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "Tunnel deleted successfully" {
		t.Errorf("unexpected delete output %q", out)
	}
}

func TestStreamsTunnelEvents(t *testing.T) {
	useMockTransfer(t, func(request) string {
		return `{"data": [{"type": "TunnelEvent", "id": 7, "attributes": {"code": "LISTENER_CONNECTED",
			"is_error": false, "details": "listener connected", "time": "2025-12-12T21:12:00"}}]}`
	})
	out, err := execute(t, StreamsCmd(), "tunnel", "events", "tun-2")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Code", "LISTENER_CONNECTED", "false", "listener connected", "2025-12-12 21:12:00"} {
		if !strings.Contains(out, want) {
			t.Errorf("events output missing %q:\n%s", want, out)
		}
	}
}

func TestStreamsAccessPointShow(t *testing.T) {
	useMockTransfer(t, func(request) string { return accessPointDoc })
	out, err := execute(t, StreamsCmd(), "access-point", "show", "ap-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Display Name:", "Buzz Dev Listener", "Organization:", "Globus",
		"Contact Email:", "a@b.org", "Host Endpoint:", "host-ep"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestLegacyTunnelGroupStillWorks(t *testing.T) {
	seen := useMockTransfer(t, func(request) string { return activeTunnelDoc })
	if _, err := execute(t, TunnelCmd(), "create", "--initiator", "ap-init", "--listener", "ap-listen"); err != nil {
		t.Fatal(err)
	}
	if (*seen)[0].Path != "/v2/tunnels" {
		t.Errorf("unexpected path %s", (*seen)[0].Path)
	}
}
