package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/BurntSushi/toml"
	"github.com/b4b4r07/qa/ssh"
	"github.com/urfave/cli"
)

type config struct {
	SelectCmd string `toml:"selectcmd"`
	Editor    string `toml:"editor"`

	Hostname     string `toml:"hostname"`
	Username     string `toml:"username"`
	IdentifyFile string `toml:"identify_file"`
	Timeout      int    `toml:"timeout"`
}

var commands = []cli.Command{
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
		Usage:   "tail log",
		Action:  cmdTailLog,
	},
	{
		Name:    "run",
		Aliases: []string{},
		Usage:   "run command",
		Action:  cmdRunCommand,
	},
	{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "config qa tool",
		Action:  cmdConfig,
	},
}

func cmdSSH(c *cli.Context) error {
	return nil
}

func cmdBranches(c *cli.Context) error {
	var q qa
	if err := q.init(); err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	for _, vhost := range q.vhosts {
		fmt.Fprintf(w, "%s \t %s\n", vhost.name, vhost.branch)
	}
	w.Flush()
	// r = q.server.Exec("pwd")
	// fmt.Printf(r.Stdout)
	// r = q.server.Exec("pwd")
	// fmt.Printf(r.Stdout)
	// r = q.server.Exec("pwd")
	// fmt.Printf(r.Stdout)

	return nil
}

func cmdTailLog(c *cli.Context) error {
	return nil
}

func cmdRunCommand(c *cli.Context) error {
	return nil
}

// inspired by mattn/memo
func cmdConfig(c *cli.Context) error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return err
	}

	file := filepath.Join(os.Getenv("HOME"), ".config", "qa", "config.toml")
	return cfg.runcmd(cfg.Editor, file)
}

func (cfg *config) runcmd(command string, args ...string) error {
	command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func selectEnv() {}

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

	cfg.SelectCmd = "peco"
	cfg.Editor = func() string {
		if len(os.Getenv("EDITOR")) > 0 {
			return os.Getenv("EDITOR")
		}
		return "vim"
	}()
	cfg.Hostname = "example.com"
	cfg.Username = os.Getenv("USER")
	cfg.IdentifyFile = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	cfg.Timeout = 10

	return toml.NewEncoder(f).Encode(cfg)
}

type vhost struct {
	name, path, branch string
}

type qa struct {
	server *ssh.Session
	vhosts []vhost
}

func (q *qa) init() error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return err
	}
	conn, err := ssh.DialKeyFile(
		cfg.Hostname, cfg.Username, cfg.IdentifyFile, cfg.Timeout,
	)
	if err != nil {
		return err
	}
	q.server = conn

	result := ssh.Run(q.server, SCRIPT_BRANCHES)
	var vs []vhost
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}
		// TODO: trim unneeded chars e.g. \r
		l := strings.Split(line, "\t")
		if len(l) != 2 {
			return errors.New("invalid line")
		}
		vs = append(vs, vhost{name: filepath.Base(l[0]), path: l[0], branch: l[1]})
	}
	q.vhosts = vs

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "qa"
	app.Usage = "qa tool"
	app.Version = "0.1"
	app.Commands = commands
	app.Run(os.Args)
}
