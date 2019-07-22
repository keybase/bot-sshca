package shared

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The body of signature request messages sent over KB chat
type SignatureRequest struct {
	SSHPublicKey string `json:"ssh_public_key"`
	UUID         string `json:"uuid"`
	Username     string `json:"-"`
}

// The preamble used at the start of signature request messages
var SignatureRequestPreamble = "Signature_Request:"

func ParseSignatureRequest(body string) (SignatureRequest, error) {
	if !strings.HasPrefix(body, SignatureRequestPreamble) {
		return SignatureRequest{}, fmt.Errorf("ParseSignatureRequest called on a body without a preamble")
	}

	body = strings.Replace(body, SignatureRequestPreamble, "", 1)
	var sr SignatureRequest
	err := json.Unmarshal([]byte(body), &sr)
	return sr, err
}

// The body of signature response messages sent over KB chat
type SignatureResponse struct {
	SignedKey string `json:"signed_key"`
	UUID      string `json:"uuid"`
}

// The preamble used at the start of signature response messages
var SignatureResponsePreamble = "Signature_Response:"

func ParseSignatureResponse(body string) (SignatureResponse, error) {
	if !strings.HasPrefix(body, SignatureResponsePreamble) {
		return SignatureResponse{}, fmt.Errorf("ParseSignatureResponse called on a body without a preamble")
	}

	body = strings.Replace(body, SignatureResponsePreamble, "", 1)
	var sr SignatureResponse
	err := json.Unmarshal([]byte(body), &sr)
	return sr, err
}

var AckRequest = "AckRequest"
var Ack = "Ack"
