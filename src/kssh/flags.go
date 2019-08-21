package kssh

import "fmt"

type CLIArgument struct {
	// The name of the flag eg "--foo"
	Name string

	// HasArgument:true if an argument comes after it (eg "--foo bar") false if it is a boolean flag (eg "--help")
	HasArgument bool

	// Preserve:true if you wish to preserve this argument into the list of remaining arguments even if found
	// eg if your command takes in a `-v` flag and the subcommand also takes in a `-v` flag
	// incompatible with HasArgument: true but only because that has not been built
	Preserve bool
}

type ParsedCLIArgument struct {
	// The CLIArgument that was found and parsed
	Argument CLIArgument

	// The value associated with it if HasArgument:true. Otherwise an empty string.
	Value string
}

// ParseArgs parses os.Args for use with kssh. This is handwritten rather than using go's flag library (or
// any other CLI argument parsing library) since we want to have custom arguments and access any other remaining
// arguments. See this Github discussion for a longer discussion of why this is implemented this way:
// https://github.com/keybase/bot-sshca/pull/3#discussion_r302740696
//
// Returns: a list of the remaining unparsed arguments, a list of the parsed arguments, error
func ParseArgs(args []string, cliArguments []CLIArgument) ([]string, []ParsedCLIArgument, error) {
	for _, cliArg := range cliArguments {
		if cliArg.Preserve && cliArg.HasArgument {
			return nil, nil, fmt.Errorf("cannot specify Preserve and HasArgument for argument %s", cliArg.Name)
		}
	}

	remainingArguments := []string{}
	found := []ParsedCLIArgument{}
OUTER:
	for i := 0; i < len(args); i++ {
		arg := args[i]
		for _, cliArg := range cliArguments {
			if cliArg.Name == arg {
				parsed := ParsedCLIArgument{Argument: cliArg}
				if cliArg.HasArgument {
					if i+1 == len(args) {
						return nil, nil, fmt.Errorf("argument %s requires a value", cliArg.Name)
					}
					nextArg := args[i+1]
					parsed.Value = nextArg
					i++
				}
				found = append(found, parsed)
				if cliArg.Preserve {
					remainingArguments = append(remainingArguments, arg)
				}
				continue OUTER
			}
		}
		remainingArguments = append(remainingArguments, arg)
	}
	return remainingArguments, found, nil
}
