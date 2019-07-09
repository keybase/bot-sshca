package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"
	"github.com/keybase/bot-ssh-ca/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

func GetKBChat(conf config.Config) (*kbchat.API, error) {
	runOptions := kbchat.RunOptions{}
	if conf.GetUseAlternateAccount() {
		runOptions = kbchat.RunOptions{HomeDir: conf.GetKeybaseHomeDir(), Oneshot: &kbchat.OneshotOptions{PaperKey: conf.GetKeybasePaperKey(), Username: conf.GetKeybaseUsername()}}
	}
	return kbchat.Start(runOptions)
}

func GetUsername(conf config.Config) (string, error) {
	kbChat, err := GetKBChat(conf)
	if err != nil {
		return "", err
	}
	return kbChat.GetUsername(), nil
}

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
			err = kbc.SendMessage(msg.Message.Channel, shared.Ack)
			if err != nil {
				LogError(msg, err)
				continue
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureRequestPreamble) {
			signatureRequest, err := shared.ParseSignatureRequest(messageBody)
			if err != nil {
				LogError(msg, err)
				continue
			}
			signatureResponse, err := sshutils.ProcessSignatureRequest(conf, signatureRequest)
			if err != nil {
				LogError(msg, err)
				continue
			}

			response, err := json.Marshal(signatureResponse)
			if err != nil {
				LogError(msg, err)
				continue
			}
			err = kbc.SendMessage(msg.Message.Channel, shared.SignatureResponsePreamble+string(response))
			if err != nil {
				LogError(msg, err)
				continue
			}
		}
	}
}

func LogError(message kbchat.SubscriptionMessage, err error) {
	// TODO: Send these to chat?
	fmt.Printf("Got error while processing a message: %v\n", err)
}

func isConfiguredChannel(conf config.Config, channelName string) bool {
	return conf.GetTeamName() == channelName
}
