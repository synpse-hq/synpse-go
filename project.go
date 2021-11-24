package synpse

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// CreateProject creates a new project for the user.
// Note: this API can only be called with personal access keys (https://cloud.synpse.net/access-keys)
// and cannot be used when using a Service Account that was created inside the projec
func (api *API) CreateProject(ctx context.Context, project Project) (*Project, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL), project)
	if err != nil {
		return nil, err
	}

	var result Project
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

type ListProjectsRequest struct{}

// ListProjects returns a list of projects that the user has access to.
// Note: this API can only be called with personal access keys (https://cloud.synpse.net/access-keys)
// and cannot be used when using a Service Account that was created inside the project.
func (api *API) ListProjects(ctx context.Context, req *ListProjectsRequest) ([]Project, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, membershipsURL+"?full"), nil)
	if err != nil {
		return nil, err
	}

	var memberships []Membership
	err = json.Unmarshal(resp, &memberships)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	var projects []Project
	for _, m := range memberships {
		projects = append(projects, m.Project)
	}
	return projects, nil
}

type Project struct {
	ID        string    `json:"id" yaml:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`
	Name      string    `json:"name" yaml:"name" gorm:"uniqueIndex" validate:"required,lowercase,userTitleRegex"`

	Quota ProjectQuota `json:"quota" yaml:"quota" gorm:"json"`

	// Extra data fields (optional)
	DeviceCount      int `json:"deviceCount" yaml:"deviceCount"`
	ApplicationCount int `json:"applicationCount" yaml:"applicationCount"`
	NamespaceCount   int `json:"namespaceCount" yaml:"namespaceCount"`
	SecretCount      int `json:"secretCount" yaml:"secretCount"`
}

type ProjectQuota struct {
	Devices          int           `json:"devices" yaml:"devices"`
	Namespaces       int           `json:"namespaces" yaml:"namespaces"`
	Applications     int           `json:"applications" yaml:"applications"`
	Secrets          int           `json:"secrets" yaml:"secrets"`
	LogsRetention    time.Duration `json:"logsRetention" yaml:"logsRetention"`
	MetricsRetention time.Duration `json:"metricsRetention" yaml:"metricsRetention"`
}

type Membership struct {
	UserID    string `json:"userId" yaml:"userId"` // composite key
	ProjectID string `json:"projectId" yaml:"projectId"`

	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`

	// extra data fields (optional)
	User    User    `json:"user,omitempty" yaml:"user,omitempty"`
	Project Project `json:"project,omitempty" yaml:"project,omitempty"`
	Roles   []Role  `json:"roles,omitempty" yaml:"roles,omitempty"`
}

type Role struct {
	ID          string    `json:"id" yaml:"id"`
	CreatedAt   time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" yaml:"updatedAt"`
	ProjectID   string    `json:"projectId" yaml:"projectId"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Config      Config    `json:"config" yaml:"config"`
}

type Config struct {
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

type Rule struct {
	Resources []Resource `json:"resource,omitempty" yaml:"resource,omitempty"`
	Actions   []Action   `json:"actions,omitempty" yaml:"actions,omitempty"`
	Effect    Effect     `json:"effect,omitempty" yaml:"effect,omitempty"`
}

type Action string

const (
	ActionGet    Action = "Get"
	ActionList   Action = "List"
	ActionDelete Action = "Delete"
	ActionUpdate Action = "Update"
	ActionCreate Action = "Create"

	ActionSSH      Action = "SSH"
	ActionViewLogs Action = "ViewLogs"
	ActionWipe     Action = "Wipe"
	ActionReboot   Action = "Reboot"
	ActionConnect  Action = "Connect"
)

// Resource represents resource type
type Resource string

const (
	ResourceProject                   Resource = "Project"
	ResourceSecret                    Resource = "Secret"
	ResourceNamespace                 Resource = "Namespace"
	ResourceRole                      Resource = "Role"
	ResourceMembership                Resource = "Membership"
	ResourceMembershipRoleBinding     Resource = "MembershipRoleBinding"
	ResourceServiceAccount            Resource = "ServiceAccount"
	ResourceServiceAccountAccessKey   Resource = "ServiceAccountAccessKey"
	ResourceServiceAccountRoleBinding Resource = "ServiceAccountRoleBinding"
	ResourceApplication               Resource = "Application"
	ResourceApplicationAllocation     Resource = "ApplicationAllocation" // Computed resource
	ResourceDevice                    Resource = "Device"
	ResourceDeviceRegistrationToken   Resource = "DeviceRegistrationToken"
	ResourceAny                       Resource = "*"
)

type Effect string

const (
	EffectAllow = Effect("allow")
	EffectDeny  = Effect("deny")
)
