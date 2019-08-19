package shared

// The maximum number of goroutines to create when doing parallel operations
const BoundedParallelismLimit = 50

// The name of the kssh client config file that is stored in KBFS
const ConfigFilename = "kssh-client.config"
