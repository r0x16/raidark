package domain

import "github.com/casdoor/casdoor-go-sdk/casdoorsdk"

// ManagedAccount represents a managed account in the domain using composition with Casdoor SDK
type ManagedAccount struct {
	casdoorsdk.ManagedAccount // Embed Casdoor managed account struct for direct access to all fields
}

// Business methods for domain logic
func (ma *ManagedAccount) IsConfigured() bool {
	return ma.Application != "" && ma.Username != ""
}

func (ma *ManagedAccount) HasValidCredentials() bool {
	return ma.Username != "" && ma.Password != ""
}
