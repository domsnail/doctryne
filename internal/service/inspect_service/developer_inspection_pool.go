package inspect_service

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/domsnail/doctryne/internal/service"
)

type DeveloperInspectionPool struct {
	github service.IGithubService

	wg  *sync.WaitGroup
	ctx context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

type InspectionSource int

const (
	InspectionSource_GitHub InspectionSource = iota
	InspectionSource_StackExchange

	InspectionSource_DevTo
	InspectionSource_HeadHunter
	InspectionSource_Telegram
)
