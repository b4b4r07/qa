package main

import (
	// "fmt"

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
	return nil
}

func main() {
	session, err := ssh.DialKeyFile(
		"strong-panda", "b4b4r07", "/Users/b4b4r07/.ssh/id_rsa", 10,
	)
	if err != nil {
		panic(err)
	}
	session.Shell()
	// r := ssh.Run(session, tail)
	// fmt.Printf(r.Stdout)
	// fmt.Printf("%#v\n", r)
}
