package idgen

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/armistcxy/shorten/internal/domain"
	"go.etcd.io/bbolt"
)

// Implement IDGenerator interface with manage partittion strategy
type ManageIDGenerator struct {
	pm *PartitionManager
}

func NewManageIDGenerator(pm *PartitionManager) *ManageIDGenerator {
	return &ManageIDGenerator{
		pm: pm,
	}
}

func (mg *ManageIDGenerator) GenerateID() string {
	id := mg.pm.GiveID()
	encodeID := domain.EncodeID(id)
	return encodeID
}

type PartitionManager struct {
	mu         sync.Mutex
	partitions []Partition
	lastStart  int64
}

const (
	DEFAULT_PARTITION_AMOUNT = 12 // Number of partitions that partition manager holds
	PARTITION_SIZE           = 1 << 25
)

var KV_STORAGE_PATH = os.Getenv("KV_STORAGE_PATH")

func NewPartitionManager() (*PartitionManager, error) {
	partitions := make([]Partition, DEFAULT_PARTITION_AMOUNT)
	db, err := bbolt.Open(KV_STORAGE_PATH, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	// load partitions from KV storage
	var (
		lastStart int64 = 0
		possible  bool  = false
		used      int
	)
	for i := range partitions {
		for !possible {
			possible, used = checkPossiblePartition(db, lastStart)
			if !possible {
				lastStart += PARTITION_SIZE
			}
		}
		partitions[i] = newPartition(lastStart, used)
		lastStart += PARTITION_SIZE
	}

	return &PartitionManager{
		mu:         sync.Mutex{},
		partitions: partitions,
		lastStart:  lastStart,
	}, nil
}

func checkPossiblePartition(db *bbolt.DB, start int64) (bool, int) {
	var (
		bucketName []byte = []byte(fmt.Sprintf("short-%d", start/PARTITION_SIZE))
		l, r, mid  int64  = start, start + PARTITION_SIZE - 1, 0
		exist      bool
		err        error
	)

	// First check whether bucket exist

	err = db.Update(func(tx *bbolt.Tx) error {
		_, createError := tx.CreateBucketIfNotExists(bucketName)
		if createError != nil {
			return err
		}
		return nil
	})

	// Failed when creating bucket
	if err != nil {
		slog.Error("failed to create bucket", "error", err.Error())
		return false, 0
	}

	// From here bucket has been created
	// Binary search to find first non-used ID
	for l < r {
		mid = (l + r) / 2
		// check whether id mid has been used
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(bucketName)
			v := b.Get([]byte(strconv.FormatInt(mid, 10)))
			exist = v != nil
			return nil
		})

		if err != nil {
			slog.Warn("check possible partition", "error", err.Error())
			continue // perform the operation again
		}

		if exist {
			l = mid + 1
		} else {
			r = mid
		}
	}

	// check whether id is suitable
	for {
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(bucketName)
			if b == nil {
				return nil // bucket not even exist
			}
			v := b.Get([]byte(strconv.FormatInt(l, 10)))
			exist = v != nil
			return nil
		})

		if err != nil {
			slog.Warn("check possible partition", "error", err.Error())
		} else {
			break
		}
	}

	return !exist, int(l - start)
}

func (pm *PartitionManager) GiveID() uint64 {
	// change this later for better distribute
	paID := rand.IntN(DEFAULT_PARTITION_AMOUNT)

	id, success := pm.partitions[paID].giveID()
	// assign new partition if already used all value in partition
	if !success {
		pm.mu.Lock()
		pm.partitions[paID] = newPartition(pm.lastStart+PARTITION_SIZE, 0)
		pm.lastStart += PARTITION_SIZE
		pm.mu.Unlock()
		id, _ = pm.partitions[paID].giveID()
	}

	return uint64(id)
}

// Partition represents a partition of IDs, with a bucket name, a starting ID, and the number of IDs used.
type Partition struct {
	mu         sync.Mutex
	bucketName string
	start      int64
	used       int
}

func newPartition(start int64, used int) Partition {
	bucketID := start / PARTITION_SIZE
	return Partition{
		mu:         sync.Mutex{},
		bucketName: fmt.Sprintf("short-%d", bucketID),
		start:      start,
		used:       used,
	}
}

func (p *Partition) giveID() (int64, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.used == PARTITION_SIZE {
		return -1, false
	}

	p.used++
	return p.start + int64(p.used-1), true
}

// Information about Partition is used both by IDGenerator and BBoltRepository

// IDGenerator needs: new id
// BBoltRepository need: get bucketName that contains ID
// actually not that hard => just decode the encodeID
// => get the ID and bucketName = "short" + strconv.Itoa(id / PARTITION SIZE)
