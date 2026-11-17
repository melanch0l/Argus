package environment

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

type inspection struct {
	Inspector
	report      Report
	ctx         context.Context
	localKernel string
	localArch   string
}

func (i Inspector) Inspect(ctx context.Context, options Options) Report {
	s := inspection{Inspector: i, ctx: ctx, report: Report{
		SchemaVersion: 1, RuntimeOS: i.GOOS, RuntimeArch: i.GOARCH,
		PrivilegedRequested: options.Privileged, Checks: []Check{}, Commands: []Command{},
		Limitations: []string{
			"Phase 2 discovery only; intended program load, verifier acceptance, attach and controlled workload validation are not performed until Phase 3.",
			"Hook filenames do not establish attach permissions or coverage. Kernel version alone does not establish support.",
			"Strict Docker checks require the standard local Unix socket with no DOCKER_HOST/DOCKER_CONTEXT overrides and matching local kernel/architecture. This is snapshot consistency, not cryptographic host identity; a socket proxy, bind mount or concurrent context change cannot be ruled out.",
			"No host names, boot IDs, container inventory, DSNs, environment dumps or raw command stderr are collected. Docker and feature-probe raw stdout are omitted. Tool stdout retains only a bounded version/identity line and llc BPF target lines, not installation paths or full diagnostics.",
			"The optional privileged bpftool feature probe performs transient kernel BPF capability tests; it is not an intended monitoring-program load/attach test.",
		},
	}}
	linux := i.GOOS == "linux"
	status := Blocked
	if linux {
		status = Pass
	}
	s.add("os.linux", status, true, "Linux is required for host eBPF inspection", nil)
	status = Fail
	if i.GOARCH == "amd64" || i.GOARCH == "arm64" {
		status = Pass
	}
	s.add("runtime.architecture", status, true, "Initial supported runtime architectures: amd64 and arm64", i.GOARCH)
	if linux {
		s.linuxFiles()
	} else {
		s.blockLinuxFiles()
	}
	for _, item := range []struct {
		id, name string
		args     []string
		required bool
	}{
		{"kernel.release", "uname", []string{"-r"}, true},
		{"kernel.architecture", "uname", []string{"-m"}, true},
		{"tool.clang", "clang", []string{"--version"}, true},
		{"tool.llvm", "llvm-config", []string{"--version"}, true},
		{"tool.llc", "llc", []string{"--version"}, true},
		{"tool.go", "go", []string{"version"}, true},
		{"tool.node", "node", []string{"--version"}, true},
		{"tool.npm", "npm", []string{"--version"}, true},
		{"tool.pg_config", "pg_config", []string{"--version"}, false},
		{"tool.psql", "psql", []string{"--version"}, false},
		{"tool.bpftool", "bpftool", []string{"version"}, true},
	} {
		c := s.command(item.name, true, item.args...)
		st := commandStatus(c)
		if st == Pass && strings.TrimSpace(c.Stdout) == "" {
			st = Fail
		}
		if strings.HasPrefix(item.id, "kernel.") && !linux {
			st = Blocked
		}
		if item.id == "kernel.release" && st == Pass && !knownKernelRelease(c.Stdout) {
			st = Fail
		}
		if item.id == "kernel.architecture" && st == Pass {
			want := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[i.GOARCH]
			if want == "" || strings.TrimSpace(c.Stdout) != want {
				st = Fail
			}
		}
		summary := "Recorded command version/identity"
		switch item.id {
		case "kernel.release":
			if st == Pass {
				s.localKernel = strings.TrimSpace(c.Stdout)
			}
		case "kernel.architecture":
			if st == Pass {
				s.localArch = strings.TrimSpace(c.Stdout)
			}
		case "tool.go":
			summary = "Require Go 1.25+"
			if st == Pass && !go125(c.Stdout) {
				st = Fail
			}
		case "tool.node":
			summary = "Require stable Node >=22.18.0 <23, matching web/package.json"
			if st == Pass && !stableVersionAtLeast(c.Stdout, "v", 22, 18, 0) {
				st = Fail
			}
		case "tool.npm":
			summary = "Require stable npm >=10.9.3 <11, matching web/package.json"
			if st == Pass && !stableVersionAtLeast(c.Stdout, "", 10, 9, 3) {
				st = Fail
			}
		}
		s.add(item.id, st, item.required, summary, nil)
		if item.id == "tool.llc" {
			bpf := st
			if bpf == Pass && !hasBPFTarget(c.Stdout) {
				bpf = Fail
			}
			s.add("llvm.bpf_target", bpf, true, "llc registered targets must explicitly include a BPF target", nil)
		}
	}
	s.docker()
	s.probe(options.Privileged)
	s.add("phase3.load_attach", NotTested, false, "Intended program load/attach and controlled workload tests have not been performed", nil)
	s.report.StrictPassed = s.report.StrictPass()
	return s.report
}

