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
	"strings"

	"github.com/spf13/cobra"

	"github.com/fractalplatform/fractal/wallet/keystore"
)

var (
	newPassphraseFlag string
)

var changePassphraseCmd = &cobra.Command{
	Use:   "changepassphrase",
	Short: "change the passphrase on a keyfile",
	Long: `
Change the passphrase of a keyfile.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyfilepath := args[0]
		// Read key from file.
		keyjson, err := ioutil.ReadFile(keyfilepath)
		if err != nil {
			fmt.Println("Failed to read the keyfile at ", keyfilepath, " : ", err)
			return
		}

		// Decrypt key with passphrase.
		passphrase := getPassphrase()
		key, err := keystore.DecryptKey(keyjson, passphrase)
		if err != nil {
			fmt.Println("Error decrypting key: ", err)
			return
		}

		// Get a new passphrase.
		fmt.Println("Please provide a new passphrase")
		var newPhrase string
		if newPassphraseFlag != "" {
			content, err := ioutil.ReadFile(newPassphraseFlag)
			if err != nil {
				fmt.Println("Failed to read new passphrase file ", newPassphraseFlag, " : ", err)
			}
			newPhrase = strings.TrimRight(string(content), "\r\n")
		} else {
			newPhrase = promptPassphrase(true)
		}

		// Encrypt the key with the new passphrase.
		newJSON, err := keystore.EncryptKey(key, newPhrase, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			fmt.Println("Error encrypting with new passphrase: ", err)
			return
		}

		// Then write the new keyfile in place of the old one.
		if err := ioutil.WriteFile(keyfilepath, newJSON, 600); err != nil {
			fmt.Println("Error writing new keyfile to disk: ", err)
			return
		}

		// Don't print anything.  Just return successfully,
		// producing a positive exit code.
	},
}

func init() {
	changePassphraseCmd.Flags().StringVar(&newPassphraseFlag, "newpasswordfile", "", "the file that contains the new passphrase for the keyfile")
	RootCmd.AddCommand(changePassphraseCmd)
}
