//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"

	"github.com/pewssh/gosdk/core/zcncrypto"
	"github.com/pewssh/gosdk/zboxcore/client"
	"github.com/pewssh/gosdk/zcncore"
)

func setWallet(clientID, publicKey, privateKey, mnemonic string) error {
	if mnemonic == "" {
		return errors.New("mnemonic is required")
	}
	keys := []zcncrypto.KeyPair{
		{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	c := client.GetClient()
	c.Mnemonic = mnemonic
	c.ClientID = clientID
	c.ClientKey = publicKey
	c.Keys = keys

	w := &zcncrypto.Wallet{
		ClientID:  clientID,
		ClientKey: publicKey,
		Mnemonic:  mnemonic,
		Keys:      keys,
	}
	err := zcncore.SetWallet(*w, false)
	if err != nil {
		return err
	}

	zboxApiClient.SetWallet(clientID, privateKey, publicKey)

	return nil
}
