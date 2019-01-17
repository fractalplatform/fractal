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
	"os"

	"github.com/spf13/cobra"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/wallet/keystore"
)

type outputSign struct {
	Signature string
}

var (
	msgfileFlag string
)

var signMessageCmd = &cobra.Command{
	Use:   "signmessage",
	Short: "sign a message",
	Long: `
Sign the message with a keyfile.

To sign a message contained in a file, use the --msgfile flag.
`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		message := getMessage(args, 1)

		// Load the keyfile.
		keyfilepath := args[0]
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

		signature, err := crypto.Sign(signHash(message), key.PrivateKey)
		if err != nil {
			fmt.Println("Failed to sign message: ", err)
			return
		}
		out := outputSign{Signature: hex.EncodeToString(signature)}
		if jsonFlag {
			mustPrintJSON(out)
		} else {
			fmt.Println("Signature:", out.Signature)
		}
	},
}

type outputVerify struct {
	Success            bool
	RecoveredAddress   string
	RecoveredPublicKey string
}

var verifyMessageCmd = &cobra.Command{
	Use:   "verifymessage",
	Short: "verify the signature of a signed message",
	Long: `
Verify the signature of the message.
It is possible to refer to a file containing the message.`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		addressStr := args[0]
		signatureHex := args[1]
		message := getMessage(args, 2)

		if !common.IsHexAddress(addressStr) {
			fmt.Println("Invalid address: ", addressStr)
			return
		}
		address := common.HexToAddress(addressStr)
		signature, err := hex.DecodeString(signatureHex)
		if err != nil {
			fmt.Println("Signature encoding is not hexadecimal: ", err)
			return
		}

		recoveredPubkey, err := crypto.SigToPub(signHash(message), signature)
		if err != nil || recoveredPubkey == nil {
			fmt.Println("Signature verification failed: ", err)
			return
		}
		recoveredPubkeyBytes := crypto.FromECDSAPub(recoveredPubkey)
		recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)
		success := address == recoveredAddress

		out := outputVerify{
			Success:            success,
			RecoveredPublicKey: hex.EncodeToString(recoveredPubkeyBytes),
			RecoveredAddress:   recoveredAddress.Hex(),
		}
		if jsonFlag {
			mustPrintJSON(out)
		} else {
			if out.Success {
				fmt.Println("Signature verification successful!")
			} else {
				fmt.Println("Signature verification failed!")
			}
			fmt.Println("Recovered public key:", out.RecoveredPublicKey)
			fmt.Println("Recovered address:", out.RecoveredAddress)
		}
	},
}

func getMessage(args []string, msgarg int) []byte {
	if msgfileFlag != "" {
		if len(args) > msgarg {
			fmt.Println("Can't use --msgfile and message argument at the same time.")
			os.Exit(1)
		}
		msg, err := ioutil.ReadFile(msgfileFlag)
		if err != nil {
			fmt.Println("Can't read message file: ", err)
			os.Exit(1)
		}
		return msg
	} else if len(args) == msgarg+1 {
		return []byte(args[msgarg])
	}
	fmt.Println("Invalid number of arguments: want ", msgarg+1, ", got ", len(args))
	return nil
}

func init() {
	signMessageCmd.Flags().StringVar(&msgfileFlag, "msgfile", "", "file containing the message to sign/verify")
	RootCmd.AddCommand(signMessageCmd)
	RootCmd.AddCommand(verifyMessageCmd)
}
