package environment

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

type fakeFiles map[string]string

func (f fakeFiles) Read(path string, limit int) ([]byte, bool, error) {
	v, ok := f[path]
	if !ok {
		return nil, false, os.ErrNotExist
	}
	if len(v) > limit {
		return []byte(v[:limit]), true, nil
	}
	return []byte(v), false, nil
}

type fakeRunner struct {
	results map[string]Command
	calls   []string
}

func commandKey(name string, args ...string) string { return name + " " + strings.Join(args, " ") }
func (f *fakeRunner) Run(_ context.Context, name string, args ...string) Command {
	key := commandKey(name, args...)
	f.calls = append(f.calls, key)
	c, ok := f.results[key]
	if !ok {
		c.Error = "not_found"
	}
	c.Name, c.Args, c.TimeoutMS = name, append([]string{}, args...), CommandTimeout.Milliseconds()
	return c
}
func success(output string) Command { code := 0; return Command{ExitCode: &code, Stdout: output} }

func fixture() (Inspector, *fakeRunner, fakeFiles) {
	files := fakeFiles{
		"/etc/os-release":                   "ID=ubuntu\nVERSION_ID=\"24.04\"\n",
		"/sys/kernel/btf/vmlinux":           "BTF",
		"/proc/self/mountinfo":              "1 2 0:1 / /sys/fs/cgroup rw - cgroup2 cgroup rw\n",
		"/sys/fs/cgroup/cgroup.controllers": "cpuset cpu io memory pids\n",
		"/sys/kernel/security/lockdown":     "[none] integrity confidentiality\n",
		"/sys/kernel/security/lsm":          "lockdown,capability,apparmor,bpf\n",
	}
	for _, h := range hookList() {
		files["/sys/kernel/tracing/events/"+h.name+"/format"] = "name: hook"
	}
	f := &fakeRunner{results: map[string]Command{}}
	for name, args := range map[string][]string{
		"clang": {"--version"}, "llvm-config": {"--version"}, "llc": {"--version"},
		"go": {"version"}, "node": {"--version"}, "npm": {"--version"},
		"pg_config": {"--version"}, "psql": {"--version"}, "bpftool": {"version"},
	} {
		f.results[commandKey(name, args...)] = success("version 1.0\n")
	}
	f.results["uname -r"] = success("6.8.0-test\n")
	f.results["uname -m"] = success("aarch64\n")
	f.results["go version"] = success("go version go1.25.0 linux/arm64\n")
	f.results["node --version"] = success("v22.18.0\n")
	f.results["npm --version"] = success("10.9.3\n")
	f.results[commandKey("docker", "context", "inspect", "--format", dockerEndpointFormat)] = success(`"unix:///var/run/docker.sock"`)
	f.results["llc --version"] = success("LLVM version 18\nRegistered Targets:\n bpf - BPF (host endian)\n aarch64 - AArch64\n")
	f.results[commandKey("docker", "version", "--format", dockerVersionFormat)] = success(`{"ClientVersion":"28.0.0","ClientAPIVersion":"1.48","ServerVersion":"28.0.0","ServerAPIVersion":"1.48"}`)
	f.results[commandKey("docker", "info", "--format", dockerInfoFormat)] = success(`{"ServerVersion":"28.0.0","OSType":"linux","Architecture":"aarch64","KernelVersion":"6.8.0-test","CgroupDriver":"systemd","CgroupVersion":"2","SecurityOptions":["name=seccomp,profile=builtin","name=cgroupns"],"Name":"PRIVATE_HOST","ID":"PRIVATE_ID"}`)
	f.results["bpftool -j feature probe kernel"] = success(`{"program_types":{"have_tracepoint_prog_type":true,"have_cgroup_skb_prog_type":true},"map_types":{"have_ringbuf_map_type":true},"helpers":{"tracepoint_available_helpers":["bpf_ringbuf_output","bpf_ringbuf_reserve","bpf_ringbuf_submit","bpf_ringbuf_discard"],"cgroup_skb_available_helpers":["bpf_ringbuf_output","bpf_ringbuf_reserve","bpf_ringbuf_submit","bpf_ringbuf_discard"]},"private":"PRIVATE_PROBE"}`)
	return Inspector{GOOS: "linux", GOARCH: "arm64", EUID: 0, Runner: f, Files: files, Getenv: func(string) string { return "" }}, f, files
}
func findCheck(t *testing.T, r Report, id string) Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("missing check %s", id)
	return Check{}
}
func assertStatus(t *testing.T, r Report, id string, want Status) {
	t.Helper()
	if got := findCheck(t, r, id).Status; got != want {
		t.Errorf("%s: got %s, want %s", id, got, want)
	}
}

