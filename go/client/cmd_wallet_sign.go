// Copyright 2019 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"github.com/keybase/cli"
	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/stellar1"
	"golang.org/x/net/context"
)

type CmdWalletSign struct {
	libkb.Contextified
	XDR       string
	AccountID stellar1.AccountID
}

func newCmdWalletSign(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	cmd := &CmdWalletSign{
		Contextified: libkb.NewContextified(g),
	}
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "xdr",
			Usage: "Transaction envelope XDR. If not provided, will be read from stdin.",
		},
		cli.StringFlag{
			Name:  "account",
			Usage: "Account ID to sign with. Optional, if not provided, SourceAccount of the transaction will be used.",
		},
	}
	return cli.Command{
		Name:  "sign",
		Usage: "Sign a Stellar transaction created elsewhere",
		Action: func(c *cli.Context) {
			cl.ChooseCommand(cmd, "sign", c)
		},
		Flags: flags,
	}
}

func (c *CmdWalletSign) ParseArgv(ctx *cli.Context) error {
	c.XDR = ctx.String("xdr")
	c.AccountID = stellar1.AccountID(ctx.String("account"))
	return nil
}

func (c *CmdWalletSign) Run() (err error) {
	defer transformStellarCLIError(&err)
	cli, err := GetWalletClient(c.G())
	if err != nil {
		return err
	}

	var maybeAccount *stellar1.AccountID
	if !c.AccountID.IsNil() {
		maybeAccount = &c.AccountID
	}
	arg := stellar1.SignTransactionXdrLocalArg{
		EnvelopeXdr: c.XDR,
		AccountID:   maybeAccount,
	}
	res, err := cli.SignTransactionXdrLocal(context.Background(), arg)
	if err != nil {
		return err
	}
	c.G().UI.GetDumbOutputUI().Printf("%s\n", res)
	return nil
}

func (c *CmdWalletSign) GetUsage() libkb.Usage {
	return libkb.Usage{
		Config:    true,
		API:       true,
		KbKeyring: true,
	}
}
