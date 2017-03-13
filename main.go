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
				Name:  "fullpath",
				Usage: "",
			},
			cli.StringFlag{
				Name:  "pattern",
				Usage: "",
			},
		},
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

type qa struct {
	session *ssh.Session
}

func (q *qa) init() error {
	hoge, err := ssh.DialKeyFile(
		"strong-panda", "b4b4r07", "/Users/b4b4r07/.ssh/id_rsa", 10,
	)
	q.session = hoge
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
