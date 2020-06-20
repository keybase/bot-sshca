package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/keybase/bot-sshca/src/keybaseca/botwrapper"
	"github.com/keybase/bot-sshca/src/kssh"

	auditlog "github.com/keybase/bot-sshca/src/keybaseca/log"

	"github.com/keybase/bot-sshca/src/keybaseca/sshutils"

	"github.com/keybase/bot-sshca/src/keybaseca/config"
	"github.com/keybase/bot-sshca/src/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"

	log "github.com/sirupsen/logrus"
)

// Bot is a SSH CA Keybase-backed bot
type Bot struct {
	conf config.Config
	api  *kbchat.API
}

// New creates a new Bot with a Keybase chat API
func New(conf config.Config) (ca Bot, err error) {
	api, err := botwrapper.GetKBChat(conf.GetKeybaseHomeDir(), conf.GetKeybasePaperKey(), conf.GetKeybaseUsername(), conf.GetKeybaseTimeout())
	if err != nil {
		return ca, fmt.Errorf("error starting Keybase chat: %v", err)
	}
	return Bot{conf: conf, api: api}, nil
}

// Start the SSH CA bot in an infinite loop. Does not return unless it
// encounters an unrecoverable error.
func (b *Bot) Start() error {
	err := b.writeClientConfig()
	if err != nil {
		return fmt.Errorf("failed to start CA bot due to error while writing client config: %v", err)
	}
	// don't let stale kssh configs stick around
	b.captureControlCToDeleteClientConfig()
	defer func() {
		if err = b.DeleteAllClientConfigs(); err != nil {
			fmt.Printf("Failed to delete all client configs on exit: %+v\n", err)
		}
	}()

	err = b.sendAnnouncementMessage()
	if err != nil {
		return fmt.Errorf("failed to start CA bot due to error while sending announcement: %v", err)
	}

	sub, err := b.api.ListenForNewTextMessages()
	if err != nil {
		return fmt.Errorf("error subscribing to messages: %v", err)
	}

	log.Debug("CA Bot now listening for messages...")
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

		if msg.Message.Sender.Username == b.api.GetUsername() {
			log.Debug("Skipping message since it comes from the CA bot user")
			if strings.Contains(messageBody, shared.AckRequestPrefix) || strings.Contains(messageBody, shared.SignatureRequestPreamble) {
				log.Warn("Ignoring AckRequest/SignatureRequest coming from the CA bot user! Are you trying to run the CA bot " +
					"and kssh as the same user?")
			}
			continue
		}

		// Note that this line is one of the main security barriers around the SSH
		// CA bot. If this line were removed or had a bug, it would cause the SSH
		// CA bot to respond to any SignatureRequest messages in any channels. This
		// would allow an attacker to provision SSH keys even though they are not
		// in the listed channels.
		if !b.isConfiguredTeam(msg.Message.Channel.Name, msg.Message.Channel.TopicName) {
			log.Debug("Skipping message since it is not in a configured team")
			continue
		}

		if shared.IsPingRequest(messageBody, b.api.GetUsername()) {
			// Respond to messages of the form `ping @botName` with `pong @senderName`
			log.Debug("Responding to ping with pong")
			_, err = b.api.SendMessageByConvID(msg.Message.ConvID, shared.GeneratePingResponse(msg.Message.Sender.Username))
			if err != nil {
				b.LogError(msg, err)
				continue
			}
		} else if shared.IsAckRequest(messageBody) {
			// Ack any AckRequests so that kssh can determine whether it has fully connected
			_, err = b.api.SendMessageByConvID(msg.Message.ConvID, shared.GenerateAckResponse(messageBody))
			if err != nil {
				b.LogError(msg, err)
				continue
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureRequestPreamble) {
			log.Debug("Responding to SignatureRequest")
			signatureRequest, err := shared.ParseSignatureRequest(messageBody)
			if err != nil {
				b.LogError(msg, err)
				continue
			}
			signatureRequest.Username = msg.Message.Sender.Username
			signatureRequest.DeviceName = msg.Message.Sender.DeviceName
			signatureResponse, err := sshutils.ProcessSignatureRequest(b.conf, signatureRequest)
			if err != nil {
				b.LogError(msg, err)
				continue
			}

			response, err := json.Marshal(signatureResponse)
			if err != nil {
				b.LogError(msg, err)
				continue
			}
			_, err = b.api.SendMessageByConvID(msg.Message.ConvID, shared.SignatureResponsePreamble+string(response))
			if err != nil {
				b.LogError(msg, err)
				continue
			}
		} else {
			log.Debug("Ignoring unparsed message")
		}
	}
}

