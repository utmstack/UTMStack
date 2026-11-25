package socai

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestAlertsQueueUntilTheServiceTakesThem(t *testing.T) {
	srv, err := server.NewServer(&server.Options{Port: -1, JetStream: true, StoreDir: t.TempDir(), NoLog: true, NoSigs: true})
	if err != nil {
		t.Fatal(err)
	}

	go srv.Start()

	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("NATS did not start")
	}

	defer srv.Shutdown()

	for deadline := time.Now().Add(20 * time.Second); !srv.JetStreamEnabled(); {
		if time.Now().After(deadline) {
			t.Fatal("JetStream did not start")
		}

		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()

	pub, err := NewJetStreamPublisher(ctx, srv.ClientURL(), 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer pub.Close()

	// Published while nobody is listening.
	if err := pub.Publish(ctx, AlertsSubject, []byte(`{"alertId":"a1"}`)); err != nil {
		t.Fatal(err)
	}

	// The service starts later and finds it waiting.
	nc, _ := nats.Connect(srv.ClientURL())
	defer nc.Close()

	js, _ := jetstream.New(nc)
	cons, err := js.CreateOrUpdateConsumer(ctx, AlertsStream, jetstream.ConsumerConfig{Durable: "soc-ai", AckPolicy: jetstream.AckExplicitPolicy})
	if err != nil {
		t.Fatal(err)
	}

	batch, err := cons.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	n := 0
	for m := range batch.Messages() {
		n++

		if string(m.Data()) != `{"alertId":"a1"}` {
			t.Errorf("got %q", m.Data())
		}

		_ = m.Ack()
	}

	if n != 1 {
		t.Errorf("the alert did not wait for the service: %d messages", n)
	}
}

func TestAStreamAlreadyThereIsReused(t *testing.T) {
	srv, err := server.NewServer(&server.Options{Port: -1, JetStream: true, StoreDir: t.TempDir(), NoLog: true, NoSigs: true})
	if err != nil {
		t.Fatal(err)
	}

	go srv.Start()

	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("NATS did not start")
	}

	defer srv.Shutdown()

	for deadline := time.Now().Add(20 * time.Second); !srv.JetStreamEnabled(); {
		if time.Now().After(deadline) {
			t.Fatal("JetStream did not start")
		}

		time.Sleep(50 * time.Millisecond)
	}

	for i := 0; i < 2; i++ {
		p, err := NewJetStreamPublisher(context.Background(), srv.ClientURL(), 10*time.Second)
		if err != nil {
			t.Fatalf("start %d: %v", i, err)
		}

		p.Close()
	}
}
