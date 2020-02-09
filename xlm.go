package xlm

import (
	"log"
	"net/url"

	"github.com/pkg/errors"

	utils "github.com/Varunram/essentials/utils"
	horizon "github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	horizonprotocol "github.com/stellar/go/protocols/horizon"
	build "github.com/stellar/go/txnbuild"

	b "github.com/stellar/go/build"

	"github.com/stellar/go/xdr"
)

// xlm is a package with stellar related handlers which are useful for interacting with horizon

// Generating a keypair on stellar doesn't mean that you can send funds to it
// you need to call the CreateAccount method in project to be able to send funds
// to it

// GetKeyPair gets a keypair that can be used to interact with the stellar blockchain
func GetKeyPair() (string, string, error) {
	pair, err := keypair.Random()
	return pair.Seed(), pair.Address(), err
}

// AccountExists checks whether an account exists
func AccountExists(publicKey string) bool {
	x, err := ReturnSourceAccountPubkey(publicKey)
	log.Println("X=", x)
	if err != nil {
		// error in the horizon api call
		return false
	}
	return x.Sequence != "0" // if the sequence is zero, the account doesn't exist yet. This equals to the ledger number at which the account was created
}

func MergeAccount(sourcePk, dstPk, signerSeed, assetIssuer string) error {

	// sourceAccount, err := ReturnSourceAccountPubkey(sourcePk)
	// if err != nil {
	// 	return err
	// }

	// op := build.AccountMerge{
	// 	Destination: dstPk,
	// }

	//changeTrust := build.RemoveTrustlineOp(build.CreditAsset{Code: "GRX", Issuer: assetIssuer})
	// thh := build.Threshold(5)
	// setOp := build.SetOptions{
	// 	HighThreshold: &thh,
	// }

	// tx := build.Transaction{
	// 	SourceAccount: &sourceAccount,
	// 	Operations:    []build.Operation{&op, &removeTrust},
	// 	Timebounds:    build.NewInfiniteTimeout(),
	// 	Network:       Passphrase,
	// }

	// tx := build.Transaction{
	// 	SourceAccount: &sourceAccount,
	// 	Operations:    []build.Operation{&setOp},
	// 	Timebounds:    build.NewInfiniteTimeout(),
	// 	Network:       Passphrase,
	// }

	// //build.SetOptions
	// //signerSeed = "SCQNWWVTB5HE47CCUQV7PMRFWZNRPFTSOYHJ4HIRGN5HK5JH2WCBMUZA"
	// kp, err := keypair.Parse(signerSeed)
	// if err != nil {
	// 	return err
	// }

	// txe, err := tx.BuildSignEncode(kp.(*keypair.Full))
	// if err != nil {
	// 	return err
	// }

	// txres, err := HorizonClient.SubmitTransactionXDR(txe)
	// log.Println("SubmitTransaction xdr:", txe)
	// if err != nil {
	// 	log.Println("Set threshold err:", err, txres)
	// 	return err
	// } else {
	// 	log.Println("Set threshold success:")
	// }

	// Set ops
	// setOp = build.SetOptions{
	// 	Signer: &build.Signer{Address: sourcePk, Weight: build.Threshold(0)},
	// }
	// tx = build.Transaction{
	// 	SourceAccount: &sourceAccount,
	// 	Operations:    []build.Operation{&setOp},
	// 	Timebounds:    build.NewInfiniteTimeout(),
	// 	Network:       Passphrase,
	// }

	// //build.SetOptions
	// //signerSeed = "SCQNWWVTB5HE47CCUQV7PMRFWZNRPFTSOYHJ4HIRGN5HK5JH2WCBMUZA"
	// kp, err = keypair.Parse(signerSeed)
	// if err != nil {
	// 	return err
	// }

	// txe, err = tx.BuildSignEncode(kp.(*keypair.Full))
	// if err != nil {
	// 	return err
	// }

	// txres, err = HorizonClient.SubmitTransactionXDR(txe)
	// log.Println("Setops xdr:", txe)
	// if err != nil {
	// 	log.Println("Setops err:", err, txres)
	// 	return err
	// } else {
	// 	log.Println("Setops success")
	// }

	// // Merge
	// tx = build.Transaction{
	// 	SourceAccount: &sourceAccount,
	// 	Operations:    []build.Operation{&op},
	// 	Timebounds:    build.NewInfiniteTimeout(),
	// 	Network:       Passphrase,
	// }

	// //build.SetOptions
	// kp, err = keypair.Parse(signerSeed)
	// if err != nil {
	// 	return err
	// }

	// txe, err = tx.BuildSignEncode(kp.(*keypair.Full))
	// if err != nil {
	// 	return err
	// }

	// txres, err = HorizonClient.SubmitTransactionXDR(txe)
	// log.Println("SubmitTransaction xdr:", txe)
	// if err != nil {
	// 	log.Println("Merge account err:", err, txres)
	// 	return err
	// }

	return nil

}

