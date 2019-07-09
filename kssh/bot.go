package kssh

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keybase/bot-ssh-ca/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

// TODO: Add a timeout or something here
func GetSignedKey(config ConfigFile, request shared.SignatureRequest) (shared.SignatureResponse, error) {
	empty := shared.SignatureResponse{}

	runOptions := kbchat.RunOptions{KeybaseLocation: "keybase"}
	kbc, err := kbchat.Start(runOptions)
	if err != nil {
		return empty, fmt.Errorf("error starting Keybase chat: %v", err)
	}

	sub, err := kbc.ListenForNewTextMessages()
	if err != nil {
		return empty, fmt.Errorf("error subscribing to messages: %v", err)
	}

	// If we just send our signature request to chat, we hit a race condition where if the CA responds fast enough
	// we will miss the response from the CA. We fix this with a simple ACKing algorithm:
	// 1. Send an AckRequest every 100ms until an Ack is received.
	// 2. Once an Ack is received, we know we are correctly receiving messages
	// 3. Send the signature request payload and get back a signed cert
	// We implement this with a terminatable goroutine that just sends acks and a while(true) loop that looks for responses
	terminateRoutineCh := make(chan interface{})
	go func() {
		for {
			select {
			case <-terminateRoutineCh:
				return
			default:

			}
			err = kbc.SendMessageByTeamName(config.TeamName, shared.AckRequest, nil)
			if err != nil {
				fmt.Printf("Failed to send AckRequest: %v\n", err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	fmt.Println("Waiting for a response from the CA....")
	hasBeenAcked := false
	for {
		msg, err := sub.Read()
		if err != nil {
			return empty, fmt.Errorf("failed to read message: %v", err)
		}

		if msg.Message.Content.Type != "text" {
			continue
		}

		if msg.Message.Sender.Username != config.BotName {
			continue
		}

		messageBody := msg.Message.Content.Text.Body

		if messageBody == shared.Ack && !hasBeenAcked {
			// We got an Ack so we terminate our AckRequests and send the real payload
			hasBeenAcked = true
			terminateRoutineCh <- true
			marshaledRequest, err := json.Marshal(request)
			if err != nil {
				return empty, err
			}
			err = kbc.SendMessageByTeamName(config.TeamName, shared.SignatureRequestPreamble+string(marshaledRequest), nil)
			if err != nil {
				return empty, err
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureResponsePreamble) {
			fmt.Println("Got a response from the CA!")
			resp, err := shared.ParseSignatureResponse(messageBody)
			if err != nil {
				fmt.Printf("Failed to parse a message from the bot: %s\n", messageBody)
				return empty, err
			}
			// A UUID mismatch just means there is a race condition and we are reading the CA bot's reply to
			// someone else's signature request
			if resp.UUID == request.UUID {
				return resp, nil
			}
		}
	}
}
