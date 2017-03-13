package main

import (
	"fmt"
	"os"

	"github.com/b4b4r07/qa/ssh"
	"github.com/urfave/cli"
)

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
		Action:  cmdRunCmd,
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
	if err := s.init(); err != nil {
		return err
	}
	r := ssh.Run(q.session, SCRIPT_BRANCHES)
	fmt.Printf(r.Stdout)
	return nil
}

func cmdTailLog(c *cli.Context) error {
	return nil
}

func cmdRunCmd(c *cli.Context) error {
	return nil
}

func cmdConfig(c *cli.Context) error {
	return nil
}

type qa struct {
	session *ssh.Session
}

func (q *qa) init() error {
	session, err := ssh.DialKeyFile(
		"strong-panda", "b4b4r07", "/Users/b4b4r07/.ssh/id_rsa", 10,
	)
	q.session = session
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
