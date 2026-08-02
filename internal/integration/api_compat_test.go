package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/goed2k/daemon/internal/model"
)

type apiEnvelope struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeAPI(t *testing.T, resp *http.Response) apiEnvelope {
	t.Helper()
	defer resp.Body.Close()
	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return env
}

func TestHealth_NoAuthRequired(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/system/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeOK {
		t.Fatalf("code = %q, want OK", env.Code)
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	for _, key := range []string{"daemon_running", "engine_running", "state_store_ok", "rpc_available"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("health missing field %q", key)
		}
	}
}

func TestProtectedRoute_RequiresAuth(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/system/info")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeUnauthorized {
		t.Fatalf("code = %q, want UNAUTHORIZED", env.Code)
	}
}

func TestSystemInfo_ResponseShape(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	resp, err := http.DefaultClient.Do(authRequest(t, http.MethodGet, ts.URL+"/api/v1/system/info", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeOK {
		t.Fatalf("code = %q, want OK", env.Code)
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"daemon_version", "engine_running", "uptime_seconds", "rpc_listen", "default_download_dir"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("info missing field %q", key)
		}
	}
}

func TestNetworkDHTv6_CompatibleWhenEngineStopped(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	resp, err := http.DefaultClient.Do(authRequest(t, http.MethodGet, ts.URL+"/api/v1/network/dht-v6", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeOK {
		t.Fatalf("code = %q, want OK", env.Code)
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"bootstrapped", "live_nodes", "listen_port", "storage_point"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("dht_v6 missing field %q", key)
		}
	}
}

func TestTransfersList_EngineNotRunning(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	resp, err := http.DefaultClient.Do(authRequest(t, http.MethodGet, ts.URL+"/api/v1/transfers", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeEngineNotRunning {
		t.Fatalf("code = %q, want ENGINE_NOT_RUNNING", env.Code)
	}
}

func TestTransferPriority_InvalidBodyRejected(t *testing.T) {
	ts := newTestHTTPServer(t, nil, true)
	defer ts.Close()
	if !ts.Engine.IsRunning() {
		t.Fatal("engine should be running for priority validation test")
	}

	hash := "31D6CFE0D16AE931B73C59D7E0C089C0"
	body := bytes.NewBufferString(`{"priority":99}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers/"+hash+"/priority", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeBadRequest {
		t.Fatalf("code = %q, want BAD_REQUEST", env.Code)
	}
}

func TestTransferPriority_EngineNotRunning(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	hash := "31D6CFE0D16AE931B73C59D7E0C089C0"
	body := bytes.NewBufferString(`{"priority":2}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers/"+hash+"/priority", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeEngineNotRunning {
		t.Fatalf("code = %q, want ENGINE_NOT_RUNNING", env.Code)
	}
}

func TestTransferPriority_MalformedJSON(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	hash := "31D6CFE0D16AE931B73C59D7E0C089C0"
	body := bytes.NewBufferString(`{priority:2}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers/"+hash+"/priority", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeBadRequest {
		t.Fatalf("code = %q, want BAD_REQUEST", env.Code)
	}
}

func TestTransferHttpSource_EngineNotRunning(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	hash := "31D6CFE0D16AE931B73C59D7E0C089C0"
	body := bytes.NewBufferString(`{"url":"https://example.com/file.bin"}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers/"+hash+"/http-sources", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeEngineNotRunning {
		t.Fatalf("code = %q, want ENGINE_NOT_RUNNING", env.Code)
	}
}

func TestTransferHttpSource_MalformedJSON(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	hash := "31D6CFE0D16AE931B73C59D7E0C089C0"
	body := bytes.NewBufferString(`{url:bad}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers/"+hash+"/http-sources", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeBadRequest {
		t.Fatalf("code = %q, want BAD_REQUEST", env.Code)
	}
}

func TestJSONRequest_UnknownFieldRejected(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	body := bytes.NewBufferString(`{"ed2k_link":"ed2k://|file|a|1|31D6CFE0D16AE931B73C59D7E0C089C0|/","unknown_field":true}`)
	req := authRequest(t, http.MethodPost, ts.URL+"/api/v1/transfers", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	env := decodeAPI(t, resp)
	if env.Code != model.CodeBadRequest {
		t.Fatalf("code = %q, want BAD_REQUEST", env.Code)
	}
}

func TestCORS_Preflight(t *testing.T) {
	ts := newTestHTTPServer(t, nil, false)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/v1/system/info", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Allow-Origin = %q", got)
	}
}

func TestClientStatusDTO_JSONFields(t *testing.T) {
	// 确保新增字段的 JSON 标签稳定，避免破坏 WS / 未来 HTTP 客户端。
	raw := `{
		"engine_running": false,
		"servers": [],
		"transfers": [],
		"peers": [],
		"dht": {"bootstrapped": false, "firewalled": false, "live_nodes": 0},
		"dht_v6": {"bootstrapped": false, "live_nodes": 0, "listen_port": 0},
		"totals": {}
	}`
	var dto model.ClientStatusDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		t.Fatalf("unmarshal ClientStatusDTO: %v", err)
	}
	if dto.DHTv6.ListenPort != 0 {
		t.Fatalf("dht_v6 not parsed")
	}
}

func TestTransferDTO_JSONPriorityFields(t *testing.T) {
	raw := `{
		"hash":"31D6CFE0D16AE931B73C59D7E0C089C0",
		"file_name":"a",
		"file_path":"/tmp/a",
		"size":1,
		"create_time":0,
		"state":"DOWNLOADING",
		"paused":false,
		"download_rate":0,
		"upload_rate":0,
		"total_done":0,
		"total_received":0,
		"total_wanted":1,
		"eta":0,
		"num_peers":0,
		"active_peers":0,
		"downloading_pieces":0,
		"progress":0,
		"download_priority":2,
		"download_priority_label":"P2",
		"ed2k_link":""
	}`
	var dto model.TransferDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		t.Fatalf("unmarshal TransferDTO: %v", err)
	}
	if dto.DownloadPriority != 2 || dto.DownloadPriorityLabel != "P2" {
		t.Fatalf("priority fields not parsed: %+v", dto)
	}
}
