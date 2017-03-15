package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/BurntSushi/toml"
	"github.com/b4b4r07/qa/ssh"
	"github.com/najeira/ltsv"
	"github.com/urfave/cli"
)

type qa struct {
	server *ssh.Session
	vhosts []vhost
	config config
}

type vhost struct {
	name, path   string
	branch, date string
}

var newline = []byte{'\n'}

const (
	Name        = "panda"
	Version     = "0.1"
	Description = "A CLI for QA tools"
)

type config struct {
	// TODO:
	SelectCmd string `toml:"selectcmd"`
	Editor    string `toml:"editor"`
	TailCmd   string `toml:"tailcmd"`

	Scripts scripts `toml:"scripts"`

	// TODO:
	Hostname      string `toml:"hostname"`
	Port          int32  `toml:"port"`
	Username      string `toml:"username"`
	IdentifyFile  string `toml:"identify_file"`
	Timeout       int    `toml:"timeout"`
	LogPathFormat string `toml:"log_path_format"`
}

type scripts struct {
	Branches string `toml:"branches"`
	Paths    string `toml:"paths"`
}

var commands = []cli.Command{
	{
		Name:    "debug",
		Aliases: []string{},
		Usage:   "",
		Action:  cmdDebug,
	},
	{
		Name:    "ssh",
		Aliases: []string{},
		Usage:   "connect to host via ssh",
		Action:  cmdSSH,
	},
	{
		Name:    "branch",
		Aliases: []string{"b"},
		Usage:   "list branches on host",
		Action:  cmdBranches,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "",
			},
			cli.BoolFlag{
				Name:  "ago",
				Usage: "",
			},
		},
	},
	{
		Name:    "log",
		Aliases: []string{"l"},
		Usage:   "do tail server log",
		Action:  cmdTailLog,
	},
	{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "configuare",
		Action:  cmdConfig,
	},
	{
		Name:    "db",
		Aliases: []string{},
		Usage:   "connet to database",
		Action:  cmdDB,
	},
}

func cmdDebug(c *cli.Context) error {
	var q qa
	if err := q.init(); err != nil {
		return err
	}
	fmt.Printf("%#v\n", q.vhosts)
	return nil
}

func cmdSSH(c *cli.Context) error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return err
	}

	privKey, err := ioutil.ReadFile(cfg.IdentifyFile)
	if err != nil {
		return err
	}

	return ssh.OpenShell(privKey, cfg.Hostname, cfg.Port, cfg.Username)
}

func cmdBranches(c *cli.Context) error {
	var q qa
	if err := q.init(); err != nil {
		return err
	}

	if err := q.setVhosts(q.config.Scripts.Branches); err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	for _, vhost := range q.vhosts {
		// TODO: flag
		fmt.Fprintf(w, "%s \t %s\n", vhost.name, vhost.branch)
	}
	w.Flush()

	return nil
}

func cmdTailLog(c *cli.Context) error {
	var q qa
	if err := q.init(); err != nil {
		return err
	}

	if err := q.setVhosts(q.config.Scripts.Branches); err != nil {
		return err
	}

	// make new session
	session, err := q.server.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	var text string
	for _, vhost := range q.vhosts {
		text += fmt.Sprintf("%s\n", vhost.name)
	}

	var buf bytes.Buffer
	err = q.config.runfilter(q.config.SelectCmd, strings.NewReader(text), &buf)
	if err != nil {
		return err
	}
	if buf.Len() == 0 {
		return errors.New("No files selected")
	}

	name := strings.Replace(buf.String(), "\n", "", -1)
	if q.config.TailCmd == "" {
		return errors.New("tail command: not found")
	}
	cmd := strings.Join([]string{q.config.TailCmd, fmt.Sprintf(q.config.LogPathFormat, name, name)}, " ")

	go pipe(stdout, os.Stdout)
	go pipe(stderr, os.Stderr)

	return session.Run(cmd)
}

// inspired by mattn/memo
func cmdConfig(c *cli.Context) error {
	// TODO:
	var q qa
	if err := q.init(); err != nil {
		return err
	}

	file := filepath.Join(os.Getenv("HOME"), ".config", "qa", "config.toml")
	return q.config.runcmd(q.config.Editor, file)
}

func cmdDB(c *cli.Context) error {
	return nil
}

func pipe(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		w.Write(scanner.Bytes())
		w.Write(newline)
	}
}

// TODO: rename all cfg methods
func (cfg *config) runcmd(command string, args ...string) error {
	command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (cfg *config) runfilter(command string, r io.Reader, w io.Writer) error {
	// TODO:
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = w
	cmd.Stdin = r
	return cmd.Run()
}

func (cfg *config) load() error {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "qa")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create directory: %v", err)
	}
	file := filepath.Join(dir, "config.toml")

	_, err := os.Stat(file)
	if err == nil {
		_, err := toml.DecodeFile(file, cfg)
		if err != nil {
			return err
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	cfg.SelectCmd = "fzf"
	cfg.TailCmd = "tail -f"
	cfg.Editor = func() string {
		if os.Getenv("EDITOR") != "" {
			return os.Getenv("EDITOR")
		}
		return "vim"
	}()
	cfg.Hostname = "example.com"
	cfg.Port = 22
	cfg.Username = os.Getenv("USER")
	cfg.IdentifyFile = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	cfg.Timeout = 10
	cfg.LogPathFormat = `/var/www/vhosts/%s/log/%s-app_error_log`

	return toml.NewEncoder(f).Encode(cfg)
}

func (q *qa) setVhosts(script string) error {
	if script == "" {
		// do nothing
		return nil
	}
	res := q.server.Exec(script)
	// res := ssh.Run(q.server, script)

	var vs []vhost
	b := bytes.NewBufferString(res.Stdout)
	reader := ltsv.NewReader(b)
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, host := range data {
		vs = append(vs, vhost{
			name:   host["name"],
			path:   host["path"],
			branch: host["branch"],
			date:   host["date"],
		})
	}
	q.vhosts = vs

	return nil
}

func (q *qa) init() error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return err
	}
	q.config = cfg

	conn, err := ssh.DialKeyFile(
		cfg.Hostname, cfg.Username, cfg.IdentifyFile, cfg.Timeout,
	)
	if err != nil {
		return err
	}
	q.server = conn

	return q.setVhosts(cfg.Scripts.Paths)
}

func main() {
	app := cli.NewApp()
	app.Name = Name
	app.Usage = Description
	app.Version = Version
	app.Commands = commands
	app.Run(os.Args)
}
