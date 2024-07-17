package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	cosmostx "cosmossdk.io/api/cosmos/tx/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	oldtxsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// WalletSigner is a struct that holds the basics for signing a transaction.
// This is not required, but is nice to have instead of passing through tons of method arguments.
type WalletSigner struct {
	Ctx       context.Context
	TxBuilder client.TxBuilder
	EncCfg    moduletestutil.TestEncodingConfig
	Keyring   keyring.Keyring
	GrpcConn  *grpc.ClientConn
}

// SetupWalletSigner sets up the wallet signer basics and returns a pointer to
// the WalletSigner struct.
func SetupWalletSigner(gRPCAddr string) *WalletSigner {
	// To sign a transaction, the AppModuleBasic must be provided here. This
	// is for the protobuf (so we can encode/decode the transaction bytes)
	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		consensus.AppModuleBasic{},
	)

	// Setup a gRPC connection
	grpcConn, err := grpc.Dial(
		gRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	// Setup a struct to share data with helper methods.
	w := &WalletSigner{
		Ctx:      context.Background(),
		GrpcConn: grpcConn,
		EncCfg:   encCfg,
		Keyring:  keyring.NewInMemory(encCfg.Codec),
	}

	// Set the TxBuilder to be empty
	w.ResetTxBuilder()

	return w
}

// ResetTxBuilder resets the TxBuilder to a new TxBuilder.
func (w *WalletSigner) ResetTxBuilder() {
	w.TxBuilder = w.EncCfg.TxConfig.NewTxBuilder()
}

// LoadKeyFromMnemonic loads a key from a mnemonic and returns the keyring record and the address.
func (w *WalletSigner) LoadKeyFromMnemonic(keyName, mnemonic, password string) (*keyring.Record, sdk.AccAddress) {
	path := sdk.GetConfig().GetFullBIP44Path()
	r, err := w.Keyring.NewAccount(keyName, mnemonic, password, path, hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	a, err := r.GetAddress()
	if err != nil {
		panic(err)
	}

	return r, a
}

// GetAccountInfo returns an SDK account from the chain.
// NOTE: the account must have at least 0.000001 tokens in it first to work.
func (w *WalletSigner) GetAccountInfo(addr sdk.AccAddress) (authtypes.BaseAccount, error) {
	var acc authtypes.BaseAccount

	// Query account info from the chain (required for sequence & account number)
	res, err := authtypes.NewQueryClient(w.GrpcConn).Account(w.Ctx, &authtypes.QueryAccountRequest{
		Address: addr.String(),
	})
	if err != nil {
		return acc, err
	}

	if err := w.EncCfg.Codec.Unmarshal(res.Account.Value, &acc); err != nil {
		return acc, err
	}

	return acc, nil
}

func (w *WalletSigner) BroadcastTx() *cosmostx.BroadcastTxResponse {
	// Generated Protobuf-encoded bytes.
	txBytes, err := w.EncCfg.TxConfig.TxEncoder()(w.TxBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	// JSON String (not required, just showing for reference)
	txBytesJson, err := w.EncCfg.TxConfig.TxJSONEncoder()(w.TxBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	fmt.Println("txBytesJson", string(txBytesJson))

	// Submit the Tx to the gRPC server
	txClient := cosmostx.NewServiceClient(w.GrpcConn)
	grpcRes, err := txClient.BroadcastTx(
		w.Ctx,
		&cosmostx.BroadcastTxRequest{
			Mode:    cosmostx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		panic(err)
	}

	return grpcRes
}

func (w *WalletSigner) SignTx(keyName string) error {
	k, err := w.Keyring.Key(keyName)
	if err != nil {
		return err
	}

	krAcc, err := k.GetAddress()
	if err != nil {
		return err
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		return err
	}

	// BaseAccount chain info
	acc, err := w.GetAccountInfo(krAcc)
	if err != nil {
		return err
	}

	// Get the base Tx bytes
	txBytes, err := w.EncCfg.TxConfig.TxEncoder()(w.TxBuilder.GetTx())
	if err != nil {
		return err
	}

	protoTxCfg := authtx.NewTxConfig(codec.NewProtoCodec(w.EncCfg.InterfaceRegistry), authtx.DefaultSignModes)

	// fee := w.TxBuilder.GetTx().GetFee()

	signMode := protoTxCfg.SignModeHandler().DefaultMode()
	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	// sig := oldtxsigning.SignatureV2{
	// 	PubKey: pubKey,
	// 	Data: &oldtxsigning.SingleSignatureData{
	// 		SignMode:  oldtxsigning.SignMode(signMode),
	// 		Signature: nil,
	// 	},
	// 	Sequence: acc.Sequence,
	// }

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	if err := w.TxBuilder.SetSignatures(oldtxsigning.SignatureV2{
		PubKey: pubKey,
		Data: &oldtxsigning.SingleSignatureData{
			SignMode:  oldtxsigning.SignMode(signMode),
			Signature: nil,
		},
		Sequence: acc.Sequence, // get sequence from query
	}); err != nil {
		panic(err)
	}

	anyPubKey, err := PubKeyToAny(pubKey)
	if err != nil {
		return err
	}

	// 2nd round: once all signer infos are set, every signer can sign.
	signerData := txsigning.SignerData{
		ChainID:       ChainId,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		PubKey:        anyPubKey,
	}

	theTx := w.TxBuilder.GetTx()
	adaptableTx, ok := theTx.(authsigning.V2AdaptableTx)
	if !ok {
		return fmt.Errorf("%T does not implement the authsigning.V2AdaptableTx interface", theTx)
	}
	txData := adaptableTx.GetSigningTxData()

	// sign with the keyring
	// txBytes, err := protoTxCfg.TxEncoder()(theTx)
	// if err != nil {
	// 	return err
	// }

	// print txBytes
	fmt.Println("txBytes", string(txBytes))

	_, _, err = w.Keyring.Sign(keyName, txBytes, oldtxsigning.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return err
	}
	// sig.Data.(*oldtxsigning.SingleSignatureData).Signature = sigBytesfirst

	signBytes, err := protoTxCfg.SignModeHandler().GetSignBytes(w.Ctx, signMode, signerData, txData)
	if err != nil {
		return err
	}

	// sign with the keyring
	sigBytes, _, err := w.Keyring.Sign(keyName, signBytes, oldtxsigning.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return err
	}
	// sig.Data.(*sdksigning.SingleSignatureData).Signature = sigBytes

	if err = w.TxBuilder.SetSignatures(oldtxsigning.SignatureV2{
		PubKey: pubKey,
		Data: &sdksigning.SingleSignatureData{
			SignMode:  sdksigning.SignMode(signMode),
			Signature: sigBytes,
		},
		Sequence: acc.Sequence,
	}); err != nil {
		return err
	}

	return nil
}

func PubKeyToAny(pubKey cryptotypes.PubKey) (*anypb.Any, error) {
	anyPk, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		TypeUrl: anyPk.TypeUrl,
		Value:   anyPk.Value,
	}, nil
}
