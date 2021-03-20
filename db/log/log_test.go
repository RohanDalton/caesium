package log

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/caesium-cloud/caesium/db/store/badger"
	"github.com/hashicorp/raft"
)

func Test_LogNewEmpty(t *testing.T) {
	path := mustTmpDir()
	defer os.Remove(path)

	l, err := NewLog(path)
	if err != nil {
		t.Fatalf("failed to create log: %s", err)
	}
	fi, err := l.FirstIndex()
	if err != nil {
		t.Fatalf("failed to get first index: %s", err)
	}
	if fi != 0 {
		t.Fatalf("got non-zero value for first index of empty log: %d", fi)
	}

	li, err := l.LastIndex()
	if err != nil {
		t.Fatalf("failed to get last index: %s", err)
	}
	if li != 0 {
		t.Fatalf("got non-zero value for last index of empty log: %d", li)
	}

	lci, err := l.LastCommandIndex()
	if err != nil {
		t.Fatalf("failed to get last command index: %s", err)
	}
	if lci != 0 {
		t.Fatalf("got wrong value for last command index of not empty log: %d", lci)
	}

}

func Test_LogNewExistNotEmpty(t *testing.T) {
	path := mustTmpDir()
	defer os.Remove(path)

	// Write some entries directory to the BoltDB Raft store.
	bs, err := badger.NewBadgerStore(path)
	if err != nil {
		t.Fatalf("failed to create badger store: %s", err)
	}
	for i := 4; i > 0; i-- {
		if err := bs.StoreLog(&raft.Log{
			Index: uint64(i),
		}); err != nil {
			t.Fatalf("failed to write entry to raft log: %s", err)
		}
	}
	if err := bs.Close(); err != nil {
		t.Fatalf("failed to close badger db: %s", err)
	}

	l, err := NewLog(path)
	if err != nil {
		t.Fatalf("failed to create new log: %s", err)
	}

	fi, err := l.FirstIndex()
	if err != nil {
		t.Fatalf("failed to get first index: %s", err)
	}
	if fi != 1 {
		t.Fatalf("got wrong value for first index of empty log: %d", fi)
	}

	li, err := l.LastIndex()
	if err != nil {
		t.Fatalf("failed to get last index: %s", err)
	}
	if li != 4 {
		t.Fatalf("got wrong value for last index of not empty log: %d", li)
	}

	lci, err := l.LastCommandIndex()
	if err != nil {
		t.Fatalf("failed to get last command index: %s", err)
	}
	if lci != 4 {
		t.Fatalf("got wrong value for last command index of not empty log: %d", lci)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("failed to close log: %s", err)
	}

	// Delete an entry, recheck index functionality.
	bs, err = badger.NewBadgerStore(path)
	if err != nil {
		t.Fatalf("failed to re-open badger store: %s", err)
	}
	if err := bs.DeleteRange(1, 1); err != nil {
		t.Fatalf("failed to delete range: %s", err)
	}
	if err := bs.Close(); err != nil {
		t.Fatalf("failed to close badger db: %s", err)
	}

	l, err = NewLog(path)
	if err != nil {
		t.Fatalf("failed to create new log: %s", err)
	}

	fi, err = l.FirstIndex()
	if err != nil {
		t.Fatalf("failed to get first index: %s", err)
	}
	if fi != 2 {
		t.Fatalf("got wrong value for first index of empty log: %d", fi)
	}

	li, err = l.LastIndex()
	if err != nil {
		t.Fatalf("failed to get last index: %s", err)
	}
	if li != 4 {
		t.Fatalf("got wrong value for last index of empty log: %d", li)
	}

	fi, li, err = l.Indexes()
	if err != nil {
		t.Fatalf("failed to get indexes: %s", err)
	}
	if fi != 2 {
		t.Fatalf("got wrong value for first index of empty log: %d", fi)
	}
	if li != 4 {
		t.Fatalf("got wrong value for last index of empty log: %d", li)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("failed to close log: %s", err)
	}
}

func Test_LogLastCommandIndexNotExist(t *testing.T) {
	path := mustTmpDir()
	defer os.Remove(path)

	// Write some entries directory to the BoltDB Raft store.
	bs, err := badger.NewBadgerStore(path)
	if err != nil {
		t.Fatalf("failed to create badger store: %s", err)
	}
	for i := 4; i > 0; i-- {
		if err := bs.StoreLog(&raft.Log{
			Index: uint64(i),
			Type:  raft.LogNoop,
		}); err != nil {
			t.Fatalf("failed to write entry to raft log: %s", err)
		}
	}
	if err := bs.Close(); err != nil {
		t.Fatalf("failed to close badger db: %s", err)
	}

	l, err := NewLog(path)
	if err != nil {
		t.Fatalf("failed to create new log: %s", err)
	}

	fi, err := l.FirstIndex()
	if err != nil {
		t.Fatalf("failed to get first index: %s", err)
	}
	if fi != 1 {
		t.Fatalf("got wrong value for first index of empty log: %d", fi)
	}

	li, err := l.LastIndex()
	if err != nil {
		t.Fatalf("failed to get last index: %s", err)
	}
	if li != 4 {
		t.Fatalf("got wrong for last index of not empty log: %d", li)
	}

	lci, err := l.LastCommandIndex()
	if err != nil {
		t.Fatalf("failed to get last command index: %s", err)
	}
	if lci != 0 {
		t.Fatalf("got wrong value for last command index of not empty log: %d", lci)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("failed to close log: %s", err)
	}

	// Delete first log.
	bs, err = badger.NewBadgerStore(path)
	if err != nil {
		t.Fatalf("failed to re-open badger store: %s", err)
	}
	if err := bs.DeleteRange(1, 1); err != nil {
		t.Fatalf("failed to delete range: %s", err)
	}
	if err := bs.Close(); err != nil {
		t.Fatalf("failed to close badger db: %s", err)
	}

	l, err = NewLog(path)
	if err != nil {
		t.Fatalf("failed to create new log: %s", err)
	}

	lci, err = l.LastCommandIndex()
	if err != nil {
		t.Fatalf("failed to get last command index: %s", err)
	}
	if lci != 0 {
		t.Fatalf("got wrong value for last command index of not empty log: %d", lci)
	}
}

// mustTmpDir returns a path to a temporary file in directory dir. It is up to the
// caller to remove the file once it is no longer needed.
func mustTmpDir() string {
	tmpDir, err := ioutil.TempDir("", "rqlite-db-test")
	if err != nil {
		panic(err.Error())
	}
	return tmpDir
}