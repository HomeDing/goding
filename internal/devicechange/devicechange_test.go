// devicechange_test.go - Tests for shared device-change subscriptions
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package devicechange

import "testing"

func TestNotifySubscribers(t *testing.T) {
	notifications, unsubscribe := Subscribe()
	defer unsubscribe()

	Notify()
	select {
	case <-notifications:
	default:
		t.Fatal("expected device-change notification")
	}
}

func TestUnsubscribeStopsNotifications(t *testing.T) {
	notifications, unsubscribe := Subscribe()
	unsubscribe()

	Notify()
	select {
	case <-notifications:
		t.Fatal("received notification after unsubscribe")
	default:
	}
}

// End.
