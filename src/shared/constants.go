package shared

// The maximum number of goroutines to create when doing parallel operations
const BoundedParallelismLimit = 50

// The name of the KV store namespace for all bot-sshca-related data
const SSHCANamespace = "__sshca"

// The name of the KV store entry key for the kssh client config
const SSHCAConfigKey = "kssh_config"
