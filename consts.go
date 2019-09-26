package xlm

import (
	"log"
	"net/http"

	horizon "github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/network"
)

// constants that are imported by other packages

var (
	// Passphrase defines the stellar network passphrase
	Passphrase string
	// Mainnet is a bool which decides which chain to connect to
	IsMainNet bool
	// TestNetClient defines the horizon client to connect to
	HorizonNetClient *horizon.Client
)

// RefillAmount defines the default stellar refill amount
var RefillAmount float64

// SetConsts XLM consts
func SetupParams(amount float64, isMainnet bool) {
	RefillAmount = amount
	IsMainNet = isMainnet
	//log.Println("SETTING MAINNET TO: ", isMainnet)
	if IsMainNet {
		Passphrase = network.PublicNetworkPassphrase
		log.Println("Pointing horizon to mainnet")
		HorizonNetClient = &horizon.Client{
			HorizonURL: "https://horizon.stellar.org/", // switch to mainnet horizon
			HTTP:       http.DefaultClient,
		}
	} else {
		log.Println("Pointing horizon to testnet")
		Passphrase = network.TestNetworkPassphrase
		HorizonNetClient = &horizon.Client{
			HorizonURL: "https://horizon-testnet.stellar.org/",
			HTTP:       http.DefaultClient,
		}
	}
}