func (s *inspection) add(id string, status Status, required bool, summary string, evidence any) {
	s.report.Checks = append(s.report.Checks, Check{id, status, required, summary, evidence})
}

func (s *inspection) command(name string, output bool, args ...string) Command {
	c := s.Runner.Run(s.ctx, name, args...)
	log := c
	log.Stdout = ""
	if output {
		log.Stdout = selectedCommandOutput(name, c.Stdout)
	}
	log.OutputOmitted = !output || log.Stdout != c.Stdout
	s.report.Commands = append(s.report.Commands, log)
	return c
}

// Version tools may print user-specific installation paths after their version
// line. Retain only the minimal identity and explicit registered BPF targets.
func selectedCommandOutput(name, output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	first := strings.TrimSpace(lines[0])
	if len(first) > 4096 {
		first = first[:4096]
	}
	if name == "llc" {
		_, targets, ok := strings.Cut(output, "Registered Targets:")
		if ok {
			for _, line := range strings.Split(targets, "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 3 && fields[1] == "-" && (fields[0] == "bpf" || fields[0] == "bpfel" || fields[0] == "bpfeb") {
					first += "\nregistered_target=" + fields[0]
				}
			}
		}
	}
	return first
}

func knownKernelRelease(output string) bool {
	value := strings.TrimSpace(output)
	if len(strings.Fields(value)) != 1 {
		return false
	}
	parts := strings.SplitN(value, ".", 3)
	if len(parts) < 2 {
		return false
	}
	major, e1 := strconv.Atoi(parts[0])
	minor, e2 := strconv.Atoi(parts[1])
	return e1 == nil && e2 == nil && major > 0 && minor >= 0
}

func fileStatus(err error) Status {
	if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
		return Blocked
	}
	return Fail
}

func (s *inspection) read(path string) ([]byte, Status) {
	b, truncated, err := s.Files.Read(path, MaxFileBytes)
	if err != nil {
		return nil, fileStatus(err)
	}
	if truncated {
		return nil, Fail
	}
	return b, Pass
}

func (s *inspection) linuxFiles() {
	b, st := s.read("/etc/os-release")
	osInfo := map[string]string{}
	if st == Pass {
		for _, line := range strings.Split(string(b), "\n") {
			key, val, ok := strings.Cut(line, "=")
			if ok && (key == "ID" || key == "VERSION_ID") {
				osInfo[key] = strings.Trim(val, "\"'")
			}
		}
		if osInfo["ID"] != "ubuntu" || osInfo["VERSION_ID"] != "24.04" {
			st = Fail
		}
	}
	s.add("os.ubuntu_24_04", st, true, "Target distribution is Ubuntu 24.04", osInfo)
	b, _, err := s.Files.Read("/sys/kernel/btf/vmlinux", 1)
	st = Pass
	if err != nil {
		st = fileStatus(err)
	} else if len(b) == 0 {
		st = Fail
	}
	s.add("kernel.btf", st, true, "Running-kernel BTF must be readable and nonempty (contents omitted)", nil)
	b, st = s.read("/proc/self/mountinfo")
	if st == Pass {
		st = Fail
		for _, line := range strings.Split(string(b), "\n") {
			left, right, ok := strings.Cut(line, " - ")
			fields := strings.Fields(left)
			if ok && len(fields) >= 6 && fields[4] == "/sys/fs/cgroup" && strings.HasPrefix(right, "cgroup2 ") {
				st = Pass
			}
		}
	}
	s.add("cgroup.v2_mount", st, true, "A cgroup2 filesystem must be mounted at /sys/fs/cgroup", nil)
	b, st = s.read("/sys/fs/cgroup/cgroup.controllers")
	controllers := []string{}
	if st == Pass {
		for _, v := range strings.Fields(string(b)) {
			switch v {
			case "cpu", "cpuset", "io", "memory", "hugetlb", "pids", "rdma", "misc", "dmem":
				controllers = append(controllers, v)
			}
		}
		if len(controllers) == 0 {
			st = Fail
		}
	}
	s.add("cgroup.v2_controllers", st, true, "Read cgroup v2 controller names; no process or cgroup identities collected", controllers)
	for _, file := range []struct{ id, path string }{
		{"security.lockdown", "/sys/kernel/security/lockdown"},
		{"security.lsm", "/sys/kernel/security/lsm"},
	} {
		b, st = s.read(file.path)
		s.add(file.id, st, false, "Read security policy state; not proof of BPF permissions", strings.TrimSpace(string(b)))
	}
	s.hooks()
}

