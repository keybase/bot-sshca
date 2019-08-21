package kssh

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

// Get a signed SSH key from interacting with the CA chatbot
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
		// Make the AckRequests send less often over time by tracking how many we've sent
		numberSent := 0
		for {
			select {
			case <-terminateRoutineCh:
				return
			default:

			}
			_, err = kbc.SendMessageByTeamName(config.TeamName, shared.GenerateAckRequest(kbc.GetUsername()), getChannel(config))
			if err != nil {
				fmt.Printf("Failed to send AckRequest: %v\n", err)
			}
			numberSent++
			time.Sleep(time.Duration(100+(10*numberSent)) * time.Millisecond)
		}
	}()

	hasBeenAcked := false
	startTime := time.Now()
	for {
		if time.Since(startTime) > 5*time.Second {
			return empty, fmt.Errorf("timed out while waiting for a response from the CA")
		}
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

		if shared.IsAckResponse(messageBody) && !hasBeenAcked {
			// We got an Ack so we terminate our AckRequests and send the real payload
			hasBeenAcked = true
			terminateRoutineCh <- true
			marshaledRequest, err := json.Marshal(request)
			if err != nil {
				return empty, err
			}
			_, err = kbc.SendMessageByTeamName(config.TeamName, shared.SignatureRequestPreamble+string(marshaledRequest), getChannel(config))
			if err != nil {
				return empty, err
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureResponsePreamble) {
			resp, err := shared.ParseSignatureResponse(messageBody)
			if err != nil {
				fmt.Printf("Failed to parse a message from the bot: %s\n", messageBody)
				return empty, err
			}

			if resp.UUID != request.UUID {
				// A UUID mismatch just means there is a race condition and we are reading the CA bot's reply to
				// someone else's signature request
				continue
			}
			return resp, nil
		}
	}
}

// Get the configured channel name from the given config file. Returns either a pointer to the channel name string
// or a null pointer.
func getChannel(config ConfigFile) *string {
	var channel *string
	if config.ChannelName != "" {
		channel = &config.ChannelName
	}
	return channel
}
