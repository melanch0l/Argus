package environment

func (s *inspection) probe(requested bool) {
	st, summary := NotTested, "Not requested; use --privileged to explicitly authorize transient bpftool kernel capability tests"
	if requested {
		st, summary = Blocked, "Privileged feature probing requires Linux and effective UID 0; root alone does not establish BPF permissions"
	}
	var data struct {
		Programs map[string]bool     `json:"program_types"`
		Maps     map[string]bool     `json:"map_types"`
		Helpers  map[string][]string `json:"helpers"`
	}
	if requested && s.GOOS == "linux" && s.EUID == 0 {
		c := s.command("bpftool", false, "-j", "feature", "probe", "kernel")
		st = decodeObject(c, &data)
		if st == Pass && (data.Programs == nil || data.Maps == nil || data.Helpers == nil) {
			st = Fail
		}
		summary = "Actual privileged bpftool JSON feature probe; specific capabilities are evaluated separately"
	}
	s.add("bpf.feature_probe", st, true, summary, nil)
	for _, item := range []struct {
		id, key string
		values  map[string]bool
	}{
		{"bpf.program.tracepoint", "have_tracepoint_prog_type", data.Programs},
		{"bpf.program.cgroup_skb", "have_cgroup_skb_prog_type", data.Programs},
		{"bpf.map.ringbuf", "have_ringbuf_map_type", data.Maps},
	} {
		feature := st
		if st == Pass {
			value, exists := item.values[item.key]
			if !exists {
				feature = Blocked
			} else if !value {
				feature = Fail
			}
		}
		s.add(item.id, feature, true, "Require explicit true feature-probe evidence: "+item.key, nil)
	}
	for _, program := range []string{"tracepoint", "cgroup_skb"} {
		for _, helper := range []string{"bpf_ringbuf_output", "bpf_ringbuf_reserve", "bpf_ringbuf_submit", "bpf_ringbuf_discard"} {
			feature := st
			if st == Pass {
				values, exists := data.Helpers[program+"_available_helpers"]
				if !exists || values == nil {
					feature = Blocked
				} else {
					feature = Fail
					for _, value := range values {
						if value == helper {
							feature = Pass
							break
						}
					}
				}
			}
			s.add("bpf.helper."+program+"."+helper, feature, true, "Require named helper in this program type's probed helper list", nil)
		}
	}
}
