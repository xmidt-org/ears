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

package template

import (
	"fmt"
	"regexp"
	"strings"
)

const maxContentCharacters = 30
const missingKeyMessage = " no entry for key "

// == templateError interface ========================================================

// isMissingKeyError is brittle test based on the template
func isMissingKeyError(e error) bool {
	return strings.Contains(e.Error(), missingKeyMessage)
}

// getMissingKeyFromError extracts the missing key based on the
// error message returned by the template parser
func getMissingKeyFromError(e error) string {
	s := e.Error()
	i := strings.LastIndex(s, missingKeyMessage)
	if i > 1 {
		return strings.Trim(s[i+len(missingKeyMessage):], ` "`)
	}
	return ""
}

// truncate limits a string's size and adds ellipsis if longer than length
func truncate(s string, length int) string {

	spaces := regexp.MustCompile(`\s+`)
	s = spaces.ReplaceAllString(s, " ")

	ellipsis := ""

	if len(s) > (length - 3) {
		length -= 3
		ellipsis = "..."
	}

	if len(s) < length {
		length = len(s)
	}
	return string([]rune(s)[:length]) + ellipsis
}

// Error prints out the error and any related details
func (e *ParseError) Error() string {
	s := fmt.Sprintf(`ParseError: content="%s"`, truncate(e.Content, maxContentCharacters))
	if e.Err != nil {
		return fmt.Sprintf("%s, err=%s", s, e.Err.Error())
	}
	return s
}

// Unwrap returns the wrapped error (Go v1.13)
func (e *ParseError) Unwrap() error {
	return e.Err
}

// Error prints out the error and any related details
func (e *MissingKeyError) Error() string {
	s := fmt.Sprintf(
		`MissingKeyError: key="%s", content="%s", values=%s`,
		truncate(e.Key, maxContentCharacters),
		truncate(e.Content, maxContentCharacters),
		truncate(fmt.Sprint(e.Values), maxContentCharacters),
	)

	if e.Err != nil {
		return fmt.Sprintf("%s, err=%s", s, e.Err.Error())
	}
	return s
}

// Unwrap returns the wrapped error (Go v1.13)
func (e *MissingKeyError) Unwrap() error {
	return e.Err
}

// Error prints out the error and any related details
func (e *ApplyValueError) Error() string {
	s := fmt.Sprintf(
		`ApplyValueError: content="%s", values=%s`,
		truncate(e.Content, maxContentCharacters),
		truncate(fmt.Sprint(e.Values), maxContentCharacters),
	)

	if e.Err != nil {
		return fmt.Sprintf("%s, err=%s", s, e.Err.Error())
	}
	return s
}

// Unwrap returns the wrapped error (Go v1.13)
func (e *ApplyValueError) Unwrap() error {
	return e.Err
}
