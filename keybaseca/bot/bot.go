package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keybase/bot-ssh-ca/keybaseca/log"

	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"

	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

func GetKBChat(conf config.Config) (*kbchat.API, error) {
	runOptions := kbchat.RunOptions{}
	if conf.GetKeybaseHomeDir() != "" {
		runOptions.HomeDir = conf.GetKeybaseHomeDir()
	}
	if conf.GetKeybasePaperKey() != "" && conf.GetKeybaseUsername() != "" {
		runOptions.Oneshot = &kbchat.OneshotOptions{PaperKey: conf.GetKeybasePaperKey(), Username: conf.GetKeybaseUsername()}
	}
	return kbchat.Start(runOptions)
}

// Get the username of the user that the keybaseca bot is running as
func GetUsername(conf config.Config) (string, error) {
	kbChat, err := GetKBChat(conf)
	if err != nil {
		return "", fmt.Errorf("failed to start Keybase chat: %v", err)
	}
	username := kbChat.GetUsername()
	if username == "" {
		return "", fmt.Errorf("failed to get a username from kbChat, got an empty string")
	}
	return username, nil
}

// Start the keybaseca bot in an infinite loop. Does not return unless it encounters an error.
func StartBot(conf config.Config) error {
	kbc, err := GetKBChat(conf)
	if err != nil {
		return fmt.Errorf("error starting Keybase chat: %v", err)
	}

	sub, err := kbc.ListenForNewTextMessages()
	if err != nil {
		return fmt.Errorf("error subscribing to messages: %v", err)
	}

	fmt.Println("Started CA bot...")
	for {
		msg, err := sub.Read()
		if err != nil {
			return fmt.Errorf("failed to read message: %v", err)
		}

		if msg.Message.Content.Type != "text" || msg.Message.Sender.Username == kbc.GetUsername() {
			continue
		}

		if !isConfiguredChannel(conf, msg.Message.Channel.Name) {
			continue
		}

		messageBody := msg.Message.Content.Text.Body

		if messageBody == shared.AckRequest {
			// Ack any AckRequests so that kssh can determine whether it has fully connected
			err = kbc.SendMessageByConvID(msg.Message.ConversationID, shared.Ack)
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureRequestPreamble) {
			signatureRequest, err := shared.ParseSignatureRequest(messageBody)
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
			signatureRequest.Username = msg.Message.Sender.Username
			signatureRequest.DeviceName = msg.Message.Sender.DeviceName
			signatureResponse, err := sshutils.ProcessSignatureRequest(conf, signatureRequest)
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}

			response, err := json.Marshal(signatureResponse)
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
			err = kbc.SendMessageByConvID(msg.Message.ConversationID, shared.SignatureResponsePreamble+string(response))
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
		}
	}
}

// Log the given error to Keybase chat and to the configured log file. Used so that the chatbot does not crash
// due to an error caused by a malformed message.
func LogError(conf config.Config, kbc *kbchat.API, msg kbchat.SubscriptionMessage, err error) {
	message := fmt.Sprintf("Encountered error while processing message from %s (messageID:%d): %v", msg.Message.Sender.Username, msg.Message.MsgID, err)
	log.Log(conf, message)
	e := kbc.SendMessageByConvID(msg.Message.ConversationID, message)
	if e != nil {
		log.Log(conf, fmt.Sprintf("failed to log an error to chat (something is probably very wrong): %v", err))
	}
}

// Whether the given channel is one of the specified channels in the config
func isConfiguredChannel(conf config.Config, channelName string) bool {
	for _, team := range conf.GetTeams() {
		if team == channelName {
			return true
		}
	}
	return false
}
