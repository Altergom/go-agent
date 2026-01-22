package memory

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

const SummaryMessageExtraKey = "_is_summary_message"

type Session struct {
	ID        string
	History   []*schema.Message
	Summary   string
	UpdatedAt time.Time
}
