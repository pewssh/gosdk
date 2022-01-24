package zcnbridge

import (
	"fmt"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"

	hdw "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
)

// ListStorageAccounts List available accounts
func ListStorageAccounts() []common.Address {
	keyDir := path.Join(GetConfigDir(), EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)
	addresses := am.Accounts()

	return addresses
}

//func DeleteAccount(address string) bool {
//	keyDir := path.Join(GetConfigDir(), EthereumWalletStorageDir)
//	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
//	config := &accounts.Config{InsecureUnlockAllowed: false}
//	am := accounts.NewManager(config, ks)
//
//	wallet, err := am.Find(accounts.Account{
//		Address: common.HexToAddress(address),
//	})
//
//	if err != nil && wallet == nil {
//		fmt.Printf("failed to find account %s, error: %s", address, err)
//		return false
//	}
//}

func AccountExists(address string) bool {
	keyDir := path.Join(GetConfigDir(), EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)

	wallet, err := am.Find(accounts.Account{
		Address: common.HexToAddress(address),
	})

	if err != nil && wallet == nil {
		fmt.Printf("failed to find account %s, error: %s\n", address, err)
		return false
	}

	status, _ := wallet.Status()
	url := wallet.URL()

	fmt.Printf("Account exists. Status: %s, Path: %s\n", status, url)

	return true
}

// CreateKeyStorage create, restore or unlock key storage
func CreateKeyStorage(password string) error {
	keyDir := path.Join(GetConfigDir(), EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		return errors.Wrap(err, "failed to create keystore")
	}
	fmt.Printf("Created account: %s", account.Address.Hex())

	return nil
}

func UpdateClientEthereumAddress(address string) (err error) {
	configFile := path.Join(GetConfigDir(), BridgeClientConfigName)
	buf, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	cfg := &Bridge{}
	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		return err
	}

	cfg.Owner.EthereumAddress = address

	text, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, text, 0644)

	return err
}

func ImportAccount(homedir, mnemonic, password string) (string, error) {
	// 1. Create storage and account if it doesn't exist and add account to it

	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 2. Init wallet

	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", errors.Wrap(err, "failed to import from mnemonic")
	}

	pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(pathD, true)
	if err != nil {
		return "", errors.Wrap(err, "failed parse derivation path")
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		return "", errors.Wrap(err, "failed to get private key")
	}

	// 3. Find key

	acc, err := ks.Find(account)
	if err == nil {
		fmt.Printf("Account already exists: %s\nPath: %s\n\n", acc.Address.Hex(), acc.URL.Path)
		return acc.Address.Hex(), nil
	}

	// 4. Import the key if it doesn't exist

	acc, err = ks.ImportECDSA(key, password)
	if err != nil {
		return "", errors.Wrap(err, "failed to get import private key")
	}

	fmt.Printf("Imported account %s to path: %s\n", acc.Address.Hex(), acc.URL.Path)

	return acc.Address.Hex(), nil
}
