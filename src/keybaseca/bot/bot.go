package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"

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

type OutstandingTwoManSignatureRequest struct {
	SignatureRequest shared.SignatureRequest
	RequestMessageID chat1.MessageID
	Approvers        []string
	ConvID           string
}

// Start the keybaseca bot in an infinite loop. Does not return unless it encounters an unrecoverable error.
func StartBot(conf config.Config) error {
	// Initialize a list for the outstanding two-man signature requests
	outstandingTwoManRequests := []OutstandingTwoManSignatureRequest{}

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

		if msg.Message.Content.TypeName != "text" && msg.Message.Content.TypeName != "reaction" {
			continue
		}

		if msg.Message.Content.TypeName == "reaction" {
			emoji := msg.Message.Content.Reaction.Body
			responseTo := msg.Message.Content.Reaction.MessageID

			log.Debug("Examining reaction...")
			for _, outstanding := range outstandingTwoManRequests {
				if outstanding.RequestMessageID == responseTo {
					log.Debug("Message is a reply to an outstanding two-man request")
					approver := msg.Message.Sender.Username
					if emoji == ":+1:" && isValidApprover(conf, approver, outstanding.SignatureRequest) {
						isDuplicateApprover := addApprover(&outstanding, approver)
						if isDuplicateApprover {
							// TODO: Test this code
							log.Debugf("Rejecting duplicate approver %s since they already approved the two-man request with ID=%s", approver, outstanding.SignatureRequest.UUID)
							continue
						}
						log.WithField("requester", outstanding.SignatureRequest.Username).
							WithField("current_approver", approver).
							WithField("all_approvers", outstanding.Approvers).
							Debugf("Message approved request")
						threshold := conf.GetNumberRequiredApprovers()
						if len(outstanding.Approvers) >= threshold {
							respondToSignatureRequest(conf, kbc, outstanding.SignatureRequest, outstanding.SignatureRequest.Username, outstanding.RequestMessageID, outstanding.ConvID)
							auditlog.Log(conf, fmt.Sprintf("Two-man SignatureRequest id=%s approved by %v", outstanding.SignatureRequest.UUID, outstanding.Approvers))
						}
					} else {
						log.Debug("Message did not approve request")
					}
					continue
				}
			}
			log.Debug("Ignoring reaction since it is not a reaction on an outstanding two-man request")
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

		if shared.IsAckRequest(messageBody) {
			log.Debug("Responding to AckMessage")
			// Ack any AckRequests so that kssh can determine whether it has fully connected
			_, err = kbc.SendMessageByConvID(msg.Message.ConvID, shared.GenerateAckResponse(messageBody))
			if err != nil {
				LogError(conf, kbc, msg.Message.Sender.Username, msg.Message.Id, msg.Message.ConvID, err)
				continue
			}
		} else if strings.HasPrefix(messageBody, shared.SignatureRequestPreamble) {
			log.Debug("Responding to SignatureRequest")
			signatureRequest, err := shared.ParseSignatureRequest(messageBody)
			if err != nil {
				LogError(conf, kbc, msg.Message.Sender.Username, msg.Message.Id, msg.Message.ConvID, err)
				continue
			}
			signatureRequest.Username = msg.Message.Sender.Username
			signatureRequest.DeviceName = msg.Message.Sender.DeviceName

			// Process the signature request depending on whether they requested two-man only principals or not
			if signatureRequest.RequestedPrincipal == "" {
				// If they didn't request a principal just respond immediately
				respondToSignatureRequest(conf, kbc, signatureRequest, msg.Message.Sender.Username, msg.Message.Id, msg.Message.ConvID)
			} else {
				if isTwoManPrincipal(conf, signatureRequest.RequestedPrincipal) {
					// If they requested a principal that doesn't require two-man authorization, respond immediately
					respondToSignatureRequest(conf, kbc, signatureRequest, msg.Message.Sender.Username, msg.Message.Id, msg.Message.ConvID)
				} else {
					// If the principal requires two-man authorization, treat it as such
					resp, err := kbc.SendMessageByConvID(msg.Message.ConvID, buildTwoManApprovalRequestMessage(conf, msg.Message.Sender.Username, signatureRequest.RequestedPrincipal))
					if err != nil {
						LogError(conf, kbc, msg.Message.Sender.Username, msg.Message.Id, msg.Message.ConvID, err)
						continue
					}

					outstandingTwoManRequests = append(outstandingTwoManRequests,
						OutstandingTwoManSignatureRequest{SignatureRequest: signatureRequest, Approvers: []string{}, RequestMessageID: *resp.Result.MessageID, ConvID: msg.Message.ConvID})
				}
			}
		} else {
			log.Debug("Ignoring unparsed message")
		}
	}
}

