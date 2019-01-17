// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/utils/console"
	"github.com/fractalplatform/fractal/wallet"
	"github.com/fractalplatform/fractal/wallet/cache"
	"github.com/fractalplatform/fractal/wallet/keystore"
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manage Fractal presale wallets",
	Long: `
    fractal wallet import /path/to/my/presale.wallet

will prompt for your password and imports your presale account.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var importWalletCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Fractal presale wallet",
	Long: `
	fractal wallet [options] /path/to/my/presale.wallet

will prompt for your password and imports your presale account.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyfile := args[0]
		keyJSON, err := ioutil.ReadFile(keyfile)
		if err != nil {
			fmt.Println("Could not read wallet file: ", err)
			return
		}

		passphrase := getPassPhrase("", false)
		w, err := getWallet()
		if err != nil {
			fmt.Println("get wallet error ", err)
			return
		}
		acct, err := w.Import(keyJSON, passphrase, passphrase)
		if err != nil {
			fmt.Println("import wallet error ", err)
		}
		fmt.Printf("Address: {%x}\n", acct.Addr)
	},
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage accounts",
	Long: `

	Manage accounts, list all existing accounts, import a private key into a new
	account, create a new account or update an existing account.
	
	It supports interactive mode, when you are prompted for password.
	
	Make sure you remember the password you gave when creating a new account (with
	either new or import). Without it you are not able to unlock your account.
	
	Note that exporting your key in unencrypted format is NOT supported.
	
	Keys are stored under <DATADIR>/keystore.
	It is safe to transfer the entire directory or the individual keys therein
	between fractal nodes by simply copying.
	
	Make sure you backup your keys regularly.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var listAccountCmd = &cobra.Command{
	Use:   "list",
	Short: "Print summary of existing accounts",
	Long:  "Print a short summary of all accounts",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		wallet, err := getWallet()
		if err != nil {
			fmt.Println("get wallet error ", err)
		}
		for _, account := range wallet.Accounts() {
			fmt.Printf("Account: {%x} %s\n", account.Addr, account.Path)
		}
	},
}

var newAccountCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new account",
	Long: `
    fractal account new

Creates a new account and prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		password := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
		wallet, err := getWallet()
		if err != nil {
			fmt.Println("get wallet error ", err)
			return
		}
		account, err := wallet.NewAccount(password)
		if err != nil {
			fmt.Println("new account error ", err)
			return
		}

		fmt.Printf("Address: {%x}\n", account.Addr)
	},
}

var updateAccountCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing account",
	Long: `
    fractal account update <address>

Update an existing account.

The account is saved in the newest version in encrypted format, you are prompted
for a passphrase to unlock the account and another to save the updated file.

This same command can therefore be used to migrate an account of a deprecated
format to the newest format or change the password for an account.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		w, err := getWallet()
		if err != nil {
			fmt.Println("get wallet error ", err)
			return
		}
		for _, addr := range args {
			account := &cache.Account{Addr: common.HexToAddress(addr)}
			oldPassword := getPassPhrase("Please give old password.", true)
			newPassword := getPassPhrase("Please give a new password. Do not forget this password.", true)
			if err := w.Update(*account, oldPassword, newPassword); err != nil {
				fmt.Println("Could not update the account: ", addr, " ", err)
			}
		}
	},
}

var importAccountCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a private key into a new account",
	Long: `
    fractal account import <keyfile>

Imports an unencrypted private key from <keyfile> and creates a new account.
Prints the address.

The keyfile is assumed to contain an unencrypted private key in hexadecimal format.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

Note:
As you can directly copy your encrypted accounts to another ethereum instance,
this import mechanism is not needed when you transfer an account between
nodes.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyfile := args[0]
		key, err := crypto.LoadECDSA(keyfile)
		if err != nil {
			fmt.Println("Failed to load the private key: ", err)
			return
		}
		passphrase := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)

		w, err := getWallet()
		if err != nil {
			fmt.Println("get wallet error ", err)
			return
		}
		acct, err := w.ImportECDSA(key, passphrase)
		if err != nil {
			fmt.Println("Could not create the account: ", err)
			return
		}
		fmt.Printf("Address: {%x}\n", acct.Addr)
	},
}

func init() {
	walletCmd.PersistentFlags().StringVarP(&ftconfig.NodeCfg.DataDir, "datadir", "d", ftconfig.NodeCfg.DataDir, "Data directory for the databases and keystore")
	walletCmd.PersistentFlags().StringVar(&ftconfig.NodeCfg.KeyStoreDir, "keystore", ftconfig.NodeCfg.KeyStoreDir, "Directory for the keystore")
	walletCmd.PersistentFlags().BoolVar(&ftconfig.NodeCfg.UseLightweightKDF, "lightkdf", ftconfig.NodeCfg.UseLightweightKDF, "Reduce key-derivation RAM & CPU usage at some expense of KDF strength")

	accountCmd.PersistentFlags().StringVarP(&ftconfig.NodeCfg.DataDir, "datadir", "d", ftconfig.NodeCfg.DataDir, "Data directory for the databases and keystore")
	accountCmd.PersistentFlags().StringVar(&ftconfig.NodeCfg.KeyStoreDir, "keystore", ftconfig.NodeCfg.KeyStoreDir, "Directory for the keystore")
	accountCmd.PersistentFlags().BoolVar(&ftconfig.NodeCfg.UseLightweightKDF, "lightkdf", ftconfig.NodeCfg.UseLightweightKDF, "Reduce key-derivation RAM & CPU usage at some expense of KDF strength")

	walletCmd.AddCommand(importWalletCmd)
	accountCmd.AddCommand(listAccountCmd, newAccountCmd, updateAccountCmd, importAccountCmd)
	RootCmd.AddCommand(walletCmd, accountCmd)
}

// getPassPhrase retrieves the password associated with an account, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool) string {
	// Otherwise prompt the user for the password
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		fmt.Println("Failed to read passphrase: ", err)
		os.Exit(1)
	}
	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			fmt.Println("Failed to read passphrase confirmation: ", err)
			os.Exit(1)
		}
		if password != confirm {
			fmt.Println("Passphrases do not match")
			os.Exit(1)
		}
	}
	return password
}

func getWallet() (*wallet.Wallet, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	if ftconfig.NodeCfg.UseLightweightKDF {
		scryptN = keystore.LightScryptN
		scryptP = keystore.LightScryptP
	}

	var (
		keydir string
		err    error
	)
	switch {
	case filepath.IsAbs(ftconfig.NodeCfg.KeyStoreDir):
		keydir = ftconfig.NodeCfg.KeyStoreDir
	case ftconfig.NodeCfg.DataDir != "":
		if ftconfig.NodeCfg.KeyStoreDir == "" {
			keydir = filepath.Join(ftconfig.NodeCfg.DataDir, "keystore")
		} else {
			keydir, err = filepath.Abs(ftconfig.NodeCfg.KeyStoreDir)
		}
	case ftconfig.NodeCfg.KeyStoreDir != "":
		keydir, err = filepath.Abs(ftconfig.NodeCfg.KeyStoreDir)
	}
	if err != nil {
		fmt.Println("get keydir error ", err)
		return nil, err
	}

	if keydir == "" {
		keydir, err = ioutil.TempDir("", "tmpkeystore")
		if err != nil {
			fmt.Println("form keydir error ", err)
			return nil, err
		}
	}
	fmt.Println("keydir ", keydir)
	if err := os.MkdirAll(keydir, 0700); err != nil {
		fmt.Println("mkdir keydir error ", err)
		return nil, err
	}
	return wallet.NewWallet(keydir, scryptN, scryptP), nil
}
