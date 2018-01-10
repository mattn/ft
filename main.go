package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ft"
	app.Usage = "File Transferer"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		downloadCommand(),
		serveCommand(),
		listCommand(),
	}
	app.Run(os.Args)
}
