package hub

import (
	"testing"
	"time"
)

func TestBroadcastSkipsSenderAndDelivers(t *testing.T) {
	h := New(1)

	chA, leaveA := h.Join("user1", "A")
	defer leaveA()

	chB, leaveB := h.Join("user1", "B")
	defer leaveB()

	h.Broadcast("user1", "A", []byte("hola"))

	select {
	case <-chA:
		t.Fatal("el emisor no debería recibir su propio mensaje")
	default:
	}

	select {
	case got := <-chB:
		if string(got) != "hola" {
			t.Fatalf("esperaba 'hola', obtuve %q", got)
		}
	default:
		t.Fatal("B no recibió el mensaje")
	}
}

func TestBroadcastDoesNotBlockWhenReceiverFull(t *testing.T) {
	h := New(1)

	chB, leaveB := h.Join("user1", "B")
	defer leaveB()

	h.Broadcast("user1", "A", []byte("x"))

	done := make(chan struct{})

	go func() {
		h.Broadcast("user1", "A", []byte("y"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Broadcast se bloqueo con canal lleno")
	}

	var got []byte
	select {
	case got = <-chB:
		if string(got) != "x" {
			t.Fatalf("quería 'x', llegó %q", got)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("B no recibió el primer mensaje")
	}

	select {
	case <-chB:
		t.Fatal("debería haberse descartado el segundo mensaje")
	default:
	}
}

func TestLeaveRemovesAndClosesChannel(t *testing.T) {
	h := New(1)

	chA, leaveA := h.Join("user1", "A")
	defer leaveA()

	chB, leaveB := h.Join("user1", "B")

	leaveB()

	select {
	case _, ok := <-chB:
		if ok {
			t.Fatal("chB debería estar cerrado tras leaveB()")
		}
	default:
		t.Fatal("esperaba canal cerrado inmediatamente después de leaveB()")
	}

	h.Broadcast("user1", "A", []byte("hola"))

	select {
	case <-chA:
	default:
	}
}

func TestRoomsAreIsolated(t *testing.T) {
	h := New(1)

	chU1, leaveU1 := h.Join("user1", "A")
	defer leaveU1()

	chU2, leaveU2 := h.Join("user2", "X")
	defer leaveU2()

	h.Broadcast("user1", "A", []byte("hola"))

	select {
	case <-chU2:
		t.Fatal("user2 no debería recibir mensajes de user1")
	default:
	}

	select {
	case <-chU1:
		t.Fatal("el emisor no debería recibir su propio mensaje")
	default:
	}
}