func TestFullEvidenceDoesNotClaimMonitoringReady(t *testing.T) {
	i, _, _ := fixture()
	r := i.Inspect(context.Background(), Options{true})
	if !r.StrictPassed || !r.StrictPass() {
		t.Fatalf("full fixture should satisfy strict prerequisites: %+v", r.Checks)
	}
	if r.MonitoringReady {
		t.Fatal("feature discovery must never establish monitoring readiness")
	}
	assertStatus(t, r, "phase3.load_attach", NotTested)
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"PRIVATE_HOST", "PRIVATE_ID", "PRIVATE_PROBE"} {
		if strings.Contains(string(b), secret) {
			t.Errorf("private data persisted: %s", secret)
		}
	}
	if r2 := i.Inspect(context.Background(), Options{true}); !reflect.DeepEqual(r, r2) {
		t.Fatal("report order/content is not deterministic")
	}
}

func TestPlatformAndPrivilegeGates(t *testing.T) {
	for _, tc := range []struct {
		name, goos string
		euid       int
		request    bool
		want       Status
	}{
		{"default_linux", "linux", 0, false, NotTested},
		{"unprivileged_linux", "linux", 1000, true, Blocked},
		{"mac_default", "darwin", 0, false, NotTested},
		{"mac_requested", "darwin", 0, true, Blocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			i.GOOS, i.EUID = tc.goos, tc.euid
			r := i.Inspect(context.Background(), Options{tc.request})
			assertStatus(t, r, "bpf.feature_probe", tc.want)
			if r.StrictPassed {
				t.Fatal("strict passed without Linux privileged probe")
			}
			for _, call := range f.calls {
				if call == "bpftool -j feature probe kernel" {
					t.Fatal("probe ran without authorization/platform/privileges")
				}
			}
			assertStatus(t, r, "tool.go", Pass)
			if tc.goos != "linux" {
				assertStatus(t, r, "kernel.btf", Blocked)
				assertStatus(t, r, "os.ubuntu_24_04", Blocked)
			}
		})
	}
}

func TestDockerMissingRootlessAndCgroupV1(t *testing.T) {
	for _, tc := range []struct {
		name, replacement, id string
		missing               bool
		want                  Status
	}{
		{"missing", "", "docker.info", true, Blocked},
		{"rootless", `{"ServerVersion":"28","KernelVersion":"6.8","Architecture":"arm64","OSType":"linux","CgroupDriver":"systemd","CgroupVersion":"2","SecurityOptions":["name=rootless"]}`, "docker.rootful", false, Fail},
		{"cgroup1", `{"ServerVersion":"28","KernelVersion":"6.8","Architecture":"arm64","OSType":"linux","CgroupDriver":"cgroupfs","CgroupVersion":"1","SecurityOptions":["name=seccomp"]}`, "docker.cgroup_v2", false, Fail},
		{"unknown_root_mode", `{"ServerVersion":"28","KernelVersion":"6.8","Architecture":"arm64","OSType":"linux","CgroupDriver":"systemd","CgroupVersion":"2"}`, "docker.rootful", false, Blocked},
		{"invalid_json", `not JSON`, "docker.info", false, Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			key := commandKey("docker", "info", "--format", dockerInfoFormat)
			if tc.missing {
				delete(f.results, key)
			} else {
				f.results[key] = success(tc.replacement)
			}
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, tc.id, tc.want)
			if r.StrictPassed {
				t.Fatal("strict passed without supported Docker evidence")
			}
		})
	}
}

func TestProbeRequiresSpecificEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, output, id string
		want             Status
	}{
		{"generic_available", `{"available":true}`, "bpf.feature_probe", Fail},
		{"missing_ringbuf", `{"program_types":{"have_tracepoint_prog_type":true,"have_cgroup_skb_prog_type":true},"map_types":{},"helpers":{}}`, "bpf.map.ringbuf", Blocked},
		{"unsupported_ringbuf", `{"program_types":{},"map_types":{"have_ringbuf_map_type":false},"helpers":{}}`, "bpf.map.ringbuf", Fail},
		{"helpers_other_program", `{"program_types":{},"map_types":{},"helpers":{"socket_filter_available_helpers":["bpf_ringbuf_reserve"]}}`, "bpf.helper.tracepoint.bpf_ringbuf_reserve", Blocked},
		{"empty_helper_list", `{"program_types":{},"map_types":{},"helpers":{"tracepoint_available_helpers":[]}}`, "bpf.helper.tracepoint.bpf_ringbuf_reserve", Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			f.results["bpftool -j feature probe kernel"] = success(tc.output)
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, tc.id, tc.want)
			if r.StrictPassed {
				t.Fatal("strict accepted generic/incomplete probe evidence")
			}
		})
	}
}