// SendTx signs and broadcasts a given stellar tx
func SendTx(mykp keypair.KP, tx *build.Transaction) (int32, string, error) {

	txe, err := tx.BuildSignEncode(mykp.(*keypair.Full))

	if err != nil {
		return -1, "", errors.Wrap(err, "could not build/sign/encode")
	}

	log.Println("SendTx-xdr:", txe)

	resp, err := HorizonClient.SubmitTransactionXDR(txe)
	if err != nil {
		log.Println("SubmitTransactionXDR err:", err)
		return -1, "", errors.Wrap(err, "could not submit tx to horizon")
	}

	log.Printf("Propagated Transaction: %s, sequence: %d\n", resp.Hash, resp.Ledger)
	return resp.Ledger, resp.Hash, nil
}

// ParseXDR parse xdr to transacation and check whether sourceAccount is valid
// and then sign transaction with the signer key
func ParseXDR(xdr, sourceAccount, secretKey string) (txresp horizonprotocol.TransactionSuccess, e error, txcode string) {
	//var e error

	txcode = "tx_invalid"
	txn := decodeFromBase64(xdr)
	if txn.E.Tx.SourceAccount.Address() != sourceAccount {
		txcode = "tx_invalid_source_account"
		return txresp, errors.New("Invalid public key"), txcode
	}
	// 4. check the source account and mutate the transaction inside the transaction envelope if needed:
	//     a. update the source account
	//     b. set the sequence number
	//     c. set the network passphrase
	//	horizonClient := horizon.DefaultTestNetClient
	// if txn.E.Tx.SourceAccount.Address() == "" {
	// 	e = txn.MutateTX(
	// 		// we assume that the accountID uses the master key, this can also be the accountID
	// 		&b.SourceAccount{AddressOrSeed: secretKey},
	// 		//&b.AutoSequence{SequenceProvider: HorizonClient},
	// 		// need to reset the network passphrase
	// 		Passphrase,
	// 	)
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// } else if txn.E.Tx.SeqNum == 0 {
	// e = txn.MutateTX(
	// 	// do not need to set the source account here, only the sequence number
	// 	&b.AutoSequence{SequenceProvider: HorizonClient},
	// 	// need to reset the network passphrase
	// 	Passphrase,
	// )
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// }

	e = txn.MutateTX(
		Network,
	)
	if e != nil {
		log.Fatal(e)
	}

	// 5. sign the transaction envelope
	e = txn.Mutate(&b.Sign{Seed: secretKey})
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}
	//	log.Println("txn:", txn.E.Tx.Operations[0].Body.ManageBuyOfferOp.Price)

	// 6. convert the transaction to base64
	reencodedTxnBase64, e := txn.Base64()
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}

	//	log.Println("reencodedTxnBase64:", reencodedTxnBase64)

	// 7. submit to the network
	txresp, e = HorizonClient.SubmitTransactionXDR(reencodedTxnBase64)
	if e != nil {
		hError := e.(*horizon.Error)
		code, err := hError.ResultCodes()
		if err == nil {
			txcode = code.TransactionCode
			return txresp, e, txcode
		} else {
			log.Println("Error submitting transaction:", code.TransactionCode, code.OperationCodes)
			return txresp, e, txcode
		}
	}
	txcode = "tx_success"
	return txresp, nil, txcode
}

