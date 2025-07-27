package token

import uuid "github.com/satori/go.uuid"

func GenUUID(name string) string {
	return uuid.NewV3(uuid.NamespaceDNS, name).String()
}
