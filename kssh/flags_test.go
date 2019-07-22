package kssh

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	var emptyCliArguments = []CLIArgument{}
	var cliArguments = []CLIArgument{
		{Name: "--set-deafult-team", HasArgument: true},
		{Name: "--team", HasArgument: true},
		{Name: "--provision", HasArgument: false},
	}
	// Parsing no args
	args := []string{}
	remaining, found, err := ParseArgs(args, cliArguments)
	assert.Nil(t, err)
	assert.Equal(t, remaining, []string{})
	assert.Equal(t, found, emptyCliArguments)

	// Parsing args that don't match
	args = []string{"hello", "world", "--foo", "--bar", "-h"}
	remaining, found, err = ParseArgs(args, cliArguments)
	assert.Nil(t, err)
	assert.Equal(t, remaining, args)
	assert.Equal(t, found, emptyCliArguments)

	// Parsing args that don't match
	args = []string{"hello", "world", "--foo", "--bar", "-h"}
	remaining, found, err = ParseArgs(args, emptyCliArguments)
	assert.Nil(t, err)
	assert.Equal(t, remaining, args)

	// Parsing an argument where HasArgument:true
	args = []string{"hello", "world", "--foo", "--bar", "-h", "--team", "foo", "bar"}
	remaining, found, err = ParseArgs(args, cliArguments)
	assert.Nil(t, err)
	assert.Equal(t, []string{"hello", "world", "--foo", "--bar", "-h", "bar"}, remaining)
	assert.Len(t, found, 1)
	assert.Equal(t, "--team", found[0].Name)
	assert.Equal(t, "foo", found[0].Value)

	// Parsing an argument where HasArgument:false
	args = []string{"hello", "world", "--foo", "--bar", "-h", "--provision", "foo", "bar"}
	remaining, found, err = ParseArgs(args, cliArguments)
	assert.Nil(t, err)
	assert.Equal(t, []string{"hello", "world", "--foo", "--bar", "-h", "foo", "bar"}, remaining)
	assert.Len(t, found, 1)
	assert.Equal(t, "--provision", found[0].Name)
	assert.Equal(t, "", found[0].Value)

	// Parsing multiple arguments
	// TODO
}

func isValidShuffle(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		nextArg := ""
		if i+1 < len(args) {
			nextArg = args[i+1]
		}
		if arg == "--team" && nextArg == "--provision" {
			return false
		}
		if arg == "--team" && nextArg == "" {
			return false
		}
	}
	return true
}

func TestParseArgsRandom(t *testing.T) {
	var cliArguments = []CLIArgument{
		{Name: "--team", HasArgument: true},
		{Name: "--provision", HasArgument: false},
	}

	s := time.Now().Unix()
	fmt.Printf("Seeding rand with %d in TestParseArgsRandom\n", s)
	rand.Seed(s)

	args := []string{"hello", "world", "--foo", "--bar", "-h", "--provision", "foo", "--team", "bar"}
	for i := 0; i < 100; i++ {
		rand.Shuffle(len(args), func(i, j int) {
			args[i], args[j] = args[j], args[i]
		})
		if isValidShuffle(args) {
			remaining, found, err := ParseArgs(args, cliArguments)
			assert.Nil(t, err)
			assert.Len(t, remaining, 6)
			assert.Len(t, found, 2)
		}
	}
}
