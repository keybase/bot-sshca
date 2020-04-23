package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keybase/bot-sshca/src/keybaseca/botwrapper"

	auditlog "github.com/keybase/bot-sshca/src/keybaseca/log"

	"github.com/keybase/bot-sshca/src/keybaseca/sshutils"

	"github.com/keybase/bot-sshca/src/keybaseca/config"
	"github.com/keybase/bot-sshca/src/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"

	log "github.com/sirupsen/logrus"
)

// Get a running instance of the keybase chat API. Will use the configured credentials if necessary.
func GetKBChat(conf config.Config) (*kbchat.API, error) {
	return botwrapper.GetKBChat(conf.GetKeybaseHomeDir(), conf.GetKeybasePaperKey(), conf.GetKeybaseUsername())
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

// Start the keybaseca bot in an infinite loop. Does not return unless it encounters an unrecoverable error.
func StartBot(conf config.Config) error {
	kbc, err := GetKBChat(conf)
	if err != nil {
		return fmt.Errorf("error starting Keybase chat: %v", err)
	}

	err = sendAnnouncementMessage(conf, kbc)
	if err != nil {
		return fmt.Errorf("failed to start bot due to error while sending announcement: %v", err)
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

		if msg.Message.Content.TypeName != "text" {
			continue
		}

		messageBody := msg.Message.Content.Text.Body

		log.Debugf("Received message in %s#%s: %s", msg.Message.Channel.Name, msg.Message.Channel.TopicName, messageBody)

		if msg.Message.Sender.Username == kbc.GetUsername() {
			log.Debug("Skipping message since it comes from the bot user")
			if strings.Contains(messageBody, shared.AckRequestPrefix) || strings.Contains(messageBody, shared.SignatureRequestPreamble) {
				log.Warn("Ignoring AckRequest/SignatureRequest coming from the bot user! Are you trying to run the bot " +
					"and kssh as the same user?")
			}
			continue
		}

		// Note that this line is one of the main security barriers around the SSH bot. If this line were removed
		// or had a bug, it would cause the SSH bot to respond to any SignatureRequest messages in any channels. This
		// would allow an attacker to provision SSH keys even though they are not in the listed channels.
		if !isConfiguredTeam(conf, msg.Message.Channel.Name, msg.Message.Channel.TopicName) {
			log.Debug("Skipping message since it is not in a configured team")
			continue
		}

		if shared.IsPingRequest(messageBody, kbc.GetUsername()) {
			// Respond to messages of the form `ping @botName` with `pong @senderName`
			log.Debug("Responding to ping with pong")
			_, err = kbc.SendMessageByConvID(msg.Message.ConvID, shared.GeneratePingResponse(msg.Message.Sender.Username))
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
		} else if shared.IsAckRequest(messageBody) {
			// Ack any AckRequests so that kssh can determine whether it has fully connected
			_, err = kbc.SendMessageByConvID(msg.Message.ConvID, shared.GenerateAckResponse(messageBody))
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureRequestPreamble) {
			log.Debug("Responding to SignatureRequest")
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
			_, err = kbc.SendMessageByConvID(msg.Message.ConvID, shared.SignatureResponsePreamble+string(response))
			if err != nil {
				LogError(conf, kbc, msg, err)
				continue
			}
		} else {
			log.Debug("Ignoring unparsed message")
		}
	}
}

// Log the given error to Keybase chat and to the configured log file. Used so that the chatbot does not crash
// due to an error caused by a malformed message.
func LogError(conf config.Config, kbc *kbchat.API, msg kbchat.SubscriptionMessage, err error) {
	message := fmt.Sprintf("Encountered error while processing message from %s (messageID:%d): %v", msg.Message.Sender.Username, msg.Message.Id, err)
	auditlog.Log(conf, message)
	_, e := kbc.SendMessageByConvID(msg.Message.ConvID, message)
	if e != nil {
		auditlog.Log(conf, fmt.Sprintf("failed to log an error to chat (something is probably very wrong): %v", err))
	}
}

// Whether the given team is one of the specified teams in the config. Note that this function is a security boundary
// since it ensures that bots will not respond to messages outside of the configured teams.
func isConfiguredTeam(conf config.Config, teamName string, channelName string) bool {
	if conf.GetChatTeam() != "" {
		return conf.GetChatTeam() == teamName && conf.GetChannelName() == channelName
	}
	// If they didn't specify a chat team/channel, we just check whether the message was in one of the listed teams
	for _, team := range conf.GetTeams() {
		if team == teamName {
			return true
		}
	}
	return false
}

type AnnouncementTemplateValues struct {
	Username    string
	CurrentTeam string
	Teams       []string
}

func buildAnnouncement(template string, values AnnouncementTemplateValues) string {
	replacements := map[string]string{
		"{USERNAME}":     values.Username,
		"{CURRENT_TEAM}": values.CurrentTeam,
		"{TEAMS}":        strings.Join(values.Teams, ", "),
	}

	templatedMessage := template
	for templateStr, templateVal := range replacements {
		templatedMessage = strings.Replace(templatedMessage, templateStr, templateVal, -1)
	}

	return templatedMessage
}

func sendAnnouncementMessage(conf config.Config, kbc *kbchat.API) error {
	if conf.GetAnnouncement() == "" {
		// No announcement to send
		return nil
	}
	for _, team := range conf.GetTeams() {
		announcement := buildAnnouncement(conf.GetAnnouncement(),
			AnnouncementTemplateValues{Username: kbc.GetUsername(),
				CurrentTeam: team,
				Teams:       conf.GetTeams()})

		var channel *string
		_, err := kbc.SendMessageByTeamName(team, announcement, channel)
		if err != nil {
			return err
		}
	}
	return nil
}
