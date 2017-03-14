package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

const (
	stdoutPipe = "\033[1;37m" + "out >>" + "\033[0m"
	stderrPipe = "\033[1;31m" + "err >>" + "\033[0m"
)

var newline = []byte{'\n'}

func fatalf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(2)
}

func main() {
	username := os.Getenv("USER")
	cmd := "tail -f "
	cmd += os.Args[1]

	key, err := ioutil.ReadFile("/Users/b4b4r07/.ssh/id_rsa")
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	client, err := ssh.Dial("tcp", "strong-panda:10022", config)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("unable to create session: %s", err)
	}
	defer session.Close()

	tail(cmd, "strong-panda:10022", config)
}

func tail(cmd, host string, config *ssh.ClientConfig) {
	defer func() {
		fmt.Fprintf(os.Stdout, "[%s] done\n", host)
	}()
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		fatalf("Failed to dial: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		fatalf("Failed to create stderr pipe: %v", err)
	}
	go pipe(host, stdoutPipe, stdout, os.Stdout)
	go pipe(host, stderrPipe, stderr, os.Stderr)
	if err := session.Run(cmd); err != nil {
		fatalf("Failed to run:", err)
	}
}

func pipe(host, name string, r io.Reader, w io.Writer) {
	prefix := fmt.Sprintf("%-10.10s %s ", host, name)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Fprintf(w, prefix)
		w.Write(scanner.Bytes())
		w.Write(newline)
	}
}
