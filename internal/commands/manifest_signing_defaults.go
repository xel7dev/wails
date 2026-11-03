package commands

import (
	"fmt"
	"os"

	"github.com/xel7dev/wails/internal/keychain"
	"github.com/xel7dev/wails/internal/signingconfig"
	"github.com/xel7dev/wails/internal/wake/manifest"
)

func manifestSigningPassword(config manifest.SigningPlatform, platform string, environment []string) (string, error) {
	if config.Credential != "" {
		password := environmentValues(mergeEnvironment(os.Environ(), environment))[config.Credential]
		if password == "" {
			return "", fmt.Errorf("%s signing password environment variable %s is empty", platform, config.Credential)
		}
		return password, nil
	}
	if !signingconfig.IsGlobalPath(config, "certificate", config.Certificate) {
		return "", nil
	}
	key := keychain.KeyWindowsCertPassword
	if platform == "linux" {
		key = keychain.KeyPGPPassword
	}
	password, _ := keychain.Get(key)
	return password, nil
}
