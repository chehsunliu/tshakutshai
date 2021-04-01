package tpex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chehsunliu/tshakutshai/pkg/client/tpex"
)

var client = tpex.NewClient(time.Second * 3)

func TestClient_FetchDayQuotes(t *testing.T) {
	date := time.Date(2021, time.March, 30, 0, 0, 0, 0, time.UTC)

	qs, err := client.FetchDayQuotes(date)
	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 6877, len(qs))

	q1 := qs["006201"]
	assert.Equal(t, "006201", q1.Code)
	assert.Equal(t, "元大富櫃50", q1.Name)
	assert.Equal(t, "20210330", q1.Date.Format("20060102"))
	assert.Equal(t, uint64(54_765), q1.Volume)
	assert.Equal(t, uint64(33), q1.Transactions)
	assert.Equal(t, uint64(1_062_607), q1.Value)
	assert.Equal(t, 19.51, q1.High)
	assert.Equal(t, 19.32, q1.Low)
	assert.Equal(t, 19.37, q1.Open)
	assert.Equal(t, 19.49, q1.Close)

	q2 := qs["8044"]
	assert.Equal(t, "8044", q2.Code)
	assert.Equal(t, "網家", q2.Name)
	assert.Equal(t, "20210330", q2.Date.Format("20060102"))
	assert.Equal(t, uint64(766_431), q2.Volume)
	assert.Equal(t, uint64(698), q2.Transactions)
	assert.Equal(t, uint64(67_759_797), q2.Value)
	assert.Equal(t, 90.00, q2.High)
	assert.Equal(t, 88.00, q2.Low)
	assert.Equal(t, 89.40, q2.Open)
	assert.Equal(t, 88.00, q2.Close)
}

func TestClient_FetchDailyQuotes(t *testing.T) {
	qs, err := client.FetchDailyQuotes("8044", 2020, time.March)
	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 22, len(qs))

	q0 := qs[0]
	assert.Equal(t, "8044", q0.Code)
	assert.Equal(t, "網家", q0.Name)
	assert.Equal(t, "20200302", q0.Date.Format("20060102"))
	assert.Equal(t, uint64(815_000), q0.Volume)
	assert.Equal(t, uint64(670), q0.Transactions)
	assert.Equal(t, uint64(84_369_000), q0.Value)
	assert.Equal(t, 105.50, q0.High)
	assert.Equal(t, 101.00, q0.Low)
	assert.Equal(t, 104.50, q0.Open)
	assert.Equal(t, 102.50, q0.Close)

	q1 := qs[21]
	assert.Equal(t, "8044", q1.Code)
	assert.Equal(t, "網家", q1.Name)
	assert.Equal(t, "20200331", q1.Date.Format("20060102"))
	assert.Equal(t, uint64(875_000), q1.Volume)
	assert.Equal(t, uint64(750), q1.Transactions)
	assert.Equal(t, uint64(61_249_000), q1.Value)
	assert.Equal(t, 71.50, q1.High)
	assert.Equal(t, 69.00, q1.Low)
	assert.Equal(t, 70.90, q1.Open)
	assert.Equal(t, 70.30, q1.Close)
}
