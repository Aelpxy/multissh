package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	ipList := flag.String("ips", "", "list of IP addresses to execute command on")
	scriptFile := flag.String("script", "", "path to shell script file containing the update commands")
	user := flag.String("user", "user", "username to use when connecting to the remote systems")
	timeout := flag.Duration("timeout", 30*time.Second, "timeout for the SSH connection")
	flag.Parse()

	script, err := os.ReadFile(*scriptFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ips := strings.Split(*ipList, ",")

	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			cmd := exec.Command("ssh", *user+"@"+ip, "bash -s")
			cmd.Stdin = strings.NewReader(string(script))

			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			cmd.Start()
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-ctx.Done():
				fmt.Printf("[ERROR] %s: timed out while connecting\n", ip)
				if err := cmd.Process.Kill(); err != nil {
					fmt.Printf("[ERROR] %s: failed to kill SSH process: %s\n", ip, err)
				}
			case err := <-done:
				if err != nil {
					fmt.Printf("[ERROR] %s: %s\n", ip, err)
				} else {
					fmt.Printf("[LOG] %s: finished successfully\n", ip)
				}
			}
		}(ip)
	}
	wg.Wait()
}
