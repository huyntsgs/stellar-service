package xlm

import (
	"log"
	//	"net/url"

	"github.com/pkg/errors"

	utils "github.com/Varunram/essentials/utils"
	horizon "github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	horizonprotocol "github.com/stellar/go/protocols/horizon"
	build "github.com/stellar/go/txnbuild"

	//b "github.com/stellar/go/build"
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
	if err != nil {
		// error in the horizon api call
		return false
	}
	return x.Sequence != "0" // if the sequence is zero, the account doesn't exist yet. This equals to the ledger number at which the account was created
}
func CheckEnabledTrustLineAsset(publicKey, asset string) (bool, error) {
	x, err := ReturnSourceAccountPubkey(publicKey)
	if err != nil {
		// error in the horizon api call
		return false, err
	}
	for _, bl := range x.Balances {
		if bl.Asset.Code == asset {
			return true, nil
		}
	}
	return false, nil
}
func MergeAccount(sourceAcc string, loanAccSeed string, asset build.CreditAsset) (int32, string, error) {

	loanAcc, mykp, err := ReturnSourceAccount(loanAccSeed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	mergedAcc := &build.SimpleAccount{AccountID: sourceAcc}

	op := &build.ChangeTrust{
		Line:          asset,
		Limit:         "0",
		SourceAccount: mergedAcc,
	}

	op1 := &build.AccountMerge{
		Destination:   loanAcc.AccountID,
		SourceAccount: mergedAcc,
	}
	tx, err := build.NewTransaction(
		build.TransactionParams{
			SourceAccount:        &loanAcc,
			Operations:           []build.Operation{op, op1},
			Timebounds:           build.NewInfiniteTimeout(),
			IncrementSequenceNum: true,
			BaseFee:              build.MinBaseFee,
			Memo:                 build.Memo(build.MemoText("merged account")),
		})

	return SendTx(mykp, tx)
}

func MergeAccountNormal(destAcc string, srcAccSeed string, asset build.CreditAsset) (int32, string, error) {

	srcAcc, mykp, err := ReturnSourceAccount(srcAccSeed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	srcAccSimple := &build.SimpleAccount{AccountID: srcAcc.AccountID}

	op := &build.ChangeTrust{
		Line:          asset,
		Limit:         "0",
		SourceAccount: srcAccSimple,
	}

	op1 := &build.AccountMerge{
		Destination:   destAcc,
		SourceAccount: srcAccSimple,
	}
	tx, err := build.NewTransaction(
		build.TransactionParams{
			SourceAccount:        &srcAcc,
			Operations:           []build.Operation{op, op1},
			Timebounds:           build.NewInfiniteTimeout(),
			IncrementSequenceNum: true,
			BaseFee:              build.MinBaseFee,
			Memo:                 build.Memo(build.MemoText("merged account")),
		})

	return SendTx(mykp, tx)
}
func MergeAccountNChangeTrust(sourceAcc string, loanAccSeed string) (int32, string, error) {

	loanAcc, mykp, err := ReturnSourceAccount(loanAccSeed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	mergedAcc := &build.SimpleAccount{AccountID: sourceAcc}

	op1 := &build.AccountMerge{
		Destination:   loanAcc.AccountID,
		SourceAccount: mergedAcc,
	}
	tx, err := build.NewTransaction(
		build.TransactionParams{
			SourceAccount:        &loanAcc,
			Operations:           []build.Operation{op1},
			Timebounds:           build.NewInfiniteTimeout(),
			IncrementSequenceNum: true,
			BaseFee:              build.MinBaseFee,
			Memo:                 build.Memo(build.MemoText("merged account")),
		})

	return SendTx(mykp, tx)
}
func PayLoan(sourceAcc string, loanAccSeed string) (int32, string, error) {

	loanAcc, mykp, err := ReturnSourceAccount(loanAccSeed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not return source account")
	}

	userAcc := &build.SimpleAccount{AccountID: sourceAcc}

	// Pay loan
	op := &build.Payment{
		Destination:   loanAcc.AccountID,
		Amount:        "2.1",
		Asset:         build.NativeAsset{},
		SourceAccount: userAcc,
	}
	// set threshold, remove signer
	h := build.Threshold(0)
	m := build.Threshold(0)
	op1 := &build.SetOptions{
		HighThreshold:   &h,
		MediumThreshold: &m,
		Signer:          &build.Signer{Address: loanAcc.AccountID, Weight: build.Threshold(0)},
		SourceAccount:   userAcc,
	}
	tx, err := build.NewTransaction(build.TransactionParams{
		SourceAccount:        &loanAcc,
		Operations:           []build.Operation{op, op1},
		Timebounds:           build.NewInfiniteTimeout(),
		IncrementSequenceNum: true,
		BaseFee:              build.MinBaseFee,
		Memo:                 build.Memo(build.MemoText("loan paid for GrayLL")),
	})

	return SendTx(mykp, tx)
}

func ParseXDR(xdrData string) (xdr.TransactionEnvelope, error) {

	var txe xdr.TransactionEnvelope

	txb, err := build.TransactionFromXDR(xdrData)
	if err != nil {
		return txe, err
	}
	tx, _ := txb.Transaction()

	txe = tx.ToXDR()
	return txe, nil
}

// ParseXDR parse xdr to transacation and check whether sourceAccount is valid
// and then sign transaction with the signer key
func ParseLoanXDR(xdrData, sourceAccount, secretKey, destPublickey string, amount float64) (string, error) {
	//var e error

	txcode := "tx_valid"

	txb, err := build.TransactionFromXDR(xdrData)
	if err != nil {
		log.Println("ParseLoanXDR- TransactionFromXDR error ", err)
		return "invalid public key", errors.New("Invalid public key")
	}
	tx, _ := txb.Transaction()

	if tx.SourceAccount().AccountID != sourceAccount {
		txcode = "tx_invalid_source_account"
		return "invalid public key", errors.New("Invalid public key")
	}

	txe := tx.ToXDR()

	if txe.Operations()[0].Body.PaymentOp.Amount != xdr.Int64(amount*float64(1000000)) {
		txcode = "invalid_amount"
		return txcode, errors.New("invalid amount")
	}

	if txe.Operations()[0].Body.PaymentOp.Destination.GoString() != destPublickey {
		txcode = "invalid_dest_addr"
		return txcode, errors.New("invalid destination address")
	}

	_, kp, err := ReturnSourceAccount(secretKey)

	tx.Sign(Passphrase, kp)

	_, err = HorizonClient.SubmitTransaction(tx)

	return txcode, err

}

func newKeypair(seed string) *keypair.Full {
	myKeypair, _ := keypair.Parse(seed)
	return myKeypair.(*keypair.Full)
}

// SendTx signs and broadcasts a given stellar tx
func SendTx(mykp *keypair.Full, tx *build.Transaction) (int32, string, error) {

	var err error
	tx, err = tx.Sign(Passphrase, mykp)

	if err != nil {
		return -1, "", errors.Wrap(err, "could not build/sign/encode")
	}
	xdrString, _ := tx.Base64()
	log.Println("SendTx - xdr:", xdrString)

	resp, err := HorizonClient.SubmitTransactionXDR(xdrString)
	if err != nil {
		log.Println("SendTx - SubmitTransactionXDR err:", err, resp.ResultXdr, resp.ResultMetaXdr)
		return -1, "", errors.Wrap(err, "could not submit tx to horizon")
	}

	return resp.Ledger, resp.Hash, nil
}

//SendXLMCreateAccount creates and sends XLM to a new account
func SendXLMCreateAccount(destination string, amountx float64, seed string) (int32, string, error) {
	// don't check if the account exists or not, hopefully it does
	sourceAccount, mykp, err := ReturnSourceAccount(seed)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not get source account of seed")
	}
	//mykp := newKeypair(seed)
	//sourceAccount := build.NewSimpleAccount(mykp.Address(), 0)
	amount, err := utils.ToString(amountx)
	if err != nil {
		return -1, "", errors.Wrap(err, "could not convert amount to string")
	}

	op := build.CreateAccount{
		Destination: destination,
		Amount:      amount,
	}

	tx, err := build.NewTransaction(
		build.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []build.Operation{&op},
			BaseFee:              build.MinBaseFee,
			Timebounds:           build.NewInfiniteTimeout(),
		},
	)

	return SendTx(mykp, tx)
}

// ReturnSourceAccount returns the source account of the seed
func ReturnSourceAccount(seed string) (horizonprotocol.Account, *keypair.Full, error) {
	var sourceAccount horizonprotocol.Account
	mykp, err := keypair.Parse(seed)
	if err != nil {
		return sourceAccount, nil, errors.Wrap(err, "could not parse keypair, quitting")
	}

	client := getHorizonClient(IsMainNet)
	ar := horizon.AccountRequest{AccountID: mykp.Address()}
	sourceAccount, err = client.AccountDetail(ar)
	if err != nil {
		log.Println(err)
		return sourceAccount, nil, errors.Wrap(err, "could not load client details, quitting")
	}

	return sourceAccount, mykp.(*keypair.Full), nil
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
	tx, err := build.NewTransaction(build.TransactionParams{
		SourceAccount:        &sourceAccount,
		Operations:           []build.Operation{&op},
		Timebounds:           build.NewInfiniteTimeout(),
		BaseFee:              build.MinBaseFee,
		IncrementSequenceNum: true,
		Memo:                 build.Memo(build.MemoText(memo)),
	})

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

	// tx := &build.Transaction{
	// 	SourceAccount: &sourceAccount,
	// 	Operations:    []build.Operation{&op},
	// 	Timebounds:    build.NewInfiniteTimeout(),
	// 	Network:       Passphrase,
	// 	Memo:          build.Memo(build.MemoText(memo)),
	// }

	tx, err := build.NewTransaction(build.TransactionParams{
		SourceAccount:        &sourceAccount,
		Operations:           []build.Operation{&op},
		Timebounds:           build.NewInfiniteTimeout(),
		BaseFee:              build.MinBaseFee,
		IncrementSequenceNum: true,
		Memo:                 build.Memo(build.MemoText(memo)),
	})

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
	tx, err := build.NewTransaction(build.TransactionParams{
		SourceAccount:        &sourceAccount,
		Operations:           []build.Operation{&op},
		Timebounds:           build.NewInfiniteTimeout(),
		IncrementSequenceNum: true,
		Memo:                 build.Memo(build.MemoText("")),
	})
	// tx := &build.Transaction{
	// 	SourceAccount: &loanAcc,
	// 	Operations:    []txbuild.Operation{&op},
	// 	Timebounds:    txbuild.NewInfiniteTimeout(),
	// 	//Network:       Passphrase,
	// 	Memo: build.Memo(txbuild.MemoText("")),
	// }
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

func newSignedTransaction(
	params build.TransactionParams,
	network string,
	keypairs ...*keypair.Full,
) (string, error) {
	tx, err := build.NewTransaction(params)
	if err != nil {
		return "", errors.Wrap(err, "couldn't create transaction")
	}

	tx, err = tx.Sign(network, keypairs...)
	if err != nil {
		return "", errors.Wrap(err, "couldn't sign transaction")
	}

	txeBase64, err := tx.Base64()
	if err != nil {
		return "", errors.Wrap(err, "couldn't encode transaction")
	}

	return txeBase64, err
}
