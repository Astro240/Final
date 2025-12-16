package api

import (
	"fmt"
	"net"
	"os"
	"runtime"
)

func AddHostEntry(domain string, ip string) error {
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

// GetIPv4 returns the local IPv4 address of the machine
func GetIPv4() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