// ParseXDR parse xdr to transacation and check whether sourceAccount is valid
// and then sign transaction with the signer key
func ParseLoanXDR(xdrData, sourceAccount, secretKey, destPublickey string, amount float64) (txresp horizonprotocol.TransactionSuccess, e error, txcode string) {
	//var e error

	txcode = "tx_invalid"
	txn := decodeFromBase64(xdrData)
	if txn.E.Tx.SourceAccount.Address() != sourceAccount {
		txcode = "tx_invalid_source_account"
		return txresp, errors.New("Invalid public key"), "invalid public key"
	}

	// Destination AccountId
	// Asset       Asset
	// Amount      Int64
	txamount := int64(txn.E.Tx.Operations[0].Body.PaymentOp.Amount)
	log.Println("txamount:", txamount)
	if txamount < int64(amount*float64(1000000)) {
		txcode = "invalid_amount"
		return txresp, errors.New("invalid amount"), txcode
	}
	//log.Println("txn.E.Tx.Operations:", txn.E.Tx.Operations)
	//dest := txn.E.Tx.Operations[0].Body.PaymentOp.Destination
	if len(txn.E.Tx.Operations) > 0 {
		if txn.E.Tx.Operations[0].Body.PaymentOp.Destination.Address() != destPublickey {
			txcode = "invalid_dest_addr"
			return txresp, errors.New("invalid destination address"), txcode
		}
		// if txn.E.Tx.Operations[0].Body.SetOptionsOp.HighThreshold != xdr.Uint32{0} {
		// 	txcode = "invalid_setops"
		// 	return txresp, errors.New("invalid set ops"), txcode
		// }
		//log.Println("txn.E.Tx.Operations[1]:", txn.E.Tx.Operations[1])
	} else {
		txcode = "invalid_payment"
		return txresp, errors.New("invalid payment"), txcode
	}
	// 4. check the source account and mutate the transaction inside the transaction envelope if needed:
	//     a. update the source account
	//     b. set the sequence number
	//     c. set the network passphrase
	//	horizonClient := horizon.DefaultTestNetClient
	// if txn.E.Tx.SourceAccount.Address() == "" {
	// 	e = txn.MutateTX(
	// 		// we assume that the accountID uses the master key, this can also be the accountID
	// 		&b.SourceAccount{AddressOrSeed: secretKey},
	// 		//&b.AutoSequence{SequenceProvider: HorizonClient},
	// 		// need to reset the network passphrase
	// 		Passphrase,
	// 	)
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// } else if txn.E.Tx.SeqNum == 0 {
	// 	e = txn.MutateTX(
	// 		// do not need to set the source account here, only the sequence number
	// 		//&b.AutoSequence{SequenceProvider: HorizonClient},
	// 		// need to reset the network passphrase
	// 		Passphrase,
	// 	)
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// }

	e = txn.MutateTX(
		Network,
	)
	if e != nil {
		log.Fatal(e)
	}

	// 5. sign the transaction envelope
	// type SetOptionsOp struct {
	// 	InflationDest *AccountId
	// 	ClearFlags    *Uint32
	// 	SetFlags      *Uint32
	// 	MasterWeight  *Uint32
	// 	LowThreshold  *Uint32
	// 	MedThreshold  *Uint32
	// 	HighThreshold *Uint32
	// 	HomeDomain    *String32
	// 	Signer        *Signer
	// }
	// so := xdr.SetOptionsOp{
	// 	HighThreshold: xdr.Uint32{0},
	// 	Signer:        &xdr.AddSigner(destPublickey, 0),
	// }
	//e = txn.MutateTX(&b.SetOptionsBuilder{SO: so})

	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}
	e = txn.Mutate(&b.Sign{Seed: secretKey})
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}
	// setOp := b.SetOptions{
	// 	Signer: &build.Signer{Address: destPublickey, Weight: build.Threshold(0)},
	// }
	// e = txn.MutateTX(&setOp)

	//	log.Println("txn:", txn.E.Tx.Operations[0].Body.ManageBuyOfferOp.Price)

	// 6. convert the transaction to base64
	reencodedTxnBase64, e := txn.Base64()
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}

	//	log.Println("reencodedTxnBase64:", reencodedTxnBase64)

	// 7. submit to the network
	txresp, e = HorizonClient.SubmitTransactionXDR(reencodedTxnBase64)
	if e != nil {
		hError := e.(*horizon.Error)
		code, err := hError.ResultCodes()
		if err == nil {
			txcode = code.TransactionCode
			return txresp, e, txcode
		} else {
			log.Println("Error submitting transaction:", code.TransactionCode, code.OperationCodes)
			return txresp, e, txcode
		}
	}
	txcode = "tx_success"
	return txresp, nil, txcode
}

