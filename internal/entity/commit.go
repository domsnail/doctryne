package entity

import "time"

type Commit struct {
	Hash    string
	Message string

	Author *Developer

	CreatedAt time.Time

	Stats *CommitStats
}

type CommitStats struct {
	ChangedFiles []string

	LinesAdded   int
	LinesDeleted int
}

func (stats *CommitStats) Add(addition *CommitStats) {
	stats.ChangedFiles = append(stats.ChangedFiles, addition.ChangedFiles...)

	stats.LinesAdded += addition.LinesAdded
	stats.LinesDeleted += addition.LinesDeleted
}

func (stats *CommitStats) Total() (lines int) {
	return stats.LinesAdded + stats.LinesDeleted
}

type DeveloperCommitStats map[*Developer]*CommitStats
