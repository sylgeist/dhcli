package main

import (
	"do/doge/version"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/sseekamp/dhcli/cli"
	"runtime"
)

type versionCmd struct{}

type updateCmd struct {
	Staging bool `kong:"optional,short='s',help='Update to the latest staging build (default is the latest production build).'"`
	DryRun  bool `kong:"optional,short='d',help='Perform a dry run (do not actually update the binary).'"`
}

var dhcli struct {
	Search  cli.SearchCmd `kong:"cmd='',help='Search for an active lease'"`
	Status  cli.StatusCmd `kong:"cmd='',help='Show Kea daemon status'"`
	Logs    cli.LogsCmd   `kong:"cmd='',help='Show logs from Kea instance'"`
	Res     cli.ResCmd    `kong:"cmd='',help='Show address reservations'"`
	Update  updateCmd     `kong:"cmd='',help='Update dhcli version'"`
	Version versionCmd    `kong:"cmd='',help='Show dhcli version'"`
}

func (d *versionCmd) Run() error {
	fmt.Printf("runtime version %s\n", runtime.Version())
	fmt.Printf("dhcli commit %s\n", version.GetCommit())
	return nil
}

func (d *updateCmd) Run() error {
	return cli.Update(d.Staging, d.DryRun)
}

func main() {
	ctx := kong.Parse(&dhcli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
