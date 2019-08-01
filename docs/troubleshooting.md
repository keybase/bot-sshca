# Troubleshooting

This file contains some general directions and thoughts on troubleshooting the code in this repo. This is not meant
to be a comprehensive troubleshooting guide and is only a jumping off point. 

# kssh is slow, but it works

When kssh starts, it has to search every team you are in for a `kssh-client.config` file which specifies the information
that is needed in order to communicate with the CA chatbot. If you are only in a few teams, this is relatively fast 
(1-2 seconds for <10 teams) but this can become much slower as the number of teams increases (6 seconds for 100 teams
in my benchmarks). This complex start up procedure can be avoided by setting a default team via 
`kssh --set-default-team foo` which should reduce kssh's startup time considerably. 

# kssh times out

If kssh times out with a message similar to:

```
Generating a new SSH key...
Requesting signature from the CA....
Failed to get a signed key from the CA: timed out while waiting for a response from the CA
```

It means that for whatever reason, kssh is not receiving a response from the CA chatbot when it sends messages in 
Keybase chat. First, ensure that the CA chatbot is currently running. Next, attempt to determine what is happening
via inspecting the chat messages inside of the teams configured with the chatbot. You should see a series of `Ack` and 
`AckRequest` messages going back and forth prior to a `Signature_Request:` and a `Signature_Response:` exchange. Ensure 
that you and the chatbot are in the correct teams such that they can read and respond to the messages. 

# SSH rejects the connection

This likely means that you have not configured the SSH server correctly. Review the directions in README.md and ensure
that you have followed the steps correctly (sshca.md also has some additional information on how SSH CAs work that may
be helpful). If you would like to follow an example, see the code in the `tests/` directory which contains integration 
tests (focus on Dockerfile-sshd for an example SSH server setup). If none of that works, the best strategy is to run
SSH on the server on an alternate port and review the debug information. On the server run `/usr/sbin/sshd -dd -D -p 2222`
and on the client run `kssh -p 2222 user@server` and inspect the debug logs.  