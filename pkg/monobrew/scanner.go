package monobrew

import (
	"os"
	"strings"
)

type Scanner struct {
	config *Config
}

func NewScanner(config *Config) *Scanner {
	return &Scanner{config: config}
}

func (s *Scanner) Scan() {
	s.ScanEtcIssue()
}

func (s *Scanner) ScanEtcIssue() {
	// Could certainly do this better. A lot better.

	_, err := os.Stat("/etc/issue")
	if os.IsNotExist(err) {
		return
	}

	contents, _ := os.ReadFile("/etc/issue")
	issue := string(contents)
	if strings.Contains(string(issue), "Debian GNU/Linux 13") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.debian", "1")
		s.UpdateStatus("os.debian.12", "1")
		s.UpdateStatus("os.debian.version", "13")
	}

	if strings.Contains(string(issue), "Debian GNU/Linux 12") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.debian", "1")
		s.UpdateStatus("os.debian.12", "1")
		s.UpdateStatus("os.debian.version", "12")
	}

	if strings.Contains(string(issue), "Debian GNU/Linux 11") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.debian", "1")
		s.UpdateStatus("os.debian.11", "1")
		s.UpdateStatus("os.debian.version", "11")
	}

	if strings.Contains(string(issue), "Ubuntu 20.04") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2004", "1")
		s.UpdateStatus("os.ubuntu.version", "20.04")
	}

	if strings.Contains(string(issue), "Ubuntu 20.10") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2010", "1")
		s.UpdateStatus("os.ubuntu.version", "20.10")
	}

	if strings.Contains(string(issue), "Ubuntu 21.04") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2104", "1")
		s.UpdateStatus("os.ubuntu.version", "2104")
	}

	if strings.Contains(string(issue), "Ubuntu 21.10") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2110", "1")
		s.UpdateStatus("os.ubuntu.version", "2110")
	}

	if strings.Contains(string(issue), "Ubuntu 22.04") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2204", "1")
		s.UpdateStatus("os.ubuntu.version", "22.04")
	}

	if strings.Contains(string(issue), "Ubuntu 22.10") {
		s.UpdateStatus("pkgs.apt", "1")
		s.UpdateStatus("os.ubuntu", "1")
		s.UpdateStatus("os.ubuntu.2210", "1")
		s.UpdateStatus("os.ubuntu.version", "22.14")
	}

}

func (s *Scanner) UpdateStatus(key string, value string) {
	s.config.Status[key] = value
}
