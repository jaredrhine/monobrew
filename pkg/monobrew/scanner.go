package monobrew

import (
	"os"
	"strings"
)

func (r *Runner) Scan() {
	r.ScanEtcIssue()
}

func (r *Runner) ScanEtcIssue() {
	// Could certainly do this better. A lot better.

	_, err := os.Stat("/etc/issue")
	if os.IsNotExist(err) {
		return
	}

	contents, _ := os.ReadFile("/etc/issue")
	issue := string(contents)
	if strings.Contains(string(issue), "Debian GNU/Linux 13") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.debian", "1")
		r.UpdateStatus("os.debian.12", "1")
		r.UpdateStatus("os.debian.version", "13")
	}

	if strings.Contains(string(issue), "Debian GNU/Linux 12") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.debian", "1")
		r.UpdateStatus("os.debian.12", "1")
		r.UpdateStatus("os.debian.version", "12")
	}

	if strings.Contains(string(issue), "Debian GNU/Linux 11") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.debian", "1")
		r.UpdateStatus("os.debian.11", "1")
		r.UpdateStatus("os.debian.version", "11")
	}

	if strings.Contains(string(issue), "Ubuntu 20.04") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2004", "1")
		r.UpdateStatus("os.ubuntu.version", "20.04")
	}

	if strings.Contains(string(issue), "Ubuntu 20.10") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2010", "1")
		r.UpdateStatus("os.ubuntu.version", "20.10")
	}

	if strings.Contains(string(issue), "Ubuntu 21.04") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2104", "1")
		r.UpdateStatus("os.ubuntu.version", "2104")
	}

	if strings.Contains(string(issue), "Ubuntu 21.10") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2110", "1")
		r.UpdateStatus("os.ubuntu.version", "2110")
	}

	if strings.Contains(string(issue), "Ubuntu 22.04") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2204", "1")
		r.UpdateStatus("os.ubuntu.version", "22.04")
	}

	if strings.Contains(string(issue), "Ubuntu 22.10") {
		r.UpdateStatus("pkgs.apt", "1")
		r.UpdateStatus("os.ubuntu", "1")
		r.UpdateStatus("os.ubuntu.2210", "1")
		r.UpdateStatus("os.ubuntu.version", "22.14")
	}

}

func (r *Runner) UpdateStatus(key string, value string) {
	r.Config.Status[key] = value
}
