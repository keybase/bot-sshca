package shared

/*
chat_types.go includes the types used when kssh and keybaseca communicate over keybase chat. kssh starts by sending a
series of AckRequests in order to determine whether keybaseca is currently active and responding to messages. Keybaseca
responds to AckRequests with an AckResponse. Both messages contain the username of the user using kssh in order to
ensure that kssh is reading AckResponses that are meant for it (as opposed to another user of kssh). Then kssh sends
a SignatureRequest. This is a json object prefix with a specific string. The json object contains the ssh public key
and a uuid that is used to track the request. keybaseca responds with a signature response that contains the same uuid.
*/

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The body of signature request messages sent over KB chat
type SignatureRequest struct {
	SSHPublicKey       string `json:"ssh_public_key"`
	UUID               string `json:"uuid"`
	RequestedPrincipal string `json:"requested_principal,omitempty"`
	Username           string `json:"-"`
	DeviceName         string `json:"-"`
}

// The preamble used at the start of signature request messages
const SignatureRequestPreamble = "Signature_Request:"

// Parse the given string as a serialized SignatureRequest
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
const SignatureResponsePreamble = "Signature_Response:"

// Parse the given string as a serialized SignatureResponse
func ParseSignatureResponse(body string) (SignatureResponse, error) {
	if !strings.HasPrefix(body, SignatureResponsePreamble) {
		return SignatureResponse{}, fmt.Errorf("ParseSignatureResponse called on a body without a preamble")
	}

	body = strings.Replace(body, SignatureResponsePreamble, "", 1)
	var sr SignatureResponse
	err := json.Unmarshal([]byte(body), &sr)
	return sr, err
}

const AckRequestPrefix = "AckRequest--"

// Generate an AckRequest for the given username
func GenerateAckRequest(username string) string {
	return AckRequestPrefix + username
}

// Generate an AckResponse in response to the given ack request
func GenerateAckResponse(ackRequest string) string {
	return strings.Replace(ackRequest, "AckRequest", "Ack", 1)
}

// Returns whether the given message is an ack request
func IsAckRequest(msg string) bool {
	return strings.HasPrefix(msg, AckRequestPrefix)
}

// Returns whether the given message is an ack response
func IsAckResponse(msg string) bool {
	return strings.HasPrefix(msg, "Ack--")
}
