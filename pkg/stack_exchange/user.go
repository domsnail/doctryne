package stack_exchange

type UserList struct {
	Items []*User `json:"items"`

	HasMore        bool `json:"has_more"`
	Backoff        int  `json:"backoff"`
	QuotaMax       int  `json:"quota_max"`
	QuotaRemaining int  `json:"quota_remaining"`
}

type User struct {
	// https://meta.stackoverflow.com/a/332787/13775941
	AccountID uint64 `json:"account_id"`
	UserID    uint64 `json:"user_id"`

	DisplayName string `json:"display_name"`
	WebsiteUrl  string `json:"website_url"`
	AboutMe     string `json:"about_me"`
	Location    string `json:"location"`

	IsEmployee bool   `json:"is_employee"`
	AcceptRate int    `json:"accept_rate"`
	UserType   string `json:"user_type"`

	Reputation              uint32 `json:"reputation"`
	ReputationChangeYear    int32  `json:"reputation_change_year"`
	ReputationChangeQuarter int32  `json:"reputation_change_quarter"`
	ReputationChangeMonth   int32  `json:"reputation_change_month"`
	ReputationChangeWeek    int32  `json:"reputation_change_week"`
	ReputationChangeDay     int32  `json:"reputation_change_day"`

	ViewCount     int `json:"view_count"`
	DownVoteCount int `json:"down_vote_count"`
	UpVoteCount   int `json:"up_vote_count"`
	AnswerCount   int `json:"answer_count"`
	QuestionCount int `json:"question_count"`

	BadgeCounts struct {
		Bronze uint32 `json:"bronze"`
		Silver uint32 `json:"silver"`
		Gold   uint32 `json:"gold"`
	} `json:"badge_counts"`

	// unix timestamps
	CreationDate     UnixTimestamp `json:"creation_date"`
	LastModifiedDate UnixTimestamp `json:"last_modified_date"`
	LastAccessDate   UnixTimestamp `json:"last_access_date"`
	TimedPenaltyDate UnixTimestamp `json:"timed_penalty_date"`

	ProfileImage string `json:"profile_image"`
	Link         string `json:"link"`
}

func (s *User) IsRegistered() bool {
	switch s.UserType {
	case "registered", "moderator", "team_admin":
		return true
	default:
		return false
	}
}
