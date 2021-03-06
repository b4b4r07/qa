package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	termType = "xterm"
)

type clientPassword string

func (p clientPassword) Password(user string) (string, error) {
	return string(p), nil
}

// CLI comprises all info resulting from running a command via ssh
type CLI struct {
	Err    error  // internal or communication errors
	Status int    // the result code of the command itself
	Stdout string // stdout from the command
	Stderr string // stderr from the command
}

// Session allows for multiple commands to be run against an ssh connection
type Session struct {
	Client   *ssh.Client
	SSH      *ssh.Session
	out, err bytes.Buffer
}

type keychain struct {
	keys []ssh.Signer
}

// Close closes the ssh session
func (s *Session) Close() {
	s.SSH.Close()
	if s.Client != nil {
		s.Client.Close()
	}
}

// Clear clears the stdout and stderr buffers
func (s *Session) Clear() {
	s.out.Reset()
	s.err.Reset()
}

func OpenShell(privateKey []byte, addr string, port int32, user string) error {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}

	// Create client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	// Connect to ssh server
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", addr, port), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	// The following two lines makes the terminal work properly because of
	// side-effects I don't understand.
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer terminal.Restore(0, oldState)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}

func (k *keychain) PrivateKey(text []byte) error {
	key, err := ssh.ParsePrivateKey(text)
	if err != nil {
		return err
	}
	k.keys = append(k.keys, key)
	return nil
}

func (k *keychain) PrivateKeyFile(file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return k.PrivateKey(buf)
}

func keyAuth(key string) (ssh.AuthMethod, error) {
	k := new(keychain)
	if err := k.PrivateKey([]byte(key)); err != nil {
		return nil, err
	}
	return ssh.PublicKeys(k.keys...), nil
}

func keyFileAuth(file string) (ssh.AuthMethod, error) {
	k := new(keychain)
	if err := k.PrivateKeyFile(file); err != nil {
		return nil, err
	}
	return ssh.PublicKeys(k.keys...), nil
}

//DialKey will open an ssh session using an key key
func DialKey(server, username, key string, timeout int) (*Session, error) {
	auth, err := keyAuth(key)
	if err != nil {
		return nil, err
	}
	return DialSSH(server, username, timeout, auth)
}

//DialKeyFile will open an ssh session using an key key stored in keyfile
func DialKeyFile(server, username, keyfile string, timeout int) (*Session, error) {
	auth, err := keyFileAuth(keyfile)
	if err != nil {
		return nil, err
	}
	return DialSSH(server, username, timeout, auth)
}

//DialPassword will open an ssh session using the specified password
func DialPassword(server, username, password string, timeout int) (*Session, error) {
	return DialSSH(server, username, timeout, ssh.Password(password))
}

//DialSSH will open an ssh session using the specified authentication
func DialSSH(server, username string, timeout int, auth ...ssh.AuthMethod) (*Session, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: auth,
	}
	if strings.Index(server, ":") < 0 {
		server += ":10022"
	}
	conn, err := net.DialTimeout("tcp", server, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, server, config)
	if err != nil {
		return nil, err
	}
	return NewSession(ssh.NewClient(c, chans, reqs))
}

// NewSession will open an ssh session using the provided connection
func NewSession(client *ssh.Client) (*Session, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	s := &Session{SSH: session, Client: client}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,      // disable echoing
		ssh.TTY_OP_ISPEED: 115200, // input speed  = 115.2kbps
		ssh.TTY_OP_OSPEED: 115200, // output speed = 115.2kbps
	}
	// Request pseudo terminal
	if err := session.RequestPty(termType, 80, 40, modes); err != nil {
		client.Close()
		return nil, err
	}

	session.Stdout = &s.out
	session.Stderr = &s.err
	return s, nil
}

// Run will run a command in the session
func Run(session *Session, cmd string) CLI {
	var rc int
	var err error
	if err = session.SSH.Run(cmd); err != nil {
		if err2, ok := err.(*ssh.ExitError); ok {
			rc = err2.Waitmsg.ExitStatus()
		}
	}
	return CLI{err, rc, session.out.String(), session.err.String()}
}

func (s *Session) Exec(command string) *CLI {
	session, err := s.Client.NewSession()
	res := new(CLI)
	if err != nil {
		errText := err.Error()
		res.Stdout = ""
		res.Stderr = errText
	}

	var stdout, stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	var rc int
	if err := session.Run(command); err != nil {
		if err2, ok := err.(*ssh.ExitError); ok {
			rc = err2.Waitmsg.ExitStatus()
		}
	}
	session.Close()

	return &CLI{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Status: rc,
		Err:    err,
	}
}
