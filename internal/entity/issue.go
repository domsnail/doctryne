package entity

import "time"

type GithubIssue struct {
	ID     int64
	NodeID string
	Number int

	State string
	Title string
	Body  string

	IsLocked    bool
	LockedCause string // populated if IsLocked = true

	IsDraft       bool
	IsPullRequest bool

	Reactions     GithubIssueReactions
	CommentCounts int
	//Comments      []*GithubIssueComment

	Author *GithubDeveloperProfile

	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
}

type GithubIssueComment struct {
}

type GithubIssueReactions struct {
	TotalCount int

	PlusOne  int
	MinusOne int
	Laugh    int
	Confused int
	Heart    int
	Hooray   int
	Rocket   int
	Eyes     int
}
