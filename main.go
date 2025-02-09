package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// Example of usage:
func main() {
	client := liteclient.NewConnectionPool()

	cfg, err := getLocalConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to add connections: %v", err)
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	ctx := client.StickyContext(context.Background())
	err = BuyTokens(ctx, api, []string{"mnemonic"}, 1, "amount Ton", "Jetton master Address", wallet.V3R2) // 1 = amount Jetton, "amount Ton" - amount TON, wallet version - wallet.V3R2
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

}

//
//Rest of code
//

// CellGenetator - generate cell for buy tokens
func CellGenetator(addressWallet string, amount uint64) *cell.Cell {
	// Current time + 300 seconds (5 minutes) for expiration
	expirationTime := uint64(time.Now().Unix() + 300)

	body := cell.BeginCell().
		MustStoreUInt(0xaf750d34, 32).    // op code
		MustStoreUInt(rand.Uint64(), 64). // query id (random 64-bit number)
		MustStoreCoins(amount).           // amount in nanotons
		MustStoreUInt(0, 1).
		MustStoreAddr(address.MustParseAddr(addressWallet)).
		MustStoreUInt(expirationTime, 32).
		EndCell()

	return body
}

// getLocalConfig - get config from file (global.config.json for example) - it's defolt Ton Blockchain config
func getLocalConfig(configPath string) (*liteclient.GlobalConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	cfg, err := liteclient.GetConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// BuyTokens - Function for buy tokens
func BuyTokens(ctx context.Context, api wallet.TonAPI, seed []string, amountJetton uint64, amount string, dest string, walletVersion wallet.Version) error {
	w, err := wallet.FromSeed(api, seed, walletVersion)
	if err != nil {
		return err
	}
	message := &wallet.Message{
		Mode: 1 + 2,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     address.MustParseAddr(dest),
			Amount:      tlb.MustFromTON(amount),
			Body:        CellGenetator(w.WalletAddress().String(), amountJetton),
		},
	}
	_, _, err = w.SendWaitTransaction(ctx, message)
	if err != nil {
		return err
	}
	return nil
}
