package main

import (
	"context"
	"strconv"
	"strings"
	"time"
)

func Handle(ctx context.Context, request string) string {
	parts := strings.Split(request, "|")
	if len(parts) != 2 || parts[0] != "PAYMENT" {
		return ResponseRejectedInvalidRequest
	}

	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return ResponseRejectedInvalidAmount
	}

	if amount <= 0 {
		return ResponseRejectedInvalidAmount
	}

	if amount > 100 {
		processingTime := min(amount, 10000)

		select {
		case <-time.After(time.Duration(processingTime) * time.Millisecond):
			// Processing completed normally.
		case <-ctx.Done():
			return ResponseRejectedCancelled
		}
	}

	return ResponseAccepted
}
