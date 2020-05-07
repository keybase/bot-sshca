package kssh

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keybase/bot-sshca/src/shared"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

type Requester struct {
	api *kbchat.API
}

// NewRequester creates a new Requester with a Keybase chat API
func NewRequester() (r Requester, err error) {
	runOptions := kbchat.RunOptions{KeybaseLocation: GetKeybaseBinaryPath()}
	api, err := kbchat.Start(runOptions)
	if err != nil {
		return r, fmt.Errorf("error starting Keybase chat: %v", err)
	}
	return Requester{api: api}, nil
}

// LoadConfigs loads kssh configs from the KV store. Returns a (listOfConfigs,
// listOfBotNames, err). Both lists are deduplicated based on Config.BotName.
func (r *Requester) LoadConfigs() (configs []Config, botNames []string, err error) {
	teams, err := r.getAllTeams()
	if err != nil {
		return nil, nil, err
	}
	botNameToConfig := make(map[string]Config)
	for _, team := range teams {
		conf, err := r.LoadConfig(team)
		if err != nil {
			return nil, nil, err
		}
		if conf != nil {
			// conf was found
			botNameToConfig[conf.BotName] = *conf
		}
	}
	for _, config := range botNameToConfig {
		configs = append(configs, config)
		botNames = append(botNames, config.BotName)
	}
	return configs, botNames, nil
}

// LoadConfig loads the kssh config for the given teamName. Will return a nil
// Config if no config was found for the teamName (and no error occurred)
func (r *Requester) LoadConfig(teamName string) (*Config, error) {
	res, err := r.api.GetEntry(&teamName, shared.SSHCANamespace, shared.SSHCAConfigKey)
	if err != nil {
		// error getting the entry
		return nil, err
	}
	if res.Revision > 0 && len(res.EntryValue) > 0 {
		// then this entry exists
		var conf Config
		if err := json.Unmarshal([]byte(res.EntryValue), &conf); err != nil {
			return nil, fmt.Errorf("Failed to parse config for team %s: %v", teamName, err)
		}
		if conf.TeamName == "" || conf.BotName == "" {
			return nil, fmt.Errorf("Found a config for team %s with missing data: %s", teamName, res.EntryValue)
		}
		return &conf, nil
	}
	// then entry doesn't exist; no config found
	return nil, nil
}

// LoadConfigForBot gets the Config associated with the given botName. Will
// need to find and load all configs from KV store.
func (r *Requester) LoadConfigForBot(botName string) (Config, error) {
	configs, _, err := r.LoadConfigs()
	if err != nil {
		return Config{}, err
	}
	for _, config := range configs {
		if config.BotName == botName {
			return config, nil
		}
	}
	return Config{}, fmt.Errorf("did not find a client config file matching botName=%s (is the CA bot running and are you in the correct teams?)", botName)
}

func (r *Requester) getAllTeams() (teams []string, err error) {
	// TODO: dedup with same method in keybaseca/bot
	memberships, err := r.api.ListUserMemberships(r.api.GetUsername())
	if err != nil {
		return teams, err
	}
	for _, m := range memberships {
		teams = append(teams, m.FqName)
	}
	return teams, nil
}

// Get a signed SSH key from interacting with the CA chatbot
func (r *Requester) GetSignedKey(botName string, request shared.SignatureRequest) (shared.SignatureResponse, error) {
	empty := shared.SignatureResponse{}

	conf, err := r.getConfig(botName)
	if err != nil {
		return empty, fmt.Errorf("failed to get config: %+v", err)
	}

	// Validate that the bot user is different than the current user
	if conf.BotName == r.api.GetUsername() {
		return empty, fmt.Errorf("cannot run kssh and keybaseca as the same user: %s", conf.BotName)
	}

	sub, err := r.api.ListenForNewTextMessages()
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
			_, err = r.api.SendMessageByTeamName(conf.TeamName, conf.getChannel(), shared.GenerateAckRequest(r.api.GetUsername()))
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

		if msg.Message.Content.TypeName != "text" {
			continue
		}

		if msg.Message.Sender.Username != conf.BotName {
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
			_, err = r.api.SendMessageByTeamName(conf.TeamName, conf.getChannel(), shared.SignatureRequestPreamble+string(marshaledRequest))
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
				// A UUID mismatch just means there is a race condition and we are
				// reading the CA bot's reply to someone else's signature request
				continue
			}
			return resp, nil
		}
	}
}

// Get the kssh config from the KV store. botName is the bot specified via
// --bot, else is an empty string
func (r *Requester) getConfig(botName string) (conf Config, err error) {
	empty := Config{}
	// They specified a bot via `kssh --bot cabot ...`
	if botName != "" {
		conf, err = r.LoadConfigForBot(botName)
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for bot=%s: %v", botName, err)
		}
		return conf, nil
	}

	// Check if they set a default bot and retrieve the config for that bot/team if so
	defaultBot, defaultTeam, err := GetDefaultBotAndTeam()
	if err != nil {
		return empty, err
	}
	if defaultBot != "" && defaultTeam != "" {
		conf, err := r.LoadConfig(defaultTeam)
		if err != nil || conf == nil {
			return empty, fmt.Errorf("Failed to load config file for default bot=%s, team=%s: %v", defaultBot, defaultTeam, err)
		}
		return *conf, nil
	}

	// No specified bot and no default bot, fallback and load all the configs
	configs, botNames, err := r.LoadConfigs()
	if err != nil {
		return empty, fmt.Errorf("Failed to load config(s): %v", err)
	}
	switch len(configs) {
	case 0:
		return empty, fmt.Errorf("Did not find any configs (is `keybaseca service` running?)")
	case 1:
		return configs[0], nil
	default:
		noDefaultTeamMessage := fmt.Sprintf("Found %d configs (%s). No default bot is configured. \n"+
			"Either specify a team via `kssh --bot cabotName` or set a default bot via `kssh --set-default-bot cabotName`", len(configs), strings.Join(botNames, ", "))
		return empty, fmt.Errorf(noDefaultTeamMessage)
	}
}
