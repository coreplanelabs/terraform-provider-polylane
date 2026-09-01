package client

type AWSConnectionRequest struct {
	ID                   string   `json:"id"`
	WorkspaceID          string   `json:"workspaceId"`
	AccountID            string   `json:"account"`
	Regions              []string `json:"regions"`
	ExternalID           string   `json:"externalId"`
	PrincipalARN         string   `json:"principalArn"`
	SubscriptionEndpoint string   `json:"subscriptionEndpoint"`
	CloudAccountID       string   `json:"cloudAccountId"`
	Region               string   `json:"region"`
	Status               string   `json:"status"`
}

type CreateAWSConnectionRequestInput struct {
	WorkspaceID    string
	AccountID      string
	Regions        []string
	IdempotencyKey string
}

type AWSConnection struct {
	ID                   string   `json:"id"`
	WorkspaceID          string   `json:"workspaceId"`
	RequestID            string   `json:"requestId"`
	AccountID            string   `json:"account"`
	Regions              []string `json:"regions"`
	Region               string   `json:"region"`
	RoleARN              string   `json:"roleArn"`
	BucketName           string   `json:"bucketName"`
	TopicARN             string   `json:"topicArn"`
	TopicSubscriptionARN string   `json:"topicSubscriptionArn"`
	CloudTrailName       string   `json:"cloudTrailName"`
	Status               string   `json:"status"`
}

type ActivateAWSConnectionInput struct {
	WorkspaceID          string
	RequestID            string
	Region               string
	RoleARN              string
	BucketName           string
	TopicARN             string
	TopicSubscriptionARN string
	CloudTrailName       string
}

type Workspace struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Description     *string `json:"description"`
	Domain          *string `json:"domain"`
	AutoJoinEnabled *string `json:"autoJoinEnabled"`
	OwnerID         string  `json:"ownerId"`
	Created         *string `json:"created"`
	Updated         *string `json:"updated"`
}

type UpdateWorkspaceInput struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	AutoJoin    *bool   `json:"autoJoin,omitempty"`
}

type WorkspaceMember struct {
	WorkspaceID string   `json:"workspaceId"`
	UserID      string   `json:"userId"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Forename    string   `json:"forename"`
	Surname     *string  `json:"surname"`
	InvitedBy   *string  `json:"invitedBy"`
	Position    string   `json:"position"`
	Scopes      []string `json:"scopes"`
	LastActive  *string  `json:"lastActive"`
	Created     *string  `json:"created"`
	Updated     *string  `json:"updated"`
}

type UpdateWorkspaceMemberInput struct {
	Position *string
	Scopes   *[]string
}

type Team struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspaceId"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	CreatedBy   string  `json:"createdBy"`
	Created     *string `json:"created"`
	Updated     *string `json:"updated"`
}

type CreateTeamInput struct {
	WorkspaceID string
	Name        string
	Slug        *string
	Description *string
	Color       *string
	Icon        *string
}

type UpdateTeamInput struct {
	WorkspaceID string
	Name        *string
	Slug        *string
	Description *string
	Color       *string
	Icon        *string
}

type TeamMember struct {
	TeamID  string  `json:"teamId"`
	UserID  string  `json:"userId"`
	AddedBy string  `json:"addedBy"`
	Created *string `json:"created"`
}

type AutofixSettings struct {
	Autofix struct {
		DisabledAt     *string `json:"disabledAt"`
		TriggerMode    string  `json:"triggerMode"`
		ReviewBotLogin *string `json:"reviewBotLogin"`
	} `json:"autofix"`
}

type DigestSettings struct {
	WeeklyDigestDisabledAt *float32 `json:"weeklyDigestDisabledAt"`
	DailyTrendsDisabledAt  *float32 `json:"dailyTrendsDisabledAt"`
}

type InvestigationsSettings struct {
	Investigations struct {
		PassesPerHypothesis int64 `json:"passesPerHypothesis"`
	} `json:"investigations"`
}

type InvestigationLimitsSettings struct {
	InvestigationLimits struct {
		Per24h                     *int64 `json:"per24h"`
		MinAutoInvestigateSeverity string `json:"minAutoInvestigateSeverity"`
	} `json:"investigationLimits"`
}

type ModelTrainingSettings struct {
	ModelTraining struct {
		AllowModelTraining *bool `json:"allowModelTraining"`
		Effective          bool  `json:"effective"`
	} `json:"modelTraining"`
}

type ObservabilitySettings struct {
	NativeObservability struct {
		Enabled bool `json:"enabled"`
	} `json:"nativeObservability"`
}

type PRReviewSettings struct {
	PRReviews struct {
		DisabledAt           *string `json:"disabledAt"`
		GuestLinksDisabledAt *string `json:"guestLinksDisabledAt"`
	} `json:"prReviews"`
}