func (s *inspection) hooks() {
	for _, hook := range hookList() {
		paths := []string{}
		st := Blocked
		for _, root := range []string{"/sys/kernel/tracing/events/", "/sys/kernel/debug/tracing/events/"} {
			path := root + hook.name + "/format"
			b, _, err := s.Files.Read(path, 1)
			if err == nil && len(b) > 0 {
				paths = append(paths, path)
				st = Pass
			}
		}
		s.add("hook."+strings.ReplaceAll(hook.name, "/", "."), st, hook.required, "Readable tracefs filename only; no hook loaded or attached", paths)
	}
}

func hookList() []struct {
	name     string
	required bool
} {
	return []struct {
		name     string
		required bool
	}{
		{"sched/sched_process_exec", true}, {"sched/sched_process_fork", false},
		{"sched/sched_process_exit", false}, {"syscalls/sys_enter_openat", false},
		{"syscalls/sys_enter_connect", false}, {"syscalls/sys_enter_setuid", false},
	}
}

func (s *inspection) blockLinuxFiles() {
	for _, c := range []struct {
		id       string
		required bool
	}{
		{"os.ubuntu_24_04", true}, {"kernel.btf", true}, {"cgroup.v2_mount", true},
		{"cgroup.v2_controllers", true}, {"security.lockdown", false}, {"security.lsm", false},
	} {
		s.add(c.id, Blocked, c.required, "Requires inspection on the actual Linux host", nil)
	}
	for _, h := range hookList() {
		s.add("hook."+strings.ReplaceAll(h.name, "/", "."), Blocked, h.required, "Tracefs discovery requires Linux", nil)
	}
}

func go125(output string) bool {
	fields := strings.Fields(output)
	if len(fields) < 3 || fields[0] != "go" || fields[1] != "version" || !strings.HasPrefix(fields[2], "go") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(fields[2], "go"), ".")
	if len(parts) < 2 {
		return false
	}
	major, e1 := strconv.Atoi(parts[0])
	minor, e2 := strconv.Atoi(parts[1])
	return e1 == nil && e2 == nil && (major > 1 || major == 1 && minor >= 25)
}

// Accept only stable numeric versions in the selected major release line.
// Generic success text, prereleases and multiline diagnostics are not evidence
// that the pinned frontend toolchain's engine requirements are satisfied.
func stableVersionAtLeast(output, prefix string, major, minor, patch int) bool {
	value := strings.TrimSpace(output)
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(value, prefix), ".")
	if len(parts) != 3 {
		return false
	}
	var version [3]int
	for index, part := range parts {
		if part == "" || len(part) > 1 && part[0] == '0' {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		version[index] = n
	}
	return version[0] == major && (version[1] > minor || version[1] == minor && version[2] >= patch)
}

func hasBPFTarget(output string) bool {
	_, targets, ok := strings.Cut(output, "Registered Targets:")
	if !ok {
		return false
	}
	for _, line := range strings.Split(targets, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && (fields[0] == "bpf" || fields[0] == "bpfel" || fields[0] == "bpfeb") && fields[1] == "-" {
			return true
		}
	}
	return false
}

func decodeObject[T any](c Command, target *T) Status {
	st := commandStatus(c)
	if st == Pass && json.Unmarshal([]byte(c.Stdout), target) != nil {
		return Fail
	}
	return st
}
