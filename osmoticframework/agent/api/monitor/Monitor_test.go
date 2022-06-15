package monitor

import (
	"net/http"
	"testing"
	"time"
)

func TestMatrix(t *testing.T) {
	_, err := http.Get(promAddress)
	if err != nil {
		t.Skip("Prometheus not reachable. Skipping")
	}
	metric, err := restQuery("prometheus_http_requests_total[5m]", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := metric.Type
	if mType != MatrixType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, VectorType)
	}
	_, ok := metric.Data.([]Matrix)
	if !ok {
		t.Errorf("Cannot assert to matrix")
	}
}

func TestVector(t *testing.T) {
	_, err := http.Get(promAddress)
	if err != nil {
		t.Skip("Prometheus not reachable. Skipping")
	}
	metric, err := restQuery("prometheus_http_requests_total", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := metric.Type
	if mType != VectorType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, VectorType)
	}
	_, ok := metric.Data.([]Vector)
	if !ok {
		t.Errorf("Cannot assert to vector")
	}
}

func TestScalar(t *testing.T) {
	_, err := http.Get(promAddress)
	if err != nil {
		t.Skip("Prometheus not reachable. Skipping")
	}
	metric, err := restQuery("scalar(prometheus_build_info)", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := metric.Type
	if mType != ScalarType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, VectorType)
	}
	_, ok := metric.Data.(Scalar)
	if !ok {
		t.Errorf("Cannot assert to scalar")
	}
}

func TestParseScalar(t *testing.T) {
	now := time.Now().UnixNano()
	tests := []struct {
		Input    []interface{}
		Expected Scalar
	}{
		{
			Input: []interface{}{
				now,
				"123456789",
			},
			Expected: Scalar{
				Time:      time.Unix(now, 0),
				Value:     123456789,
				Undefined: false,
			},
		},
		{
			Input: []interface{}{
				now,
				"-3.141592654",
			},
			Expected: Scalar{
				Time:      time.Unix(now, 0),
				Value:     -3.141592654,
				Undefined: false,
			},
		},
		{
			Input: []interface{}{
				now,
				"NaN",
			},
			Expected: Scalar{
				Time:      time.Unix(now, 0),
				Value:     0,
				Undefined: true,
			},
		},
	}
	for _, test := range tests {
		scalar, err := parseScalar(test.Input)
		if err != nil {
			t.Error("Cannot parse scalar")
			continue
		}
		if scalar.Time.UnixNano() != test.Expected.Time.UnixNano() {
			t.Errorf("Time not correct. Got %d, Want %d", scalar.Time.Unix(), test.Expected.Time.Unix())
		}
		if scalar.Value != test.Expected.Value {
			t.Errorf("Value not correct. Got %f, Want %f", scalar.Value, test.Expected.Value)
		}
		if scalar.Undefined != test.Expected.Undefined {
			t.Errorf("Undefined not correct. Got %t, Want %t", scalar.Undefined, test.Expected.Undefined)
		}
	}
}

