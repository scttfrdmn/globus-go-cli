// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package flows

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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
)

// useMockFlows points getClient at handler for the duration of the test.
func useMockFlows(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	orig := getClient
	getClient = func(ctx context.Context) (*flows.Client, error) {
		return flows.NewClient(ctx, &core.Config{BaseURL: srv.URL, AccessToken: "test-token", Scopes: []string{"test"}})
	}
	t.Cleanup(func() { getClient = orig })
}

// runCmd executes cmd with args in the given output format and returns stdout.
func runCmd(t *testing.T, cmd *cobra.Command, format string, args ...string) (string, error) {
	t.Helper()
	viper.Set("format", format)
	t.Cleanup(func() { viper.Set("format", "") })
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestParseOrderBy(t *testing.T) {
	got, err := parseOrderBy([]string{"created_timestamp:desc", "Edited_Timestamp:ASC"}, webInputOrderByFields)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "created_timestamp DESC,edited_timestamp ASC" {
		t.Errorf("got %q", got)
	}
	for _, bad := range []string{"created_timestamp", "name:ASC", "created_timestamp:UP"} {
		if _, err := parseOrderBy([]string{bad}, webInputOrderByFields); err == nil {
			t.Errorf("parseOrderBy(%q): expected error", bad)
		}
	}
}

func TestWebInputListPaginatesAndFilters(t *testing.T) {
	var queries []string
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/web_inputs" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		queries = append(queries, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("marker") == "" {
			_, _ = io.WriteString(w, `{"web_input_summaries":[{"id":"wi-1","status":"open","input_type":"selection","title":"Approve?","flow":{"id":"f1","title":"My Flow"},"run":{"id":"r1","label":"run one"}}],"marker":"m2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"web_input_summaries":[{"id":"wi-2","status":"closed","flow":{"id":"f1"},"run":{"id":"r2"}}],"marker":null}`)
	})

	out, err := runCmd(t, WebInputCmd(), "text", "list", "--filter-state", "open", "--filter-state", "closed", "--filter-role", "respondent")
	if err != nil {
		t.Fatal(err)
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 page requests, got %d (%v)", len(queries), queries)
	}
	if !strings.Contains(queries[0], "filter_states=open%2Cclosed") || !strings.Contains(queries[0], "filter_roles=respondent") ||
		!strings.Contains(queries[0], "orderby=created_timestamp+DESC") {
		t.Errorf("first query missing filters/orderby: %s", queries[0])
	}
	for _, want := range []string{"wi-1", "Approve?", "My Flow", "run one", "wi-2"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWebInputListLimitAndJSON(t *testing.T) {
	calls := 0
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = io.WriteString(w, `{"web_input_summaries":[{"id":"a","flow":{"id":"f"},"run":{"id":"r"}},{"id":"b","flow":{"id":"f"},"run":{"id":"r"}}],"marker":"more"}`)
	})
	out, err := runCmd(t, WebInputCmd(), "json", "list", "--limit", "1")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("limit reached on first page; expected 1 request, got %d", calls)
	}
	var doc struct {
		WebInputSummaries []map[string]interface{} `json:"web_input_summaries"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if len(doc.WebInputSummaries) != 1 || doc.WebInputSummaries[0]["id"] != "a" {
		t.Errorf("unexpected JSON: %s", out)
	}
}

func TestWebInputListRejectsBadChoice(t *testing.T) {
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) { t.Error("no request expected") })
	if _, err := runCmd(t, WebInputCmd(), "text", "list", "--filter-state", "pending"); err == nil {
		t.Error("expected error for invalid --filter-state")
	}
}

func TestWebInputShow(t *testing.T) {
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/web_inputs/wi-1" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"wi-1","status":"open","user_roles":["viewer","respondent"],
			"context":{"title":"Approve the run?","presentation_style":"text","text":"x"},
			"options":[{"option_id":"opt-yes","label":"Yes"}],
			"roles":{"viewer_urns":["urn:globus:auth:identity:v1"],"respondent_urns":["urn:globus:auth:identity:r1"]},
			"creator_urn":"urn:globus:auth:identity:c1",
			"flow":{"id":"f1","title":"My Flow"},"run":{"id":"r1","label":"run one"},
			"created_timestamp":"2026-10-01T12:00:00+00:00","edited_timestamp":"2026-10-01T12:00:00+00:00"}`)
	})
	out, err := runCmd(t, WebInputCmd(), "text", "show", "wi-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Web Input ID:", "wi-1", "Approve the run?", "viewer, respondent",
		"My Flow (f1)", "run one (r1)", "urn:globus:auth:identity:r1", "2026-10-01 12:00:00", "opt-yes  Yes"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWebInputRespond(t *testing.T) {
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/web_inputs/wi-1/respond" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["response"]["value"] != "opt-yes" {
			t.Errorf("unexpected body %v", body)
		}
		_, _ = io.WriteString(w, `{"status":"closed"}`)
	})
	out, err := runCmd(t, WebInputCmd(), "text", "respond", "wi-1", "opt-yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Status:") || !strings.Contains(out, "closed") {
		t.Errorf("unexpected output %q", out)
	}
}

func TestRegisteredAPIListAndShow(t *testing.T) {
	useMockFlows(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/registered_apis":
			if got := r.URL.Query()["orderby"]; len(got) != 1 || got[0] != "name ASC,id DESC" {
				t.Errorf("orderby = %v", got)
			}
			if r.URL.Query().Get("filter_roles") != "owner" {
				t.Errorf("filter_roles = %q", r.URL.Query().Get("filter_roles"))
			}
			_, _ = io.WriteString(w, `{"registered_apis":[{"id":"api-1","name":"Weather","created_timestamp":"2026-09-01T00:00:00Z","updated_timestamp":"2026-09-02T00:00:00Z"}],"has_next_page":false}`)
		case "/registered_apis/api-1":
			_, _ = io.WriteString(w, `{"id":"api-1","name":"Weather","status":"active","subscription_id":"sub-1",
				"roles":{"owners":["urn:globus:auth:identity:o1"]},
				"target":{"type":"openapi","openapi_version":"3.1","destination":{"method":"POST","url":"https://api.example/run"}},
				"created_timestamp":"2026-09-01T00:00:00Z"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	out, err := runCmd(t, RegisteredAPICmd(), "text", "list", "--filter-role", "owner", "--orderby", "name:ASC", "--orderby", "id:DESC")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "api-1") || !strings.Contains(out, "Weather") || !strings.Contains(out, "2026-09-02 00:00:00") {
		t.Errorf("list output:\n%s", out)
	}

	out, err = runCmd(t, RegisteredAPICmd(), "text", "show", "api-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Registered API ID:", "sub-1", "urn:globus:auth:identity:o1", "openapi", "3.1", "POST", "https://api.example/run"} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q:\n%s", want, out)
		}
	}
}
