package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/b4b4r07/qa/ssh"
	"github.com/urfave/cli"
)

type config struct {
	SelectCmd string `toml:"selectcmd"`

	Hostname     string `toml:"hostname"`
	Username     string `toml:"username"`
	IdentifyFile string `toml:"identify_file"`
	Timeout      int    `toml:"timeout"`
}

var commands = []cli.Command{
	{
		Name:    "ssh",
		Aliases: []string{"s"},
		Usage:   "ssh",
		Action:  cmdSSH,
	},
	{
		Name:    "branch",
		Aliases: []string{"b"},
		Usage:   "list branches",
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
		Name:    "tail",
		Aliases: []string{},
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

	result := ssh.Run(q.server, SCRIPT_BRANCHES)

	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}
		fmt.Println(line)
	}
	return nil
}

func cmdTailLog(c *cli.Context) error {
	return nil
}

func cmdRunCommand(c *cli.Context) error {
	return nil
}

func cmdConfig(c *cli.Context) error {
	return nil
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
	cfg.Hostname = "example.com"
	cfg.Username = os.Getenv("USER")
	cfg.IdentifyFile = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	cfg.Timeout = 10

	return toml.NewEncoder(f).Encode(cfg)
}

type qa struct {
	server *ssh.Session
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
	q.server = conn
	return err
}

func main() {
	app := cli.NewApp()
	app.Name = "qa"
	app.Usage = "qa tool"
	app.Version = "0.1"
	app.Commands = commands
	app.Run(os.Args)
}
