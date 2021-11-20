package synpse

import "time"

type User struct {
	ID        string    `json:"id" yaml:"id"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`

	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty" `
	UserType string `json:"userType" yaml:"userType" gorm:"index"`

	// external users attributes
	ProviderName string            `json:"providerName,omitempty" yaml:"providerName,omitempty"`
	ProviderID   string            `json:"providerId,omitempty" yaml:"providerId,omitempty"`
	Info         map[string]string `json:"info,omitempty" yaml:"info,omitempty"`

	Quota UserQuota `json:"quota" yaml:"quota" gorm:"json"`
}

type UserQuota struct {
	Projects int `json:"projects" yaml:"projects"`
}
