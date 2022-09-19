package eventd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type testReceiver struct {
	c chan interface{}
}

func (r testReceiver) Receiver() chan<- interface{} {
	return r.c
}

func TestEventdMonitor(t *testing.T) {
	ed, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := ed.NewEmbeddedClient()

	livenessFactory := liveness.EtcdFactory(context.Background(), client)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)

	if err := bus.Start(); err != nil {
		assert.FailNow(t, "message bus failed to start")
	}

	eventChan := make(chan interface{}, 2)

	subscriber := testReceiver{
		c: eventChan,
	}
	sub, err := bus.Subscribe(messaging.TopicEvent, "testReceiver", subscriber)
	if err != nil {
		assert.FailNow(t, "failed to subscribe to message bus topic event")
	}

	eventStore, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	store := etcdstorev2.NewStore(eventStore.Client)
	nsStore := etcdstorev2.NewNamespaceStore(eventStore.Client)

	if err := seeds.SeedInitialDataWithContext(context.Background(), store, nsStore); err != nil {
		assert.FailNow(t, err.Error())
	}

	e := newEventd(store, eventStore, bus, livenessFactory)

	if err := e.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Interval = 1
	event.Check.Ttl = 5

	wrappedEntity, err := wrap.V2Resource(event.Entity)
	if err != nil {
		t.Fatal(err)
	}

	req := storev2.NewResourceRequestFromV2Resource(event.Entity)
	if err := store.CreateOrUpdate(context.Background(), req, wrappedEntity); err != nil {
		t.Fatal(err)
	}

	if err := bus.Publish(messaging.TopicEventRaw, event); err != nil {
		assert.FailNow(t, "failed to publish event to TopicEventRaw")
	}

	msg, ok := <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}

	okEvent, ok := msg.(*corev2.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(0), okEvent.Check.Status)

	msg, ok = <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}
	warnEvent, ok := msg.(*corev2.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(1), warnEvent.Check.Status)

	assert.NoError(t, sub.Cancel())
	close(eventChan)
}
