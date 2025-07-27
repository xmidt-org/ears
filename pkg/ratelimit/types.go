// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"context"
)

type RateLimiter interface {

	//Take given units
	//Returns LimitReached if limit reached
	//Returns InvalidUnitError if unit <= 0 or > the max limit
	Take(ctx context.Context, unit int) error

	//Limit gets the current limit
	Limit() int

	//SetLimit sets the limit. 0 means there is no limit
	//Returns InvalidUnitError if newLimit < 0
	SetLimit(newLimit int) error
}
