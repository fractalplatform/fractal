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
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/wallet/keystore"
	"github.com/spf13/cobra"
)

const (
	defaultKeyfileName = "keyfile.json"
)

type outputGenerate struct {
	Address      string
	AddressEIP55 string
}

var (
	privatekeyFlag string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate new keyfile",
	Long: `
	Generate a new keyfile.
	
	If you want to encrypt an existing private key, it can be specified by setting
	--privatekey with the location of the file containing the private key.
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		keyfilepath := defaultKeyfileName
		if len(args) > 0 && args[0] != "" {
			keyfilepath = args[0]
		}

		if _, err := os.Stat(keyfilepath); err == nil {
			fmt.Println("Keyfile already exists at ", keyfilepath)
			return
		} else if !os.IsNotExist(err) {
			fmt.Println("Error checking if keyfile exists: ", err)
			return
		}

		var privateKey *ecdsa.PrivateKey
		var err error
		if privatekeyFlag != "" {
			// Load private key from file.
			privateKey, err = crypto.LoadECDSA(privatekeyFlag)
			if err != nil {
				fmt.Println("Can't load private key: ", err)
				return
			}
		} else {
			// If not loaded, generate random.
			privateKey, err = crypto.GenerateKey()
			if err != nil {
				fmt.Println("Failed to generate random private key: ", err)
				return
			}
		}

		// Create the keyfile object with a random UUID.
		key := &keystore.Key{
			Addr:       crypto.PubkeyToAddress(privateKey.PublicKey),
			PrivateKey: privateKey,
		}

		// Encrypt key with passphrase.
		passphrase := promptPassphrase(true)
		keyjson, err := keystore.EncryptKey(key, passphrase, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			fmt.Println("Error encrypting key: ", err)
			return
		}

		// Store the file to disk.
		if err := os.MkdirAll(filepath.Dir(keyfilepath), 0700); err != nil {
			fmt.Println("Could not create directory ", filepath.Dir(keyfilepath))
			return
		}
		if err := ioutil.WriteFile(keyfilepath, keyjson, 0600); err != nil {
			fmt.Println("Failed to write keyfile to : ", keyfilepath, " ", err)
			return
		}

		// Output some information.
		out := outputGenerate{
			Address: key.Addr.Hex(),
		}

		if jsonFlag {
			mustPrintJSON(out)
		} else {
			fmt.Println("Address:", out.Address)
		}
	},
}

func init() {
	generateCmd.Flags().StringVar(&privatekeyFlag, "privatekey", "", "file containing a raw private key to encrypt")
	RootCmd.AddCommand(generateCmd)
}
