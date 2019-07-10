package shared

func KeyPathToPubKey(keyPath string) string {
	return keyPath + ".pub"
}

func KeyPathToCert(keyPath string) string {
	return keyPath + "-cert.pub"
}
