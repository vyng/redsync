package redsync

import (
	"time"
	"github.com/gomodule/redigo/redis"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration

// A Mutex is a distributed mutual exclusion lock.
type Mutex struct {
	name   string
	expiry time.Duration

	tries     int
	delayFunc DelayFunc

	factor float64

	quorum int

	genValueFunc func() (string, error)
	value        string
	until        time.Time

	pools []Pool
}

// Lock locks m with a specific value.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	value, err := m.genValueFunc()
	if err != nil {
		return err
	}

	for i := 0; i < m.tries; i++ {
		if i != 0 {
			time.Sleep(m.delayFunc(i))
		}

		start := time.Now()

		n := m.actOnPoolsAsync(func(pool Pool, m *Mutex) bool {
			return m.acquire(pool, value)
		})

		end := time.Now()
		drift := time.Duration(int64(float64(m.expiry) * m.factor)) + 2
		validTime := m.expiry - end.Sub(start) - drift

		if n >= m.quorum && validTime > 0 {
			m.value = value
			m.until = end.Add(validTime)
			return nil
		}

		m.actOnPoolsAsync(func(pool Pool, m *Mutex) bool {
			return m.release(pool, value)
		})
	}

	return ErrFailed
}

// Unlock unlocks m and returns the status of unlock.
func (m *Mutex) Unlock() bool {
	return m.actOnPoolsAsync(func(pool Pool, m *Mutex) bool {
		return m.release(pool, m.value)
	}) > m.quorum
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *Mutex) Extend() bool {
	return m.actOnPoolsAsync(func(pool Pool, m *Mutex) bool {
		return m.touch(pool, m.value, int64(m.expiry / time.Millisecond))
	}) >= m.quorum

}

func (m *Mutex) acquire(pool Pool, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	reply, err := redis.String(
		conn.Do("SET", m.name, value, "NX", "PX", int64(m.expiry / time.Millisecond)),
	)
	return err == nil && reply == "OK"
}

var deleteScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (m *Mutex) release(pool Pool, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	status, err := deleteScript.Do(conn, m.name, value)
	return err == nil && status != 0
}

var touchScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("pexpire", KEYS[1], ARGV[2])
	else
		return 0
	end
`)

func (m *Mutex) touch(pool Pool, value string, expiry int64) bool {
	conn := pool.Get()
	defer conn.Close()
	status, err := touchScript.Do(conn, m.name, value, expiry)
	return err == nil && status != 0
}

func (m *Mutex) actOnPoolsAsync(actFn func(Pool, *Mutex) bool) int {
	ch := make(chan bool, len(m.pools))
	for _, pool := range m.pools {
		go func(pool Pool) {
			ch <- actFn(pool, m)
		}(pool)
	}
	n := 0
	for range m.pools {
		if <-ch {
			n++
		}
	}
	return n
}
