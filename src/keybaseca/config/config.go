package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keybase/bot-sshca/src/keybaseca/constants"

	"github.com/keybase/bot-sshca/src/keybaseca/botwrapper"

	"github.com/keybase/bot-sshca/src/shared"

	log "github.com/sirupsen/logrus"
)

// Represents a loaded and validated config for keybaseca
type Config interface {
	GetCAKeyLocation() string
	GetKeybaseHomeDir() string
	GetKeybasePaperKey() string
	GetKeybaseUsername() string
	GetKeyExpiration() string
	GetTeams() []string
	GetChatTeam() string
	GetChannelName() string
	GetLogLocation() string
	GetStrictLogging() bool
	DebugString() string
}

// Validate the given config file. If offline, do so without connecting to keybase (used in code that is meant
// to function without any reliance on Keybase).
func ValidateConfig(conf EnvConfig, offline bool) error {
	if len(conf.GetTeams()) == 0 {
		return fmt.Errorf("must specify at least one team via the TEAMS environment variable")
	}
	if conf.GetKeyExpiration() != "" && !strings.HasPrefix(conf.GetKeyExpiration(), "+") {
		// Only a basic check for this since ssh will error out later on if it is bogus
		return fmt.Errorf("KEY_EXPIRATION must be of the form `+<number><unit> where unit is one of `m`, `h`, `d`, `w`. Eg `+1h`. ")
	}
	if conf.GetLogLocation() != "" && !offline {
		err := validatePath(conf.GetLogLocation())
		if err != nil {
			return fmt.Errorf("LOG_LOCATION '%s' is not a valid path: %v", conf.GetLogLocation(), err)
		}
	}
	if conf.getChatChannel() != "" && !offline {
		team, channel, err := splitTeamChannel(conf.getChatChannel())
		if err != nil {
			return fmt.Errorf("Failed to parse CHAT_CHANNEL=%s: %v", conf.getChatChannel(), err)
		}
		err = validateChannel(&conf, team, channel)
		if err != nil {
			return fmt.Errorf("failed to validate CHAT_CHANNEL '%s': %v", channel, err)
		}
	}
	if conf.getStrictLogging() != "" {
		if conf.getStrictLogging() != "true" && conf.getStrictLogging() != "false" {
			return fmt.Errorf("STRICT_LOGGING must be either 'true' or 'false', '%s' is not valid", conf.getStrictLogging())
		}
	}
	if conf.GetKeybaseUsername() != "" || conf.GetKeybasePaperKey() != "" {
		if conf.GetKeybaseUsername() == "" && conf.GetKeybasePaperKey() != "" {
			return fmt.Errorf("you must set set a username if you set a paper key (username='%s', key='%s')", conf.GetKeybaseUsername(), conf.GetKeybasePaperKey())
		}
		if conf.GetKeybasePaperKey() == "" && conf.GetKeybaseUsername() != "" {
			return fmt.Errorf("you must set set a paper key if you set a username (username='%s', key='%s')", conf.GetKeybaseUsername(), conf.GetKeybasePaperKey())
		}
		if !offline {
			err := validateUsernamePaperkey(conf.GetKeybaseHomeDir(), conf.GetKeybaseUsername(), conf.GetKeybasePaperKey())
			if err != nil {
				return fmt.Errorf("failed to validate KEYBASE_USERNAME and KEYBASE_PAPERKEY: %v", err)
			}
		}
	}
  log.Debugf("Validated config: %s", conf.DebugString())
	return nil
}

func validateUsernamePaperkey(homedir, username, paperkey string) error {
	api, err := botwrapper.GetKBChat(homedir, username, paperkey)
	if err != nil {
		return err
	}
	validatedUsername := api.GetUsername()
	if validatedUsername == "" {
		return fmt.Errorf("failed to get a username from kbChat, got an empty string")
	}
	if validatedUsername != username {
		return fmt.Errorf("validated_username=%s and expected_username=%s do not match", validatedUsername, username)
	}
	return nil
}