func TestTargetFileFailures(t *testing.T) {
	for _, tc := range []struct {
		name, path, contents, id string
		remove                   bool
	}{
		{"wrong_os", "/etc/os-release", "ID=debian\nVERSION_ID=12", "os.ubuntu_24_04", false},
		{"cgroup1_mount", "/proc/self/mountinfo", "1 2 0:1 / /sys/fs/cgroup rw - cgroup cgroup rw", "cgroup.v2_mount", false},
		{"missing_btf", "/sys/kernel/btf/vmlinux", "", "kernel.btf", true},
		{"truncated_os", "/etc/os-release", strings.Repeat("x", MaxFileBytes+1), "os.ubuntu_24_04", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, _, files := fixture()
			files[tc.path] = tc.contents
			want := Fail
			if tc.remove {
				delete(files, tc.path)
				want = Blocked
			}
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, tc.id, want)
			if r.StrictPassed {
				t.Fatal("strict accepted missing target prerequisite")
			}
		})
	}
}

func TestCommandFailureAndTruncation(t *testing.T) {
	code := 7
	for _, tc := range []struct {
		name string
		c    Command
		want Status
	}{
		{"missing", Command{Error: "not_found"}, Blocked},
		{"nonzero", Command{ExitCode: &code, Error: "exit_nonzero"}, Fail},
		{"timeout", Command{TimedOut: true, Error: "deadline_exceeded"}, Fail},
		{"truncated", func() Command { c := success("go version go1.25.0 linux/arm64"); c.StdoutTruncated = true; return c }(), Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			f.results["go version"] = tc.c
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "tool.go", tc.want)
			if r.StrictPassed {
				t.Fatal("strict accepted failed/truncated command")
			}
		})
	}
}

func TestStrictPassPolicy(t *testing.T) {
	if (Report{}).StrictPass() {
		t.Fatal("empty report passed strict policy")
	}
	for _, status := range []Status{Fail, Blocked, NotTested} {
		r := Report{Checks: []Check{{Status: Pass, Required: true}, {Status: status, Required: true}}}
		if r.StrictPass() {
			t.Errorf("strict accepted required %s", status)
		}
	}
	if !(Report{Checks: []Check{{Status: Pass, Required: true}, {Status: NotTested}}}).StrictPass() {
		t.Fatal("optional Phase 3 check blocked Phase 2")
	}
}

func TestKernelRequiresActualRelease(t *testing.T) {
	i, f, _ := fixture()
	f.results["uname -r"] = success("available")
	r := i.Inspect(context.Background(), Options{true})
	assertStatus(t, r, "kernel.release", Fail)
	if r.StrictPassed {
		t.Fatal("strict accepted a generic kernel response")
	}
}

func TestVersionOutputMinimization(t *testing.T) {
	i, f, _ := fixture()
	f.results["clang --version"] = success("clang version 18.0\nInstalledDir: /Users/PRIVATE_USER/bin\n")
	r := i.Inspect(context.Background(), Options{true})
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "PRIVATE_USER") {
		t.Fatal("installation path leaked into report")
	}
	if !strings.Contains(string(b), "clang version 18.0") {
		t.Fatal("version evidence lost")
	}
}

func TestGoAndLLVMParsing(t *testing.T) {
	for _, value := range []string{"available", "go version go1.24.9 linux/arm64", "go version unknown linux/amd64"} {
		if go125(value) {
			t.Errorf("accepted %q", value)
		}
	}
	for _, value := range []string{"go version go1.25.0 linux/arm64", "go version go1.26.1 linux/amd64"} {
		if !go125(value) {
			t.Errorf("rejected %q", value)
		}
	}
	for _, value := range []string{"BPF available", "Registered Targets:\n aarch64 - AArch64"} {
		if hasBPFTarget(value) {
			t.Errorf("accepted %q", value)
		}
	}
}