// ParseXDR parse xdr to transacation and check whether sourceAccount is valid
// and then sign transaction with the signer key
func ParseLoanXDR1(xdrData, sourceAccount, secretKey, destPublickey string, amount float64) (txresp horizonprotocol.TransactionSuccess, e error, txcode string) {
	//var e error

	txcode = "tx_invalid"
	txn := decodeFromBase64(xdrData)
	if txn.E.Tx.SourceAccount.Address() != sourceAccount {
		txcode = "tx_invalid_source_account"
		return txresp, errors.New("Invalid public key"), "invalid public key"
	}

	// Destination AccountId
	// Asset       Asset
	// Amount      Int64
	txamount := int64(txn.E.Tx.Operations[0].Body.PaymentOp.Amount)
	log.Println("txamount:", txamount)
	if txamount < int64(amount*float64(1000000)) {
		txcode = "invalid_amount"
		return txresp, errors.New("invalid amount"), txcode
	}
	//log.Println("txn.E.Tx.Operations:", txn.E.Tx.Operations)
	//dest := txn.E.Tx.Operations[0].Body.PaymentOp.Destination
	if len(txn.E.Tx.Operations) > 0 {
		if txn.E.Tx.Operations[0].Body.PaymentOp.Destination.Address() != destPublickey {
			txcode = "invalid_dest_addr"
			return txresp, errors.New("invalid destination address"), txcode
		}
		// if txn.E.Tx.Operations[0].Body.SetOptionsOp.HighThreshold != xdr.Uint32{0} {
		// 	txcode = "invalid_setops"
		// 	return txresp, errors.New("invalid set ops"), txcode
		// }
		//log.Println("txn.E.Tx.Operations[1]:", txn.E.Tx.Operations[1])
	} else {
		txcode = "invalid_payment"
		return txresp, errors.New("invalid payment"), txcode
	}
	// 4. check the source account and mutate the transaction inside the transaction envelope if needed:
	//     a. update the source account
	//     b. set the sequence number
	//     c. set the network passphrase
	//	horizonClient := horizon.DefaultTestNetClient
	// if txn.E.Tx.SourceAccount.Address() == "" {
	// 	e = txn.MutateTX(
	// 		// we assume that the accountID uses the master key, this can also be the accountID
	// 		&b.SourceAccount{AddressOrSeed: secretKey},
	// 		//&b.AutoSequence{SequenceProvider: HorizonClient},
	// 		// need to reset the network passphrase
	// 		Passphrase,
	// 	)
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// } else if txn.E.Tx.SeqNum == 0 {
	// 	e = txn.MutateTX(
	// 		// do not need to set the source account here, only the sequence number
	// 		//&b.AutoSequence{SequenceProvider: HorizonClient},
	// 		// need to reset the network passphrase
	// 		Passphrase,
	// 	)
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// }

	e = txn.MutateTX(
		Network,
	)
	if e != nil {
		log.Fatal(e)
	}

	// 5. sign the transaction envelope
	// type SetOptionsOp struct {
	// 	InflationDest *AccountId
	// 	ClearFlags    *Uint32
	// 	SetFlags      *Uint32
	// 	MasterWeight  *Uint32
	// 	LowThreshold  *Uint32
	// 	MedThreshold  *Uint32
	// 	HighThreshold *Uint32
	// 	HomeDomain    *String32
	// 	Signer        *Signer
	// }
	// so := xdr.SetOptionsOp{
	// 	HighThreshold: xdr.Uint32{0},
	// 	Signer:        &xdr.AddSigner(destPublickey, 0),
	// }
	//e = txn.MutateTX(&b.SetOptionsBuilder{SO: so})

	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}
	e = txn.Mutate(&b.Sign{Seed: secretKey})
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}

	// 6. convert the transaction to base64
	reencodedTxnBase64, e := txn.Base64()
	if e != nil {
		log.Println(e)
		return txresp, e, txcode
	}

	//	log.Println("reencodedTxnBase64:", reencodedTxnBase64)

	// 7. submit to the network
	txresp, e = HorizonClient.SubmitTransactionXDR(reencodedTxnBase64)
	if e != nil {
		hError := e.(*horizon.Error)
		code, err := hError.ResultCodes()
		if err == nil {
			txcode = code.TransactionCode
			return txresp, e, txcode
		} else {
			log.Println("Error submitting transaction:", code.TransactionCode, code.OperationCodes)
			return txresp, e, txcode
		}
	}
	txcode = "tx_success"
	return txresp, nil, txcode
}