// Write kssh config for kssh to use
func (b *Bot) writeClientConfig() error {
	username := b.api.GetUsername()
	if username == "" {
		return fmt.Errorf("failed to get a username from kbChat, got an empty string")
	}

	teams := b.conf.GetTeams()
	if b.conf.GetChatTeam() != "" {
		// Make sure we write the kssh config in the chat team, which may not be in
		// the list of teams
		teams = append(teams, b.conf.GetChatTeam())
	}
	log.Debugf("Attempting to write kssh configs for the teams: %v", teams)

	// If they configured a chat team, have messages go there
	config := kssh.Config{TeamName: b.conf.GetChatTeam(), BotName: username, ChannelName: b.conf.GetChannelName()}

	for _, team := range teams {
		if b.conf.GetChatTeam() == "" {
			// If they didn't configure a chat team, messages should be sent to any
			// channel. This is done by having each client config reference the team
			// it is found in.
			config.TeamName = team
			config.ChannelName = ""
		}

		var bytes []byte
		bytes, err := json.Marshal(config)
		if err != nil {
			log.Debugf("Failed to serialize kssh config (%v) for team %+v: %v", config, team, err)
			return err
		}
		_, err = b.api.PutEntry(&team, shared.SSHCANamespace, shared.SSHCAConfigKey, string(bytes))
		if err != nil {
			log.Debugf("Failed to write kssh config (%v) for team %v: %v", config, team, err)
			return err
		}
	}

	log.Debugf("Wrote kssh client configs for the teams: %v", teams)
	return nil
}

// Attempts to delete the kssh configs for the specified teams.
func (b *Bot) deleteClientConfig(teams []string) (found []string, err error) {
	log.Debugf("Attempting to delete kssh configs for the teams: %v", teams)
	for _, team := range teams {
		_, err := b.api.DeleteEntry(&team, shared.SSHCANamespace, shared.SSHCAConfigKey)
		if err != nil {
			kerr, ok := err.(kbchat.Error)
			if ok && kerr.Code == kbchat.DeleteNonExistentErrorCode {
				// ignore if we couldn't find the kssh config for this team
				log.Debugf("Did not find kssh config to delete for the team: %v", team)
			} else {
				// unexpected error
				log.Debugf("Unexpected error deleting kssh config for the team: %v", team)
				return found, err
			}
		} else {
			found = append(found, team)
		}
	}
	log.Debugf("Deleted kssh configs for the teams: %v", found)
	return found, nil
}

// Set up a signal handler in order to catch SIGTERMS that will delete all kssh
// configs for the configured teams when it receives a sigterm. This ensures
// that a simple Control-C does not leave behind stale kssh configs.
func (b *Bot) captureControlCToDeleteClientConfig() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("Losing CA bot, now deleting client configs...")
		teams := b.conf.GetTeams()
		if b.conf.GetChatTeam() != "" {
			// Make sure we delete the client config in the chat team which may not
			// be in the list of teams
			teams = append(teams, b.conf.GetChatTeam())
		}
		found, err := b.deleteClientConfig(teams)
		if err != nil {
			fmt.Printf("Failed to delete client configs: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted kssh configs for the teams: %v", found)
		os.Exit(0)
	}()
}

// DeleteAllClientConfigs deletes all found kssh configs for all teams the
// CA bot is a member of
func (b *Bot) DeleteAllClientConfigs() error {
	teams, err := b.getAllTeams()
	if err != nil {
		fmt.Printf("Failed to get teams to delete client configs: %v", err)
		return err
	}
	found, err := b.deleteClientConfig(teams)
	if err != nil {
		fmt.Printf("Failed to delete client configs: %v", err)
		return err
	}
	fmt.Printf("Deleted kssh configs for the teams: %v", found)
	return nil
}

func (b *Bot) getAllTeams() (teams []string, err error) {
	return shared.GetAllTeams(b.api)
}

// LogError logs the given error to Keybase chat and to the configured log file. Used so
// that the SSHCA bot does not crash due to an error caused by a malformed
// message.
func (b *Bot) LogError(msg kbchat.SubscriptionMessage, err error) {
	message := fmt.Sprintf("Encountered error while processing message from %s (messageID:%d): %v", msg.Message.Sender.Username, msg.Message.Id, err)
	auditlog.Log(b.conf, message)
	_, e := b.api.SendMessageByConvID(msg.Message.ConvID, message)
	if e != nil {
		auditlog.Log(b.conf, fmt.Sprintf("Failed to log an error to chat (something is probably very wrong): %v", err))
	}
}

// Whether the given team is one of the specified teams in the config. Note
// that this function is a security boundary since it ensures that CA bots will
// not respond to messages outside of the configured teams.
func (b *Bot) isConfiguredTeam(teamName string, channelName string) bool {
	if b.conf.GetChatTeam() != "" {
		return b.conf.GetChatTeam() == teamName && b.conf.GetChannelName() == channelName
	}
	// If they didn't specify a chat team/channel, we just check whether the
	// message was in one of the listed teams
	for _, team := range b.conf.GetTeams() {
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

func (b *Bot) sendAnnouncementMessage() error {
	if b.conf.GetAnnouncement() == "" {
		// No announcement to send
		return nil
	}
	for _, team := range b.conf.GetTeams() {
		announcement := buildAnnouncement(b.conf.GetAnnouncement(),
			AnnouncementTemplateValues{Username: b.api.GetUsername(),
				CurrentTeam: team,
				Teams:       b.conf.GetTeams()})

		var channel *string
		_, err := b.api.SendMessageByTeamName(team, channel, announcement)
		if err != nil {
			return err
		}
	}
	return nil
}