// Add the given approver to the list of approvers in the given outstanding two man request if the
// given user has not already approved the request. Returns whether the given approver has already
// approved the request.
func addApprover(request *OutstandingTwoManSignatureRequest, approver string) bool {
	for _, curApprover := range request.Approvers {
		if curApprover == approver {
			return true
		}
	}
	request.Approvers = append(request.Approvers, approver)
	return false
}

func buildTwoManApprovalRequestMessage(conf config.Config, sender string, requestedPrincipal string) string {
	approvers := []string{}
	for _, approver := range conf.GetTwoManApprovers() {
		approvers = append(approvers, "@"+approver)
	}

	return fmt.Sprintf("@%s has requested access to the two-man realm %s! In order to approve this access, "+
		"reply with a thumbs-up to this message. (Configured approvers: %s)", sender, requestedPrincipal, strings.Join(approvers, ", "))
}

func isTwoManPrincipal(conf config.Config, requestedPrincipal string) bool {
	for _, team := range conf.GetTwoManApprovers() {
		if team == requestedPrincipal {
			return true
		}
	}
	return false
}

// Note that this function is a key security barrier for the two-man feature. This function checks that only people
// in the define list of approvers can approve a request and that people cannot approve their own request.
func isValidApprover(conf config.Config, senderUsername string, signatureRequest shared.SignatureRequest) bool {
	validApprover := false
	for _, knownApprover := range conf.GetTwoManApprovers() {
		if knownApprover == senderUsername {
			validApprover = true
		}
	}
	if !validApprover {
		log.Debug("Reply came from someone who isn't a valid two man approver, rejecting!")
		return false
	}
	//if senderUsername == signatureRequest.Username {
	//	log.Debug("Reply came from the sender of the signature request, rejecting!")
	//	return false
	//}
	return true
}

// Respond to the given SignatureRequest and log any errors that are produced. This function does not return any error.
func respondToSignatureRequest(conf config.Config, kbc *kbchat.API, signatureRequest shared.SignatureRequest, Username string, MessageID chat1.MessageID, conversationID string) {
	signatureResponse, err := sshutils.ProcessSignatureRequest(conf, signatureRequest)
	if err != nil {
		LogError(conf, kbc, Username, MessageID, conversationID, err)
		return
	}

	response, err := json.Marshal(signatureResponse)
	if err != nil {
		LogError(conf, kbc, Username, MessageID, conversationID, err)
		return
	}
	_, err = kbc.SendMessageByConvID(conversationID, shared.SignatureResponsePreamble+string(response))
	if err != nil {
		LogError(conf, kbc, Username, MessageID, conversationID, err)
		return
	}
}

// Log the given error to Keybase chat and to the configured log file. Used so that the chatbot does not crash
// due to an error caused by a malformed message.
func LogError(conf config.Config, kbc *kbchat.API, Username string, MessageID chat1.MessageID, conversationID string, err error) {
	message := fmt.Sprintf("Encountered error while processing message from %s (messageID:%d): %v", Username, MessageID, err)
	auditlog.Log(conf, message)
	_, e := kbc.SendMessageByConvID(conversationID, message)
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
