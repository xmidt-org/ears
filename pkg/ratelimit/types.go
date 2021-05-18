// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
