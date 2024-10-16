package idgen

import (
	"testing"
)

func TestCreatePartitionManager(t *testing.T) {
	pm, err := NewPartitionManager()
	if err != nil {
		t.Error(err)
	}

	var lastStart int64 = -1
	for i := range pm.partitions {
		t.Logf("Partition %d start with %d", i, pm.partitions[i].start)
		if pm.partitions[i].start == lastStart {
			t.Errorf("Partitions shouldn't have the same start = %d", lastStart)
		}
		lastStart = pm.partitions[i].start
	}
}
