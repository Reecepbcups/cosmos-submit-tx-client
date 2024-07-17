package main

import (
	"fmt"
	"time"

	cosmostx "cosmossdk.io/api/cosmos/tx/v1beta1"
	testutildata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	// https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json
	ChainId  = "local-1"
	GrpcAddr = "0.0.0.0:9090"
	keyName  = "myv50key"
	Denom    = "utest" // uatom
	// This is a test account on the cosmoshub with some uatom. You should use your own via an env variable such as `TestMnemonic := os.Getenv("MY_CLIENT_MNEMONIC")`
	// https://dev.mintscan.io/cosmos/account/cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5

	// cosmos12jz2v5psq5hu4px9ld23uztzek8y0xmzdxuheh
	TestMnemonic = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
)

func main() {
	// Setup the wallet signer basics and close the gRPC connection after the main function exits.
	w := SetupWalletSigner(GrpcAddr)
	defer w.GrpcConn.Close()

	// random account
	// Use `sdk.MustAccAddressFromBech32("cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5")` for specific accounts to test with
	_, _, addr1 := testutildata.KeyTestPubAddr()

	// Load our private keyring account from Mnemonic
	_, acc := w.LoadKeyFromMnemonic(keyName, TestMnemonic, "1234567890")

	// Create a bank send message from our account (acc) -> some other account address
	msg1 := banktypes.NewMsgSend(acc, addr1, sdk.NewCoins(sdk.NewInt64Coin(Denom, 1)))
	if err := w.TxBuilder.SetMsgs(msg1); err != nil {
		panic(err)
	}

	// Set the transaction information (do before signing the Tx)
	w.TxBuilder.SetGasLimit(100_000)
	w.TxBuilder.SetMemo("my test memo")
	w.TxBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(Denom, 750)))

	// Signs the transaction
	if err := w.SignTx(keyName); err != nil {
		panic(err)
	}

	// broadcast the Tx that has been signed
	broadcastRes := w.BroadcastTx()
	fmt.Printf("broadcastRes: %+v\n", broadcastRes)

	if broadcastRes.TxResponse.Code != 0 {
		panic(fmt.Errorf("tx failed: %s", broadcastRes.TxResponse.RawLog))
	}

	// Only `broadcastRes.TxResponse.Txhash` and `broadcastRes.TxResponse.RawLog` are returned.
	// If rawlog is filled with anything other than '[]', your transaction was not accepted into the chain at all. Re-submit with the required fix.
	// if only Txhash is set, query the TxHash every second until some data comes back. This is not elegant, but it is required :(
	txRes, err := w.PollForTxHash(broadcastRes.TxResponse.Txhash)
	if err != nil {
		panic(err)
	}

	fmt.Printf("txRes: %+v\n", txRes)

}

func (w *WalletSigner) PollForTxHash(txhash string) (sdk.TxResponse, error) {
	resp := sdk.TxResponse{}

	sc := cosmostx.NewServiceClient(w.GrpcConn)

	attemptSec := 10

	// try the following for attemptSec every 1 second. If an error is thrown but we are not yet to an i of attemptSec, keep trying.
	for i := 0; i < int(attemptSec); i++ {
		// You can also just use tx here and parse the data you need.
		tx, err := sc.GetTx(w.Ctx, &cosmostx.GetTxRequest{
			Hash: txhash,
		})

		if err != nil {
			fmt.Println("POLLING: error getting tx. Trying again... ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if i > int(attemptSec) {
			return resp, fmt.Errorf("POLLING: tx not found after %d seconds", int(attemptSec))
		}

		resp.Code = tx.TxResponse.Code
		resp.RawLog = tx.TxResponse.RawLog
		resp.TxHash = tx.TxResponse.Txhash
		resp.Data = tx.TxResponse.Data
		resp.GasUsed = tx.TxResponse.GasUsed
		resp.GasWanted = tx.TxResponse.GasWanted
		resp.Height = tx.TxResponse.Height
		resp.Info = tx.TxResponse.Info
		// resp.Logs = tx.TxResponse.Logs
		resp.Timestamp = tx.TxResponse.Timestamp

		return resp, nil

	}

	return resp, nil
}
