package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	ipList := flag.String("ips", "", "list of IP addresses to execute command on")
	scriptFile := flag.String("script", "", "path to shell script file containing the update commands")
	user := flag.String("user", "user", "username to use when connecting to the remote systems")
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
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[ERROR] %s: %s\n", ip, err)
			} else {
				fmt.Printf("[LOG] %s:\n%s\n", ip, strings.TrimSpace(string(output)))
			}
		}(ip)
	}
	wg.Wait()
}
