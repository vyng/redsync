package redsync

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server

func TestMain(m *testing.M) {
	for i := 0; i < 8; i++ {
		server, err := tempredis.Start(tempredis.Config{})
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	result := m.Run()
	for _, server := range servers {
		if err := server.Term(); err != nil {
			panic("Testing server errored during termination")
		}
	}
	os.Exit(result)
}

func TestRedsync(t *testing.T) {
	pools := newMockPools(8, servers)
	rs := New(pools)

	mutex := rs.NewMutex("test-redsync")
	err := mutex.Lock()
	assert.NoError(t, err, "Mutex errored during locking")

	assertAcquired(t, pools, mutex)
}
