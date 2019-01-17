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
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/wallet/keystore"
	"github.com/spf13/cobra"
)

var (
	privateFlag bool
)

type outputInspect struct {
	Address    string
	PublicKey  string
	PrivateKey string
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "inspect a keyfile",
	Long: `
Print various information about the keyfile.

Private key information can be printed by using the --private flag;
make sure to use this feature with great caution!`,
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

		// Output all relevant information we can retrieve.
		out := outputInspect{
			Address: key.Addr.Hex(),
			PublicKey: hex.EncodeToString(
				crypto.FromECDSAPub(&key.PrivateKey.PublicKey)),
		}
		if privateFlag {
			out.PrivateKey = hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))
		}

		if jsonFlag {
			mustPrintJSON(out)
		} else {
			fmt.Println("Address:       ", out.Address)
			fmt.Println("Public key:    ", out.PublicKey)
			if privateFlag {
				fmt.Println("Private key:   ", out.PrivateKey)
			}
		}
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&privateFlag, "private", false, "include the private key in the output")
	RootCmd.AddCommand(inspectCmd)
}