// unescape decodes the URL-encoded and base64 encoded txn
func unescape(escaped string) string {
	unescaped, e := url.QueryUnescape(escaped)
	if e != nil {
		log.Fatal(e)
	}
	return unescaped
}

// decodeFromBase64 decodes the transaction from a base64 string into a TransactionEnvelopeBuilder
func decodeFromBase64(encodedXdr string) *b.TransactionEnvelopeBuilder {
	// Unmarshall from base64 encoded XDR format
	var decoded xdr.TransactionEnvelope
	e := xdr.SafeUnmarshalBase64(encodedXdr, &decoded)
	if e != nil {
		log.Fatal(e)
	}

	// convert to TransactionEnvelopeBuilder
	txEnvelopeBuilder := b.TransactionEnvelopeBuilder{E: &decoded}
	txEnvelopeBuilder.Init()

	return &txEnvelopeBuilder
}

// SendXLMCreateAccount creates and sends XLM to a new account
func SendXLMCreateAccount(destination string, amountx float64, seed string) (int32, string, error) {
	// don't check if the account exists or not, hopefully it does
	sourceAccount, mykp, err := ReturnSourceAccount(seed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not get source account of seed")
	}

	amount, err := utils.ToString(amountx)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not convert amount to string")
	}

	op := build.CreateAccount{
		Destination: destination,
		Amount:      amount,
	}

	tx := &build.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []build.Operation{&op},
		Timebounds:    build.NewInfiniteTimeout(),
		Network:       Passphrase,
	}

	return SendTx(mykp, tx)
}

// ReturnSourceAccount returns the source account of the seed
func ReturnSourceAccount(seed string) (horizonprotocol.Account, keypair.KP, error) {
	var sourceAccount horizonprotocol.Account
	mykp, err := keypair.Parse(seed)
	if err != nil {
		return sourceAccount, mykp, errors.Wrap(err, "could not parse keypair, quitting")
	}
	client := getHorizonClient(IsMainNet)
	ar := horizon.AccountRequest{AccountID: mykp.Address()}
	sourceAccount, err = client.AccountDetail(ar)
	if err != nil {
		log.Println(err)
		return sourceAccount, mykp, errors.Wrap(err, "could not load client details, quitting")
	}

	return sourceAccount, mykp, nil
}

