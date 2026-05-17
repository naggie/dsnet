package jsonfile

import (
	"context"
	"net/url"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConcurrentSaveSerialises(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dsnetconfig.json")
	openOne := func() *Backend {
		be, err := Open(&url.URL{Scheme: "jsonfile", Path: path})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		return be.(*Backend)
	}

	// Seed the file so both Backends start from a non-empty State.
	seed := openOne()
	if err := seed.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	_ = seed.Close()

	beA := openOne()
	defer beA.Close()
	beB := openOne()
	defer beB.Close()

	stateA, _, err := beA.Load(context.Background())
	if err != nil {
		t.Fatalf("beA.Load: %v", err)
	}
	stateB, _, err := beB.Load(context.Background())
	if err != nil {
		t.Fatalf("beB.Load: %v", err)
	}

	// Append distinct peers in each State (use NewPeer to avoid duplicate-key
	// rejection; the goal is just to get two non-trivial concurrent writes).
	srvA := stateA.Networks["dsnet"].Server
	srvB := stateB.Networks["dsnet"].Server
	srvA.ListenPort = 51821
	srvB.ListenPort = 51822

	var wg sync.WaitGroup
	wg.Add(2)
	errs := make(chan error, 2)
	go func() {
		defer wg.Done()
		// version "" means no optimistic check; both writes must complete.
		errs <- beA.Save(context.Background(), stateA, "")
	}()
	go func() {
		defer wg.Done()
		errs <- beB.Save(context.Background(), stateB, "")
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Save: %v", err)
		}
	}

	// One of the two ListenPorts must have won; the file must be valid.
	beC := openOne()
	defer beC.Close()
	state, _, err := beC.Load(context.Background())
	if err != nil {
		t.Fatalf("final Load: %v", err)
	}
	final := state.Networks["dsnet"].Server.ListenPort
	if final != 51821 && final != 51822 {
		t.Fatalf("unexpected final ListenPort: %d", final)
	}
}

func TestSaveBlocksUntilReadReleases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dsnetconfig.json")
	open := func() *Backend {
		be, err := Open(&url.URL{Scheme: "jsonfile", Path: path})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		return be.(*Backend)
	}

	seed := open()
	if err := seed.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	_ = seed.Close()

	reader := open()
	defer reader.Close()
	if err := reader.acquireRead(context.Background()); err != nil {
		t.Fatalf("acquireRead: %v", err)
	}

	writer := open()
	defer writer.Close()

	saveErr := make(chan error, 1)
	go func() {
		saveErr <- writer.Save(context.Background(), wrapState(sampleServer(t)), "")
	}()

	select {
	case err := <-saveErr:
		t.Fatalf("Save completed while read lock was held: err=%v", err)
	case <-time.After(150 * time.Millisecond):
	}

	reader.releaseLock()
	select {
	case err := <-saveErr:
		if err != nil {
			t.Fatalf("Save after release: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Save did not complete after read lock released")
	}
}
