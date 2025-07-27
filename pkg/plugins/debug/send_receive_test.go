// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package debug_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
)

func TestSendReceive(t *testing.T) {
	a := NewWithT(t)
	p, err := debug.NewPlugin()
	a.Expect(err).To(BeNil())

	logger := zerolog.Nop()
	r, err := p.NewReceiver(tid, "debug", "mydebug", "", nil, syncer.NewInMemoryDeltaSyncer(&logger, viper.New()))

	a.Expect(err).To(BeNil())

	s, err := p.NewSender(tid, "debug", "mydebug", "", nil, nil)
	a.Expect(err).To(BeNil())

	err = r.Receive(s.Send)
	a.Expect(err).To(BeNil())
	r.StopReceiving(context.Background())
}
