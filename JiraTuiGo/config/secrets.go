package config

import (
	"encoding/base64"
	"log"

	"github.com/99designs/keyring"
)

const (
	keyringService = "JiraTuiGo"
	jiraTokenKey   = "jira-api-token"
	aiTokenKey     = "ai-api-key"
)

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: keyringService,
		KeychainTrustApplication: true,
	})
}

func Protect(plain string) string {
	kr, err := openKeyring()
	if err != nil {
		log.Printf("keyring unavailable (%v); falling back to base64", err)
		return "b64:" + base64.StdEncoding.EncodeToString([]byte(plain))
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	if err := kr.Set(keyring.Item{Key: jiraTokenKey, Data: []byte(encoded)}); err != nil {
		log.Printf("keyring set failed (%v); falling back to base64", err)
		return "b64:" + encoded
	}
	return "keyring:" + jiraTokenKey
}

func Unprotect(enc string) string {
	if len(enc) >= 4 && enc[:4] == "b64:" {
		data, err := base64.StdEncoding.DecodeString(enc[4:])
		if err != nil {
			return ""
		}
		return string(data)
	}

	if len(enc) >= 8 && enc[:8] == "keyring:" {
		key := enc[8:]
		kr, err := openKeyring()
		if err != nil {
			return ""
		}
		item, err := kr.Get(key)
		if err != nil {
			return ""
		}
		data, err := base64.StdEncoding.DecodeString(string(item.Data))
		if err != nil {
			return ""
		}
		return string(data)
	}

	return enc
}

func ProtectAIKey(plain string) string {
	kr, err := openKeyring()
	if err != nil {
		log.Printf("keyring unavailable (%v); falling back to base64", err)
		return "b64:" + base64.StdEncoding.EncodeToString([]byte(plain))
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	if err := kr.Set(keyring.Item{Key: aiTokenKey, Data: []byte(encoded)}); err != nil {
		log.Printf("keyring set failed (%v); falling back to base64", err)
		return "b64:" + encoded
	}
	return "keyring:" + aiTokenKey
}
