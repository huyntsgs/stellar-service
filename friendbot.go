package xlm

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// FundXLMTestNet makes an API call to the stellar friendbot, which gives 10000 testnet XLM
func FundXLMTestNet(PublicKey string) error {
	if IsMainNet {
		return errors.New("no friendbot on mainnet, quitting")
	}
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + PublicKey)
	if err != nil || resp.Status != "200 OK" {
		log.Println("GetXLM:resp: ", resp)
		log.Println("GetXLM:err: ", err)
		return err
	}
	return nil
}
