package token

import (
	"crypto"
	"log"
	"net/http"

	"github.com/sec51/twofactor"
)

type token struct {
	totp *twofactor.Totp
}

func New() *token {
	//todo save secret to disk and load from there (but where?)
	//todo multiple tokens so we can expire/rotate them?
	totp, err := twofactor.NewTOTP("friend", "wg-sync", crypto.SHA256, 6)
	if err != nil {
		panic(err) //can't generate otop token
	}
	//Todo have this also serve up a one time token also by returning a mux?
	return &token{totp: totp}

}

func (o *token) ServeHTTP(resp http.ResponseWriter, _ *http.Request) {
	qr, err := o.totp.QR()
	if err != nil {
		log.Printf("Error generating QR code: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-Type", "image/png")
	_, err = resp.Write(qr)
	if err != nil {
		log.Printf("Error generating QR code: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.WriteHeader(http.StatusOK)
}
