package kssh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	args      []string
	err       error
	remaining []string
	found     []ParsedCLIArgument
}

func TestParseArgs(t *testing.T) {
	var cliArguments = []CLIArgument{
		{Name: "--arg-with-value", HasArgument: true},
		{Name: "--arg2-with-value", HasArgument: true},
		{Name: "--arg-without-value", HasArgument: false},
		{Name: "--preserved-flag", Preserve: true},
	}
	testCases := []testCase{
		// No arguments
		{
			args:      []string{},
			err:       nil,
			remaining: []string{},
			found:     []ParsedCLIArgument{},
		},
		// Arguments that don't match
		{
			args:      []string{"hello", "world", "--foo", "--bar", "-h"},
			err:       nil,
			remaining: []string{"hello", "world", "--foo", "--bar", "-h"},
			found:     []ParsedCLIArgument{},
		},
		// HasArgument:true
		{
			args:      []string{"hello", "world", "--foo", "--bar", "-h", "--arg2-with-value", "foo", "bar"},
			err:       nil,
			remaining: []string{"hello", "world", "--foo", "--bar", "-h", "bar"},
			found: []ParsedCLIArgument{
				{Argument: CLIArgument{Name: "--arg2-with-value", HasArgument: true}, Value: "foo"},
			},
		},
		// HasArgument:false
		{
			args:      []string{"hello", "world", "--foo", "--bar", "-h", "--arg-without-value", "foo", "bar"},
			err:       nil,
			remaining: []string{"hello", "world", "--foo", "--bar", "-h", "foo", "bar"},
			found: []ParsedCLIArgument{
				{Argument: CLIArgument{Name: "--arg-without-value", HasArgument: false}, Value: ""},
			},
		},
		// Multiple arguments
		{
			args:      []string{"--arg-without-value", "foo", "--arg2-with-value", "bar", "--arg-with-value", "foobar"},
			err:       nil,
			remaining: []string{"foo"},
			found: []ParsedCLIArgument{
				{Argument: CLIArgument{Name: "--arg-without-value", HasArgument: false}, Value: ""},
				{Argument: CLIArgument{Name: "--arg2-with-value", HasArgument: true}, Value: "bar"},
				{Argument: CLIArgument{Name: "--arg-with-value", HasArgument: true}, Value: "foobar"},
			},
		},
		// HasArgument:true but the argument is missing
		{
			args:      []string{"--arg-without-value", "foo", "--arg2-with-value", "bar", "--arg-with-value"},
			err:       fmt.Errorf("argument --arg-with-value requires a value"),
			remaining: nil,
			found:     nil,
		},
		// Preserve:true
		{
			args:      []string{"--preserved-flag", "--arg-without-value", "unused"},
			err:       nil,
			remaining: []string{"--preserved-flag", "unused"},
			found: []ParsedCLIArgument{
				{Argument: CLIArgument{Name: "--preserved-flag", HasArgument: false, Preserve: true}, Value: ""},
				{Argument: CLIArgument{Name: "--arg-without-value", HasArgument: false}, Value: ""},
			},
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("Running %d\n", i)
		remaining, found, err := ParseArgs(testCase.args, cliArguments)
		require.Equal(t, testCase.err, err)
		require.Equal(t, testCase.remaining, remaining)
		require.Equal(t, testCase.found, found)
	}
}
