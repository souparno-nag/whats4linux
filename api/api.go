package api

import (
	"context"
	"fmt"

	"github.com/lugvitc/whats4linux/internal/mstore"
	"github.com/lugvitc/whats4linux/internal/wa"
	"github.com/nyaruka/phonenumbers"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type Group struct {
	JID              string `json:"jid"`
	Name             string `json:"name"`
	Topic            string `json:"topic"`
	OwnerJID         string `json:"owner_jid"`
	ParticipantCount int    `json:"participant_count"`
}

type Contact struct {
	JID        string `json:"jid"`
	Short      string `json:"short"`
	FullName   string `json:"full_name"`
	PushName   string `json:"push_name"`
	IsBusiness bool   `json:"is_business"`
}

// Api struct
type Api struct {
	ctx          context.Context
	waClient     *whatsmeow.Client
	messageStore *mstore.MessageStore
}

// NewApi creates a new Api application struct
func New() *Api {
	return &Api{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
	a.waClient = wa.NewClient(ctx)
	a.messageStore = mstore.NewMessageStore()
}

func (a *Api) Login() error {
	var err error
	a.waClient.AddEventHandler(a.mainEventHandler)
	if a.waClient.Store.ID == nil {
		qrChan, _ := a.waClient.GetQRChannel(a.ctx)
		err = a.waClient.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				runtime.EventsEmit(a.ctx, "wa:qr", evt.Code)
			} else {
				runtime.EventsEmit(a.ctx, "wa:status", evt.Event)
			}
		}
	} else {
		runtime.EventsEmit(a.ctx, "wa:status", "logged_in")
		fmt.Println("Already logged in, connecting...")
		// Already logged in, just connect
		err = a.waClient.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Api) FetchGroups() ([]Group, error) {
	groups, err := a.waClient.GetJoinedGroups(a.ctx)
	if err != nil {
		return nil, err
	}

	var result []Group
	for _, g := range groups {
		result = append(result, Group{
			JID:              g.JID.String(),
			Name:             g.Name,
			Topic:            g.Topic,
			OwnerJID:         g.OwnerJID.String(),
			ParticipantCount: len(g.Participants),
		})
	}
	return result, nil
}

func (a *Api) FetchContacts() ([]Contact, error) {
	rawContacts, err := a.waClient.Store.Contacts.GetAllContacts(a.ctx)
	if err != nil {
		return nil, err
	}
	contacts := make([]Contact, 0, len(rawContacts))

	var result []Contact
	for jid, c := range rawContacts {
		rawNum := "+" + jid.User
		// Parse phone number to use as International Format
		num, err := phonenumbers.Parse(rawNum, "")
		if err != nil && !phonenumbers.IsValidNumber(num) {
			continue
		}

		contacts = append(contacts, Contact{
			JID:        phonenumbers.Format(num, phonenumbers.INTERNATIONAL),
			FullName:   c.FullName,
			Short:      c.FirstName,
			PushName:   c.PushName,
			IsBusiness: c.BusinessName != "",
		})
	}
	return result, nil
}

func (a *Api) mainEventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		a.messageStore.ProcessMessageEvent(v)
	default:
		// Ignore other events for now
	}
}
