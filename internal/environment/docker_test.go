package environment

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDockerEndpointGate(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, hostOverride, contextOverride string
		unknownEnvironment                            bool
		goos                                          string
		want                                          Status
	}{
		{name: "standard_local", endpoint: "unix:///var/run/docker.sock", want: Pass},
		{name: "tcp_remote_matching_kernel", endpoint: "tcp://PRIVATE_HOST:2376", want: Fail},
		{name: "ssh_remote_matching_kernel", endpoint: "ssh://PRIVATE_USER:PRIVATE_CREDENTIAL@PRIVATE_HOST", want: Fail},
		{name: "tcp_loopback_not_standard_socket", endpoint: "tcp://127.0.0.1:2375", want: Fail},
		{name: "desktop_vm", endpoint: "unix:///PRIVATE_HOME/.docker/desktop/docker.sock", want: Fail},
		{name: "rootless_socket", endpoint: "unix:///run/user/1000/docker.sock", want: Fail},
		{name: "alternate_unix_socket", endpoint: "unix:///PRIVATE_PATH/docker.sock", want: Fail},
		{name: "empty_endpoint", want: Fail},
		{name: "host_override_ignores_local_context", endpoint: "unix:///var/run/docker.sock", hostOverride: "tcp://PRIVATE_HOST:2376", want: NotTested},
		{name: "context_override", endpoint: "unix:///var/run/docker.sock", contextOverride: "PRIVATE_CONTEXT", want: NotTested},
		{name: "conflicting_overrides", endpoint: "unix:///var/run/docker.sock", hostOverride: "tcp://PRIVATE_HOST:2376", contextOverride: "PRIVATE_CONTEXT", want: NotTested},
		{name: "even_local_override_is_conservative", endpoint: "unix:///var/run/docker.sock", hostOverride: "unix:///var/run/docker.sock", want: NotTested},
		{name: "nil_getter_is_unknown", endpoint: "unix:///var/run/docker.sock", unknownEnvironment: true, want: NotTested},
		{name: "macos_cannot_bind_linux_daemon", endpoint: "unix:///var/run/docker.sock", goos: "darwin", want: Blocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			endpoint, err := json.Marshal(tc.endpoint)
			if err != nil {
				t.Fatal(err)
			}
			f.results[commandKey("docker", "context", "inspect", "--format", dockerEndpointFormat)] = success(string(endpoint))
			if tc.unknownEnvironment {
				i.Getenv = nil
			} else {
				i.Getenv = func(key string) string {
					switch key {
					case "DOCKER_HOST":
						return tc.hostOverride
					case "DOCKER_CONTEXT":
						return tc.contextOverride
					default:
						t.Fatalf("unexpected environment lookup: %s", key)
						return ""
					}
				}
			}
			if tc.goos != "" {
				i.GOOS = tc.goos
			}
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "docker.local_endpoint", tc.want)
			assertStatus(t, r, "docker.host_consistency", tc.want)
			if r.StrictPassed != (tc.want == Pass) {
				t.Fatalf("strict gate disagrees with endpoint gate: %v", r.StrictPassed)
			}
			if r.MonitoringReady {
				t.Fatal("endpoint consistency must not establish monitoring readiness")
			}
			b, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(b), "PRIVATE_") {
				t.Fatal("endpoint, context or environment values leaked into report")
			}
			for _, command := range r.Commands {
				if command.Name == "docker" && len(command.Args) > 0 && command.Args[0] == "context" {
					if command.Stdout != "" || !command.OutputOmitted {
						t.Fatal("context stdout was not omitted")
					}
					if command.TimeoutMS != CommandTimeout.Milliseconds() {
						t.Fatal("context command lost bounded timeout")
					}
				}
			}
		})
	}
}

func TestDockerContextCommandFailures(t *testing.T) {
	code := 1
	truncated := success(`"unix:///var/run/docker.sock"`)
	truncated.StdoutTruncated = true
	for _, tc := range []struct {
		name    string
		command Command
		want    Status
	}{
		{"missing", Command{Error: "not_found"}, Blocked},
		{"failed", Command{ExitCode: &code, Error: "exit_nonzero"}, Fail},
		{"timed_out", Command{TimedOut: true, Error: "deadline_exceeded"}, Fail},
		{"truncated", truncated, Fail},
		{"malformed", success("not JSON PRIVATE_ENDPOINT"), Fail},
		{"wrong_type", success(`{"Host":"PRIVATE_ENDPOINT"}`), Fail},
		{"null", success(`null`), Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			f.results[commandKey("docker", "context", "inspect", "--format", dockerEndpointFormat)] = tc.command
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "docker.local_endpoint", tc.want)
			assertStatus(t, r, "docker.host_consistency", tc.want)
			if r.StrictPassed {
				t.Fatal("strict accepted unavailable/unusable context evidence")
			}
			b, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(b), "PRIVATE_ENDPOINT") {
				t.Fatal("unusable raw context response leaked")
			}
		})
	}
}

func TestDockerHostConsistency(t *testing.T) {
	for _, tc := range []struct {
		name, runtimeArch, unameArch, daemonArch, daemonKernel string
		want                                                   Status
	}{
		{"arm64_alias", "arm64", "aarch64", "arm64", "6.8.0-test", Pass},
		{"arm64_uname_name", "arm64", "aarch64", "aarch64", "6.8.0-test", Pass},
		{"amd64_alias", "amd64", "x86_64", "amd64", "6.8.0-test", Pass},
		{"amd64_uname_name", "amd64", "x86_64", "x86_64", "6.8.0-test", Pass},
		{"different_arch", "arm64", "aarch64", "x86_64", "6.8.0-test", Fail},
		{"unknown_arch", "arm64", "aarch64", "available", "6.8.0-test", Fail},
		{"vm_kernel", "arm64", "aarch64", "aarch64", "6.8.0-linuxkit", Fail},
		{"runtime_uname_disagree", "arm64", "x86_64", "arm64", "6.8.0-test", Blocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			i.GOARCH = tc.runtimeArch
			f.results["uname -m"] = success(tc.unameArch)
			key := commandKey("docker", "info", "--format", dockerInfoFormat)
			var info dockerInfo
			if err := json.Unmarshal([]byte(f.results[key].Stdout), &info); err != nil {
				t.Fatal(err)
			}
			info.Architecture, info.KernelVersion = tc.daemonArch, tc.daemonKernel
			b, err := json.Marshal(info)
			if err != nil {
				t.Fatal(err)
			}
			f.results[key] = success(string(b))
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "docker.local_endpoint", Pass)
			assertStatus(t, r, "docker.host_consistency", tc.want)
			if r.StrictPassed != (tc.want == Pass) {
				t.Fatal("strict gate disagrees with host consistency")
			}
		})
	}
}

func TestDockerBindingRequiresSuccessfulLocalUname(t *testing.T) {
	for _, key := range []string{"uname -r", "uname -m"} {
		t.Run(key, func(t *testing.T) {
			i, f, _ := fixture()
			delete(f.results, key)
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, "docker.host_consistency", Blocked)
			if r.StrictPassed {
				t.Fatal("strict accepted unknown local kernel facts")
			}
		})
	}
}
