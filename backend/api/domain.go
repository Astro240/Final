package api

import (
	"fmt"
	"os"
	"runtime"
)

func addHostEntry(domain string, ip string) error {
	hostsPath := ""
	if runtime.GOOS == "windows" {
		hostsPath = `C:\Windows\System32\drivers\etc\hosts`
	} else {
		hostsPath = `/etc/hosts`
	}

	// Open the hosts file in append mode
	file, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add the new domain entry
	entry := fmt.Sprintf("%s %s\n", ip, domain)
	if _, err := file.WriteString(entry); err != nil {
		return err
	}

	return nil
}
