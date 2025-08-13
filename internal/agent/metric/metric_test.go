package metric

import (
	"encoding/json"
	"strconv"
	"testing"
	"unsafe"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	am := New()

	require.NotNil(t, am)
	assert.Equal(t, int64(0), am.PollCount)
	assert.Len(t, am.Metrics, len(gaugeNames)+len(counterNames))

	for _, name := range gaugeNames {
		m, exists := am.Metrics[name]
		assert.True(t, exists, "gauge метрика %s должна существовать", name)
		assert.Equal(t, entity.Gauge, m.Type)
		assert.Equal(t, "0", m.Value)
	}

	for _, name := range counterNames {
		m, exists := am.Metrics[name]
		assert.True(t, exists, "counter метрика %s должна существовать", name)
		assert.Equal(t, entity.Counter, m.Type)
		assert.Equal(t, "0", m.Value)
	}
}

func TestUpdate(t *testing.T) {
	am := New()

	initialPollCount := am.PollCount

	am.Update()

	assert.Equal(t, initialPollCount+1, am.PollCount)
	counterMetric := am.Metrics["PollCount"]
	assert.Equal(t, strconv.FormatInt(am.PollCount, 10), counterMetric.Value)

	alloc := am.Metrics["Alloc"]
	allocVal, err := strconv.ParseUint(alloc.Value, 10, 64)
	require.NoError(t, err)
	assert.Greater(t, allocVal, uint64(0), "Alloc должен быть > 0 после Update")

	random := am.Metrics["RandomValue"]
	randomVal, err := strconv.ParseFloat(random.Value, 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, randomVal, 0.0)
	assert.Less(t, randomVal, 1.0)
}

func TestUpdate_MultipleCalls(t *testing.T) {
	am := New()

	am.Update()

	assert.Equal(t, int64(1), am.PollCount)
	assert.Equal(t, "1", am.Metrics["PollCount"].Value)
}

func TestGetByName_Exists(t *testing.T) {
	am := New()
	am.Update() // чтобы Alloc был > 0

	metric := am.GetByName("Alloc")
	assert.Equal(t, "Alloc", metric.Name)
	assert.Equal(t, entity.Gauge, metric.Type)
	assert.NotEqual(t, "0", metric.Value)
}

func TestGetByName_NotExists(t *testing.T) {
	am := New()
	metric := am.GetByName("unknown")
	assert.Equal(t, Metric{}, metric)
}

func TestGetAllGaugeNames(t *testing.T) {
	am := New()
	names := am.GetAllGaugeNames()
	assert.ElementsMatch(t, gaugeNames, names)
	assert.Len(t, names, len(gaugeNames))
}

func TestGetAllCounterNames(t *testing.T) {
	am := New()
	names := am.GetAllCounterNames()
	assert.ElementsMatch(t, counterNames, names)
	assert.Len(t, names, len(counterNames))
}

func TestClearCounter(t *testing.T) {
	am := New()
	am.Update()
	am.Update()

	assert.Equal(t, int64(2), am.PollCount)
	assert.Equal(t, "2", am.Metrics["PollCount"].Value)

	am.ClearCounter("PollCount")

	assert.Equal(t, int64(0), am.PollCount)
	assert.Equal(t, "0", am.Metrics["PollCount"].Value)
}

func TestClearCounter_IgnoredForOther(t *testing.T) {
	am := New()
	am.Update()
	am.ClearCounter("SomeOther")
	// Должно не изменить состояние
	assert.Equal(t, int64(1), am.PollCount)
}

func TestMetric_JSONTags(t *testing.T) {
	var m Metric
	typ := unsafe.Sizeof(m)
	_ = typ // избежать unused

	m = Metric{
		Type:  "gauge",
		Name:  "test_gauge",
		Value: "3.14",
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	expected := `{"type":"gauge","name":"test_gauge","value":"3.14"}`
	assert.JSONEq(t, expected, string(data))
}
