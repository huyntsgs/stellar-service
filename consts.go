package xlm

import (
	"log"
	"net/http"

	//b "github.com/stellar/go/build"
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
	HorizonClient *horizon.Client
	//Network       network.PublicNetworkPassphrase
	//Client horizonc.
)

// RefillAmount defines the default stellar refill amount
var RefillAmount float64

func GetHorizonClient() *horizon.Client {
	return HorizonClient
}

// SetConsts XLM consts
func SetupParams(amount float64, isMainnet bool) {
	RefillAmount = amount
	IsMainNet = isMainnet
	if IsMainNet {
		Passphrase = network.PublicNetworkPassphrase
		log.Println("Pointing horizon to mainnet")
		HorizonClient = &horizon.Client{
			HorizonURL: "https://horizon.stellar.org/",
			HTTP:       http.DefaultClient,
		}
		//Network = b.PublicNetwork
	} else {
		log.Println("Pointing horizon to testnet")
		Passphrase = network.TestNetworkPassphrase
		HorizonClient = &horizon.Client{
			HorizonURL: "https://horizon-testnet.stellar.org/",
			HTTP:       http.DefaultClient,
		}
		//Network = b.TestNetwork
	}
}
func SetupParam(amount float64, isMainnet bool, horizonUrl string) {
	RefillAmount = amount
	IsMainNet = isMainnet

	if IsMainNet {
		Passphrase = network.PublicNetworkPassphrase
		log.Println("Pointing horizon to mainnet")
		HorizonClient = &horizon.Client{
			HorizonURL: horizonUrl,
			HTTP:       http.DefaultClient,
		}
		//Network = b.PublicNetwork
	} else {
		log.Println("Pointing horizon to testnet")
		Passphrase = network.TestNetworkPassphrase
		HorizonClient = &horizon.Client{
			HorizonURL: "https://horizon-testnet.stellar.org/",
			HTTP:       http.DefaultClient,
		}
		//Network = b.TestNetwork
	}
}
