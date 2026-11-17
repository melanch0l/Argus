package environment

import (
	"context"
	"encoding/json"
	"testing"
)

// Start with otherwise complete evidence so each negative assertion proves
// the specific capability prevents strict success, not an unrelated blocker.
func TestProbeCapabilityCannotBorrowOtherEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, check string
		change      func(map[string]any)
		want        Status
	}{
		{"tracepoint_false", "bpf.program.tracepoint", func(data map[string]any) {
			data["program_types"].(map[string]any)["have_tracepoint_prog_type"] = false
		}, Fail},
		{"cgroup_skb_false", "bpf.program.cgroup_skb", func(data map[string]any) {
			data["program_types"].(map[string]any)["have_cgroup_skb_prog_type"] = false
		}, Fail},
		{"ringbuf_false", "bpf.map.ringbuf", func(data map[string]any) {
			data["map_types"].(map[string]any)["have_ringbuf_map_type"] = false
		}, Fail},
		{"tracepoint_cannot_borrow_cgroup_helpers", "bpf.helper.tracepoint.bpf_ringbuf_reserve", func(data map[string]any) {
			data["helpers"].(map[string]any)["tracepoint_available_helpers"] = []string{"bpf_ringbuf_output", "bpf_ringbuf_submit", "bpf_ringbuf_discard"}
		}, Fail},
		{"cgroup_cannot_borrow_tracepoint_helpers", "bpf.helper.cgroup_skb.bpf_ringbuf_reserve", func(data map[string]any) {
			data["helpers"].(map[string]any)["cgroup_skb_available_helpers"] = []string{"bpf_ringbuf_output", "bpf_ringbuf_submit", "bpf_ringbuf_discard"}
		}, Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			var data map[string]any
			if err := json.Unmarshal([]byte(f.results["bpftool -j feature probe kernel"].Stdout), &data); err != nil {
				t.Fatal(err)
			}
			tc.change(data)
			b, err := json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			f.results["bpftool -j feature probe kernel"] = success(string(b))
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "bpf.feature_probe", Pass)
			assertStatus(t, r, tc.check, tc.want)
			for _, check := range r.Checks {
				if check.Required && check.ID != tc.check && check.Status != Pass {
					t.Errorf("unexpected unrelated blocker: %+v", check)
				}
			}
			if r.StrictPassed || r.MonitoringReady {
				t.Fatal("strict accepted absent/unsupported intended capability")
			}
			assertStatus(t, r, "phase3.load_attach", NotTested)
		})
	}
}