func testCPUEdgeAvg(t *testing.T) {
	_, err := CPUEdgeAvg(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testCPUContainerAvg(t *testing.T) {
	_, err := CPUContainerAvg("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testCPUTimeTotal(t *testing.T) {
	_, err := CPUTimeTotal(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testCPUUtilization(t *testing.T) {
	_, err := CPUUtilization(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testMemoryContainer(t *testing.T) {
	_, err := MemoryContainer("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testMemoryEdge(t *testing.T) {
	_, err := MemoryEdge(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testMemoryContainerPeak(t *testing.T) {
	_, err := MemoryContainerPeak("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testMemoryEdgePeak(t *testing.T) {
	_, err := MemoryEdgePeak(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testMemoryContainerReachLimitSeconds(t *testing.T) {
	_, err := MemoryContainerReachLimitSeconds("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOEdgeTime(t *testing.T) {
	_, err := IOEdgeTime(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOContainerTime(t *testing.T) {
	_, err := IOContainerTime("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOReadEdgeBytes(t *testing.T) {
	_, err := IOReadEdgeBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOReadContainerBytes(t *testing.T) {
	_, err := IOReadContainerBytes("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOWriteEdgeBytes(t *testing.T) {
	_, err := IOWriteEdgeBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOWriteContainerBytes(t *testing.T) {
	_, err := IOWriteContainerBytes("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOFilesystemUsedBytes(t *testing.T) {
	_, err := IOFilesystemUsedBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testIOFilesystemSizeBytes(t *testing.T) {
	_, err := IOFilesystemSizeBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeRxBytes(t *testing.T) {
	_, err := NetworkEdgeRxBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerRxBytes(t *testing.T) {
	_, err := NetworkContainerRxBytes("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeTxBytes(t *testing.T) {
	_, err := NetworkEdgeTxBytes(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerTxBytes(t *testing.T) {
	_, err := NetworkContainerTxBytes("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeRxPackets(t *testing.T) {
	_, err := NetworkEdgeRxPackets(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerRxPackets(t *testing.T) {
	_, err := NetworkContainerRxPackets("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeTxPackets(t *testing.T) {
	_, err := NetworkEdgeTxPackets(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerTxPackets(t *testing.T) {
	_, err := NetworkContainerTxPackets("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgePacketRxDropped(t *testing.T) {
	_, err := NetworkEdgePacketRxDropped(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkPacketContainerRxDropped(t *testing.T) {
	_, err := NetworkContainerPacketRxDropped("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgePacketTxDropped(t *testing.T) {
	_, err := NetworkEdgePacketTxDropped(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkPacketContainerTxDropped(t *testing.T) {
	_, err := NetworkContainerPacketTxDropped("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeRxError(t *testing.T) {
	_, err := NetworkEdgeRxError(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerRxError(t *testing.T) {
	_, err := NetworkContainerRxError("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkEdgeTxError(t *testing.T) {
	_, err := NetworkEdgeTxError(time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func testNetworkContainerTxError(t *testing.T) {
	_, err := NetworkContainerTxError("foo-container-id", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
	}
}

func TestMonitorEndpoint(t *testing.T) {
	t.Run("CPUEdgeAvg", testCPUEdgeAvg)
	t.Run("CPUContainerAvg", testCPUContainerAvg)
	t.Run("CPUTimeTotal", testCPUTimeTotal)
	t.Run("CPUUtilization", testCPUUtilization)
	t.Run("MemoryEdge", testMemoryEdge)
	t.Run("MemoryContainer", testMemoryContainer)
	t.Run("MemoryEdgePeak", testMemoryEdgePeak)
	t.Run("MemoryContainerPeak", testMemoryContainerPeak)
	t.Run("MemoryContainerReachLimitSeconds", testMemoryContainerReachLimitSeconds)
	t.Run("IOEdgeTime", testIOEdgeTime)
	t.Run("IOContainerTime", testIOContainerTime)
	t.Run("IOReadEdgeBytes", testIOReadEdgeBytes)
	t.Run("IOReadContainerBytes", testIOReadContainerBytes)
	t.Run("IOWriteEdgeBytes", testIOWriteEdgeBytes)
	t.Run("IOWriteContainerBytes", testIOWriteContainerBytes)
	t.Run("IOFilesystemSizeBytes", testIOFilesystemSizeBytes)
	t.Run("IOFilesystemUsedBytes", testIOFilesystemUsedBytes)
	t.Run("NetworkEdgeRxBytes", testNetworkEdgeRxBytes)
	t.Run("NetworkContainerRxBytes", testNetworkContainerRxBytes)
	t.Run("NetworkEdgeTxBytes", testNetworkEdgeTxBytes)
	t.Run("NetworkContainerTxBytes", testNetworkContainerTxBytes)
	t.Run("NetworkEdgeRxPackets", testNetworkEdgeRxPackets)
	t.Run("NetworkContainerRxPackets", testNetworkContainerRxPackets)
	t.Run("NetworkEdgeTxPackets", testNetworkEdgeTxPackets)
	t.Run("NetworkContainerTxPackets", testNetworkContainerTxPackets)
	t.Run("NetworkEdgePacketRxDropped", testNetworkEdgePacketRxDropped)
	t.Run("NetworkPacketContainerRxDropped", testNetworkPacketContainerRxDropped)
	t.Run("NetworkEdgePacketTxDropped", testNetworkEdgePacketTxDropped)
	t.Run("NetworkPacketContainerTxDropped", testNetworkPacketContainerTxDropped)
	t.Run("NetworkEdgeRxError", testNetworkEdgeRxError)
	t.Run("NetworkContainerRxError", testNetworkContainerRxError)
	t.Run("NetworkEdgeTxError", testNetworkEdgeTxError)
	t.Run("NetworkContainerTxError", testNetworkContainerTxError)
}
