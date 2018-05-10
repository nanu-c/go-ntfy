package main

import (
	"encoding/json"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/nanu-c/qml-go"
	dbus "launchpad.net/go-dbus/v1"
)

func main() {
	err := qml.Run(run)
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	engine := qml.NewEngine()
	component, err := engine.LoadFile("qml/main.qml")
	if err != nil {
		return err
	}

	testvar := TestStruct{Message: "Testmessage", Number: 1}
	context := engine.Context()
	context.SetVar("testvar", &testvar)
	testvar.GetMessage()
	win := component.CreateWindow(nil)
	testvar.Root = win.Root()
	win.Show()
	win.Wait()
	return nil
}

type TestStruct struct {
	Root    qml.Object
	Message string
	Output  string
	Number  int
}

func (testvar *TestStruct) GetMessage() {
	log.Print("fetch second Message...")
	testvar.notification()
}

var (
	sessionBus *dbus.Connection
	err        error
	Nh         *NotificationHandler
)

func (testvar *TestStruct) notification() {

	if sessionBus, err = dbus.Connect(dbus.SessionBus); err != nil {
		log.Fatal("Connection error: ", err)
	}
	Nh = NewLegacyHandler(sessionBus, "nanuc.gontfy_gontfy")
	n := Nh.NewStandardPushMessage(
		"Go notify",
		"test", "")
	if err := Nh.Send(n); err != nil {
		testvar.Output = err.Error()
		log.Printf(err.Error())
	} else {
		testvar.Output = "no error"
	}
	qml.Changed(testvar, &testvar.Output)
}

/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * ciborium is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * nuntium is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// Notifications lives on a well-knwon bus.Address

const (
	dbusName       = "com.ubuntu.Postal"
	dbusInterface  = "com.ubuntu.Postal"
	dbusPathPart   = "/com/ubuntu/Postal/"
	dbusPostMethod = "Post"
)

type VariantMap map[string]dbus.Variant

type NotificationHandler struct {
	dbusObject  *dbus.ObjectProxy
	application string
}

func NewLegacyHandler(conn *dbus.Connection, application string) *NotificationHandler {
	return &NotificationHandler{
		dbusObject:  conn.Object(dbusName, "/com/ubuntu/Postal/nanuc_2Egontfy"),
		application: application,
	}
}

func (n *NotificationHandler) Send(m *PushMessage) error {
	var pushMessage string
	if out, err := json.Marshal(m); err == nil {
		pushMessage = string(out)
	} else {
		return err
	}
	_, err := n.dbusObject.Call(dbusInterface, dbusPostMethod, "nanuc.gontfy_gontfy", pushMessage)
	return err
}

// NewStandardPushMessage creates a base Notification with common
// components (members) setup.
func (n *NotificationHandler) NewStandardPushMessage(summary, body, icon string) *PushMessage {
	pm := &PushMessage{
		Message: "foobar",
		Notification: Notification{
			Card: &Card{
				Summary: summary,
				Body:    body,
				Actions: []string{"appid://nanuc.gontfy/gontfy/current-user-version"},
				// Icon:    icon,
				Popup:     true,
				Persist:   true,
				Tag:       "chat",
				Timestamp: time.Now().Unix(),
			},
			RawSound:     json.RawMessage(`"sounds/ubuntu/notifications/Slick.ogg"`),
			RawVibration: json.RawMessage(`{"pattern": [100, 100], "repeat": 2}`),
		},
	}
	return pm
}

// PushMessage represents a data structure to be sent over to the
// Post Office. It consists of a Notification and a Message.
type PushMessage struct {
	// Notification (optional) describes the user-facing notifications
	// triggered by this push message.
	Notification Notification `json:"notification,omitempty"`
	Message      string       `json:"message,omitempty"`
}

// Notification (optional) describes the user-facing notifications
// triggered by this push message.
type Notification struct {
	Card         *Card           `json:"card,omitempty"`
	RawSound     json.RawMessage `json:"sound"`
	RawVibration json.RawMessage `json:"vibrate"`
}

// Card is part of a notification and represents the user visible hints for
// a specific notification.
type Card struct {
	// Summary is a required title. The card will not be presented if this is missing.
	Summary string `json:"summary"`
	// Body is the longer text.
	Body string `json:"body,omitempty"`
	// Whether to show a bubble. Users can disable this, and can easily miss
	// them, so don’t rely on it exclusively.
	Popup bool `json:"popup,omitempty"`
	// Actions provides actions for the bubble's snap decissions.
	Actions []string `json:"actions,omitempty"`
	// Icon is a path to an icon to display with the notification bubble.
	// Icon string `json:"icon,omitempty"`
	// Whether to show in notification centre.
	Persist   bool   `json:"persist,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Timestamp int64  `json:"timestamp"`
}
