package utils

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ConvertToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
