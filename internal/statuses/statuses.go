package statuses

import (
	"strings"
)

const (
	StatusWait     Status = "âī¸"
	StatusReview   Status = "đ"
	StatusNeedWork Status = "đŠī¸"
	StatusFixing   Status = "đ§"
	StatusFixed    Status = "\U0001FA7C"
	StatusApproved Status = "â"
	StatusClosed   Status = "âī¸"
)

type Status string

func (s Status) IsFinal() bool {
	if s == StatusApproved || s == StatusClosed {
		return true
	}
	return false
}

func (s Status) String() string {
	return string(s)
}

func (s Status) Description() string {
	return s.String() + " " + StatusDescriptions[s]
}

var (
	StatusDescriptions = map[Status]string{ // Status:comment
		StatusWait:     "new",
		StatusReview:   "review",
		StatusNeedWork: "need work",
		StatusFixing:   "fixing",
		StatusFixed:    "fixed",
		StatusApproved: "approve",
		StatusClosed:   "close",
	}

	StatusTransitions = map[Status][]Status{
		StatusWait:     {StatusReview},
		StatusReview:   {StatusNeedWork, StatusApproved, StatusClosed},
		StatusNeedWork: {StatusFixing, StatusClosed},
		StatusFixing:   {StatusFixed, StatusClosed},
		StatusFixed:    {StatusReview},
	}
)

func ReplaceStatusInText(text string, dst Status) string {
	for s := range StatusDescriptions {
		if strings.Contains(text, s.String()) {
			return strings.Replace(text, s.String(), dst.String(), 1)
		}
	}
	return text
}