// ReturnSourceAccountPubkey returns the source account of the pubkey
func ReturnSourceAccountPubkey(pubkey string) (horizonprotocol.Account, error) {
	client := getHorizonClient(IsMainNet)
	ar := horizon.AccountRequest{AccountID: pubkey}
	sourceAccount, err := client.AccountDetail(ar)
	if err != nil {
		return sourceAccount, errors.Wrap(err, "could not load client details, quitting")
	}

	return sourceAccount, nil
}

// SendXLM sends xlm to a destination address
func SendXLM(destination string, amountx float64, seed string, memo string) (int32, string, error) {
	// don't check if the account exists or not, hopefully it does
	sourceAccount, mykp, err := ReturnSourceAccount(seed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	amount, err := utils.ToString(amountx)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not convert amount to string")
	}

	op := build.Payment{
		Destination:   destination,
		Amount:        amount,
		Asset:         build.NativeAsset{},
		SourceAccount: &sourceAccount,
	}

	tx := &build.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []build.Operation{&op},
		Timebounds:    build.NewInfiniteTimeout(),
		Network:       Passphrase,
		Memo:          build.Memo(build.MemoText(memo)),
	}
	//log.Println("Build tx")

	return SendTx(mykp, tx)
}

// SendAsset sends asset to a destination address
func SendAsset(destination string, amountx float64, seed string, asset build.Asset, memo string) (int32, string, error) {
	// don't check if the account exists or not, hopefully it does
	sourceAccount, mykp, err := ReturnSourceAccount(seed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	amount, err := utils.ToString(amountx)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not convert amount to string")
	}

	op := build.Payment{
		Destination:   destination,
		Amount:        amount,
		Asset:         asset,
		SourceAccount: &sourceAccount,
	}

	tx := &build.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []build.Operation{&op},
		Timebounds:    build.NewInfiniteTimeout(),
		Network:       Passphrase,
		Memo:          build.Memo(build.MemoText(memo)),
	}
	//log.Println("Build tx")

	return SendTx(mykp, tx)
}

func RemoveSigner(sourceAcc string, loanAccSeed string) (int32, string, error) {

	loanAcc, mykp, err := ReturnSourceAccount(loanAccSeed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	client := getHorizonClient(IsMainNet)
	ar := horizon.AccountRequest{AccountID: mykp.Address()}
	sourceAccount, err := client.AccountDetail(ar)

	op := build.SetOptions{
		Signer: &build.Signer{Address: loanAcc.GetAccountID(), Weight: build.Threshold(0)},
	}

	tx := &build.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []build.Operation{&op},
		Timebounds:    build.NewTimeout(180),
		Network:       Passphrase,
		Memo:          build.Memo(build.MemoText("")),
	}
	//log.Println("Build tx")

	return SendTx(mykp, tx)
}

func getHorizonClient(isMainNet bool) *horizon.Client {
	if isMainNet {
		return horizon.DefaultPublicNetClient
	} else {
		return horizon.DefaultTestNetClient
	}
}

// RefillAccount refills an account
func RefillAccount(publicKey string, refillSeed string) error {
	if IsMainNet {
		return errors.New("can't give free xlm on mainnet, quitting")
	}
	var err error
	if !AccountExists(publicKey) {
		// there is no account under the user's name
		// means we need to setup an account first
		log.Println("Account does not exist, creating: ", publicKey)
		_, _, err = SendXLMCreateAccount(publicKey, RefillAmount, refillSeed)
		if err != nil {
			log.Println("Account Could not be created")
			return errors.Wrap(err, "Account Could not be created")
		}
	}
	// balance is in string, convert to float
	balance, err := GetNativeBalance(publicKey)
	if err != nil {
		return errors.Wrap(err, "could not get native balance")
	}
	balanceI, _ := utils.ToFloat(balance)
	if balanceI < 3 { // to setup trustlines
		_, _, err = SendXLM(publicKey, RefillAmount, refillSeed, "Sending XLM to refill")
		if err != nil {
			return errors.Wrap(err, "Account doesn't have funds or invalid seed")
		}
	}
	return nil
}
