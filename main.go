package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"time"
)

var (
	cmdString string
	cmd       *exec.Cmd
)

func init() {
	flag.StringVar(&cmdString, "cmdString", "ping baidu.com", "command string")
}

func main() {
	flag.Parse()

	if cmdString == "" {
		fmt.Println("cmdString must not be empty...")
		return
	}

	ctx := context.Background()

	go func(ctx context.Context) {

		defer func() {
			if e := recover(); e != nil {
				fmt.Printf("crashed, err: %s\nstack:%s", e, string(debug.Stack()))
			}
		}()

		for {
			if cmd != nil {
				cmd.Process.Kill()
				time.Sleep(time.Second * 5)
			}
			cmd = exec.CommandContext(ctx, "bash", "-c", cmdString)
			cmdReaderStderr, err := cmd.StderrPipe()
			if err != nil {
				log.Printf("ERR:%s,restarting...\n", err)
				continue
			}
			cmdReader, err := cmd.StdoutPipe()
			if err != nil {
				log.Printf("ERR:%s,restarting...\n", err)
				continue
			}
			scanner := bufio.NewScanner(cmdReader)
			scannerStdErr := bufio.NewScanner(cmdReaderStderr)
			go func() {
				defer func() {
					if e := recover(); e != nil {
						fmt.Printf("crashed, err: %s\nstack:%s", e, string(debug.Stack()))
					}
				}()
				for scanner.Scan() {
					fmt.Println(scanner.Text())
				}
			}()
			go func() {
				defer func() {
					if e := recover(); e != nil {
						fmt.Printf("crashed, err: %s\nstack:%s", e, string(debug.Stack()))
					}
				}()
				for scannerStdErr.Scan() {
					fmt.Println(scannerStdErr.Text())
				}
			}()
			if err := cmd.Start(); err != nil {
				log.Printf("ERR:%s,restarting...\n", err)
				continue
			}
			pid := cmd.Process.Pid
			log.Printf("worker %s [PID] %d running...\n", os.Args[0], pid)
			if err := cmd.Wait(); err != nil {
				log.Printf("ERR:%s,restarting...", err)
				continue
			}
			log.Printf("worker %s [PID] %d unexpected exited, restarting...\n", os.Args[0], pid)

		}
	}(ctx)
	select {}
}