// Validates the given teamName and channelName to determine whether or not the given channelName is the name
// of a channel inside the given team. Returns nil if everything validates.
func validateChannel(conf Config, teamName string, channelName string) error {
	api, err := botwrapper.GetKBChat(conf.GetKeybaseHomeDir(), conf.GetKeybasePaperKey(), conf.GetKeybaseUsername())
	if err != nil {
		return err
	}
	result, err := api.ListChannels(teamName)
	if err != nil {
		return err
	}

	for _, channel := range result {
		if channel == channelName {
			// The channel does exist, but the bot may or may not be in it. So join the channel in order to ensure
			// the bot will receive chat events from it
			_, err := api.JoinChannel(teamName, channelName)
			if err != nil {
				return fmt.Errorf("failed to join bot to the configured channel: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("did not find a channel named %s in %s", channelName, teamName)
}

// Returns an error if the given path is not a writable path on the local filesystem or on KBFS
func validatePath(path string) error {
	if strings.HasPrefix(path, "/keybase/") {
		// If it exists it is valid
		exists, _ := constants.GetDefaultKBFSOperationsStruct().KBFSFileExists(path)
		if exists {
			return nil
		}

		// Otherwise try to write to it
		err := constants.GetDefaultKBFSOperationsStruct().KBFSWrite(path, "", false)
		if err != nil {
			return fmt.Errorf("path is not writable: %v", err)
		}

		err = constants.GetDefaultKBFSOperationsStruct().KBFSDelete(path)
		if err != nil {
			return fmt.Errorf("failed to delete temp file: %v", err)
		}
		return nil
	}
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	var d []byte
	err = ioutil.WriteFile(path, d, 0600)
	if err != nil {
		return fmt.Errorf("path is not writable: %v", err)
	}
	os.Remove(path)
	return nil
}

// A Config struct that pulls transparently from the environment
type EnvConfig struct{}

var _ Config = (*EnvConfig)(nil)

// Get the location of the CA key
func (ef *EnvConfig) GetCAKeyLocation() string {
	if os.Getenv("CA_KEY_LOCATION") != "" {
		return shared.ExpandPathWithTilde(os.Getenv("CA_KEY_LOCATION"))
	}
	return shared.ExpandPathWithTilde("/mnt/keybase-ca-key")
}

// Get the keybase home directory. Used if you are running a separate instance of keybase for the chatbot. May be empty.
func (ef *EnvConfig) GetKeybaseHomeDir() string {
	return os.Getenv("KEYBASE_HOME_DIR")
}

// Get the keybase paper key for the bot account. Used if you are running a separate instance of keybase for the chatbot.
// May be empty.
func (ef *EnvConfig) GetKeybasePaperKey() string {
	return os.Getenv("KEYBASE_PAPERKEY")
}

// Get the keybase username for the bot account. Used if you are running a separate instance of keybase for the chatbot.
// May be empty.
func (ef *EnvConfig) GetKeybaseUsername() string {
	return os.Getenv("KEYBASE_USERNAME")
}

// Get the expiration period for signatures generated by the bot.
func (ef *EnvConfig) GetKeyExpiration() string {
	if os.Getenv("KEY_EXPIRATION") != "" {
		return os.Getenv("KEY_EXPIRATION")
	}
	return "+1h"
}

// Get the list of keybase teams configured to be used with the bot.
func (ef *EnvConfig) GetTeams() []string {
	split := strings.Split(os.Getenv("TEAMS"), ",")
	var teams []string
	for _, item := range split {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			teams = append(teams, trimmed)
		}
	}
	return teams
}

// Get the location for the bot's audit logs. May be empty.
func (ef *EnvConfig) GetLogLocation() string {
	return os.Getenv("LOG_LOCATION")
}

func (ef *EnvConfig) getStrictLogging() string {
	return strings.ToLower(os.Getenv("STRICT_LOGGING"))
}

// Get whether or not strict logging (see env.md for a description of this feature) is enabled
func (ef *EnvConfig) GetStrictLogging() bool {
	return ef.getStrictLogging() == "true"
}

// Get the Keybase chat location configured to be used for all communication. A chat channel consists of
// team.subteam#channel-name. May be empty.
func (ef *EnvConfig) getChatChannel() string {
	return os.Getenv("CHAT_CHANNEL")
}

// Get the team used for all communication. May be empty.
func (ef *EnvConfig) GetChatTeam() string {
	if ef.getChatChannel() == "" {
		return ""
	}
	team, _, err := splitTeamChannel(ef.getChatChannel())
	if err != nil {
		panic("Failed to retrieve chat team! This should never happen due to config validation...")
	}
	return team
}

// Get the channel used for all communication. May be empty.
func (ef *EnvConfig) GetChannelName() string {
	if ef.getChatChannel() == "" {
		return ""
	}
	_, channel, err := splitTeamChannel(ef.getChatChannel())
	if err != nil {
		panic("Failed to retrieve channel name! This should never happen due to config validation...")
	}
	return channel
}

// Dump this EnvConfig to a string for debugging purposes
func (ef *EnvConfig) DebugString() string {
	return fmt.Sprintf("CAKeyLocation='%s'; KeybaseHomeDir='%s'; KeybasePaperKey='%s'; KeybaseUsername='%s'; "+
		"KeyExpiration='%s'; Teams='%s'; ChatTeam='%s'; ChannelName='%s'; LogLocation='%s'; StrictLogging='%s'",
		ef.GetCAKeyLocation(), ef.GetKeybaseHomeDir(), ef.GetKeybasePaperKey(), ef.GetKeybaseUsername(),
		ef.GetKeyExpiration(), ef.GetTeams(), ef.GetChatTeam(), ef.GetChannelName(), ef.GetLogLocation(), ef.getStrictLogging())
}

// Split a teamChannel of the form team.foo.bar#chan into "team.foo.bar", "chan"
func splitTeamChannel(teamChannel string) (string, string, error) {
	split := strings.Split(teamChannel, "#")
	if len(split) != 2 {
		return "", "", fmt.Errorf("'%s' is not a valid specifier for a team and a channel", teamChannel)
	}
	return split[0], split[1], nil
}
