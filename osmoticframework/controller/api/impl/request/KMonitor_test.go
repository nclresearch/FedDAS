package request

import (
	"net/http"
	"osmoticframework/controller/types/metric"
	"osmoticframework/controller/vars"
	"testing"
	"time"
)

//Prometheus test address
const promAddress = "http://localhost:30000/api/v1"

//Prometheus test address - For Github actions
//Because apparently it can't change its default port, even if you overwrite its systemd script.
//It just ignores the arguments. Thanks GitHub. You wasted me a lot of time.
const promAddressGithub = "http://localhost:9090/api/v1"

func setup(t *testing.T) {
	//Change the config path as needed
	var config = `
{
  "rabbitAddress": "",
  "databaseAddress": "",
  "prometheusAddress": "` + promAddress + `",
  "kuberConfigPath": "",
  "networks": []
}
`
	vars.LoadConfig([]byte(config))
	_, err := http.Get(vars.GetPrometheusAddress())
	if err != nil {
		t.Log(err)
		t.Log("Prometheus not reachable. Retrying with default Prometheus port")
		config = `
{
  "rabbitAddress": "",
  "databaseAddress": "",
  "prometheusAddress": "` + promAddressGithub + `",
  "kuberConfigPath": "",
  "networks": []
}
`
		vars.LoadConfig([]byte(config))
		_, err := http.Get(vars.GetPrometheusAddress())
		if err != nil {
			t.Log(err)
			t.Error("Prometheus not reachable. It must be hosted on either localhost:30000 or localhost:9090")
		}
		t.Log("Looks like this is Github jank. Continue")
	}
}

func TestMatrix(t *testing.T) {
	setup(t)
	promMetric, err := restQuery("prometheus_http_requests_total[5m]", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := promMetric.Type
	if mType != metric.MatrixType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, metric.VectorType)
	}
	_, ok := promMetric.Data.([]metric.Matrix)
	if !ok {
		t.Errorf("Cannot assert to matrix")
	}
}

func TestVector(t *testing.T) {
	setup(t)
	promMetric, err := restQuery("prometheus_http_requests_total", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := promMetric.Type
	if mType != metric.VectorType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, metric.VectorType)
	}
	_, ok := promMetric.Data.([]metric.Vector)
	if !ok {
		t.Errorf("Cannot assert to vector")
	}
}

func TestScalar(t *testing.T) {
	setup(t)
	promMetric, err := restQuery("scalar(prometheus_build_info)", time.Now())
	if err != nil {
		t.Error("Request failed")
		t.Error(err)
		t.FailNow()
	}
	mType := promMetric.Type
	if mType != metric.ScalarType {
		t.Errorf("Incorrect return type. Got %s, Want %s", mType, metric.VectorType)
	}
	_, ok := promMetric.Data.(metric.Scalar)
	if !ok {
		t.Errorf("Cannot assert to scalar")
	}
}

func TestParseScalar(t *testing.T) {
	now := time.Now().UnixNano()
	tests := []struct {
		Input    []interface{}
		Expected metric.Scalar
	}{
		{
			Input: []interface{}{
				now,
				"123456789",
			},
			Expected: metric.Scalar{
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
			Expected: metric.Scalar{
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
			Expected: metric.Scalar{
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
			t.Errorf("CPUTime not correct. Got %d, Want %d", scalar.Time.Unix(), test.Expected.Time.Unix())
		}
		if scalar.Value != test.Expected.Value {
			t.Errorf("Value not correct. Got %f, Want %f", scalar.Value, test.Expected.Value)
		}
		if scalar.Undefined != test.Expected.Undefined {
			t.Errorf("Undefined not correct. Got %t, Want %t", scalar.Undefined, test.Expected.Undefined)
		}
	}
}

func testKCPUCoreAvg(t *testing.T) {
	_, err := KCPUCoreAvg("foo-node", time.Now())
	if err != nil {
		t.Error("KCPUCoreAvg failed")
		t.Error(err)
	}
}

func testKCPUPodAvg(t *testing.T) {
	_, err := KCPUPodAvg("foo-pod", time.Now())
	if err != nil {
		t.Error("KCPUPodAvg failed")
		t.Error(err)
	}
}

func testKCPUTime(t *testing.T) {
	_, err := KCPUTime("foo-node", time.Now())
	if err != nil {
		t.Error("KCPUTime failed")
		t.Error(err)
	}
}

func testKCPUUtilization(t *testing.T) {
	_, err := KCPUUtilization("foo-node", time.Now())
	if err != nil {
		t.Error("KCPUUtilization failed")
		t.Error(err)
	}
}

func testKEndpointInfo(t *testing.T) {
	_, err := KEndpointInfo()
	if err != nil {
		t.Error("KEndpointInfo failed")
		t.Error(err)
	}
}

func testKIOFilesystemSizeBytes(t *testing.T) {
	_, err := KIOFilesystemSizeBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KIOFilesystemSizeBytes failed")
		t.Error(err)
	}
}

func testKIOFilesystemUsedBytes(t *testing.T) {
	_, err := KIOFilesystemUsedBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KIOFilesystemUsedBytes failed")
		t.Error(err)
	}
}

func testKIONodeTime(t *testing.T) {
	_, err := KIONodeTime("foo-node", time.Now())
	if err != nil {
		t.Error("KIONodeTime failed")
		t.Error(err)
	}
}

func testKIOPodTime(t *testing.T) {
	_, err := KIOPodTime("foo-pod", time.Now())
	if err != nil {
		t.Error("KIOPodTime failed")
		t.Error(err)
	}
}

func testKIOReadNodeBytes(t *testing.T) {
	_, err := KIOReadNodeBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KIOReadNodeBytes failed")
		t.Error(err)
	}
}

func testKIOReadPodBytes(t *testing.T) {
	_, err := KIOReadPodBytes("foo-pod", time.Now())
	if err != nil {
		t.Error("KIOReadPodBytes failed")
		t.Error(err)
	}
}

func testKIOWriteNodeBytes(t *testing.T) {
	_, err := KIOWriteNodeBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KIOWriteNodeBytes failed")
		t.Error(err)
	}
}

func testKIOWritePodBytes(t *testing.T) {
	_, err := KIOWritePodBytes("foo-pod", time.Now())
	if err != nil {
		t.Error("KIOWritePodBytes failed")
		t.Error(err)
	}
}

func testKMemoryNode(t *testing.T) {
	_, err := KMemoryNode("foo-node", time.Now())
	if err != nil {
		t.Error("KMemoryNode failed")
		t.Error(err)
	}

}

func testKMemoryNodePeak(t *testing.T) {
	_, err := KMemoryNodePeak("foo-node", time.Now())
	if err != nil {
		t.Error("KMemoryNodePeak failed")
		t.Error(err)
	}
}

func testKMemoryPod(t *testing.T) {
	_, err := KMemoryPod("foo-pod", time.Now())
	if err != nil {
		t.Error("KMemoryPod failed")
		t.Error(err)
	}
}

func testKMemoryPodPeak(t *testing.T) {
	_, err := KMemoryPodPeak("foo-pod", time.Now())
	if err != nil {
		t.Error("KMemoryPodPeak failed")
		t.Error(err)
	}
}

func testKMemoryPodReachLimitSeconds(t *testing.T) {
	_, err := KMemoryPodReachLimitSeconds("foo-pod", time.Now())
	if err != nil {
		t.Error("KMemoryPodReachLimitSeconds failed")
		t.Error(err)
	}
}

func testKNetworkNodeRxBytes(t *testing.T) {
	_, err := KNetworkNodeRxBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeRxBytes failed")
		t.Error(err)
	}
}

func testKNetworkNodeRxDropped(t *testing.T) {
	_, err := KNetworkNodeRxDropped("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeRxDropped failed")
		t.Error(err)
	}
}

func testKNetworkNodeRxError(t *testing.T) {
	_, err := KNetworkNodeRxError("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeRxError failed")
		t.Error(err)
	}
}

func testKNetworkNodeRxPackets(t *testing.T) {
	_, err := KNetworkNodeRxPackets("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeRxPackets failed")
		t.Error(err)
	}
}

func testKNetworkNodeTxBytes(t *testing.T) {
	_, err := KNetworkNodeTxBytes("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeTxBytes failed")
		t.Error(err)
	}
}

func testKNetworkNodeTxDropped(t *testing.T) {
	_, err := KNetworkNodeTxDropped("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeTxDropped failed")
		t.Error(err)
	}
}

func testKNetworkNodeTxError(t *testing.T) {
	_, err := KNetworkNodeTxError("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeTxError failed")
		t.Error(err)
	}
}

func testKNetworkNodeTxPackets(t *testing.T) {
	_, err := KNetworkNodeTxPackets("foo-node", time.Now())
	if err != nil {
		t.Error("KNetworkNodeTxPackets failed")
		t.Error(err)
	}
}

func testKNetworkPodRxBytes(t *testing.T) {
	_, err := KNetworkPodRxBytes("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodRxBytes failed")
		t.Error(err)
	}
}

func testKNetworkPodRxDropped(t *testing.T) {
	_, err := KNetworkPodRxDropped("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodRxDropped failed")
		t.Error(err)
	}
}

func testKNetworkPodRxError(t *testing.T) {
	_, err := KNetworkPodRxError("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodRxError failed")
		t.Error(err)
	}
}

func testKNetworkPodRxPackets(t *testing.T) {
	_, err := KNetworkPodRxPackets("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodRxPackets failed")
		t.Error(err)
	}
}

func testKNetworkPodTxBytes(t *testing.T) {
	_, err := KNetworkPodTxBytes("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodTxBytes failed")
		t.Error(err)
	}
}

func testKNetworkPodTxDropped(t *testing.T) {
	_, err := KNetworkPodTxDropped("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodTxDropped failed")
		t.Error(err)
	}
}

func testKNetworkPodTxError(t *testing.T) {
	_, err := KNetworkPodTxError("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodTxError failed")
		t.Error(err)
	}
}

func testKNetworkPodTxPackets(t *testing.T) {
	_, err := KNetworkPodTxPackets("foo-pod", time.Now())
	if err != nil {
		t.Error("KNetworkPodTxPackets failed")
		t.Error(err)
	}
}

func TestKMonitorEndpoints(t *testing.T) {
	setup(t)
	t.Run("KCPUCoreAvg", testKCPUCoreAvg)
	t.Run("KCPUPodAvg", testKCPUPodAvg)
	t.Run("KCPUTime", testKCPUTime)
	t.Run("KCPUUtilization", testKCPUUtilization)
	t.Run("KEndpointInfo", testKEndpointInfo)
	t.Run("KIOFilesystemSizeBytes", testKIOFilesystemSizeBytes)
	t.Run("KIOFilesystemUsedBytes", testKIOFilesystemUsedBytes)
	t.Run("KIONodeTime", testKIONodeTime)
	t.Run("KIOPodTime", testKIOPodTime)
	t.Run("KIOReadNodeBytes", testKIOReadNodeBytes)
	t.Run("KIOReadPodBytes", testKIOReadPodBytes)
	t.Run("KIOWriteNodeBytes", testKIOWriteNodeBytes)
	t.Run("KIOWritePodBytes", testKIOWritePodBytes)
	t.Run("KMemoryNode", testKMemoryNode)
	t.Run("KMemoryPod", testKMemoryPod)
	t.Run("KMemoryNodePeak", testKMemoryNodePeak)
	t.Run("KMemoryPodPeak", testKMemoryPodPeak)
	t.Run("KMemoryPodReachLimitSeconds", testKMemoryPodReachLimitSeconds)
	t.Run("KNetworkNodeRxBytes", testKNetworkNodeRxBytes)
	t.Run("KNetworkNodeRxDropped", testKNetworkNodeRxDropped)
	t.Run("KNetworkNodeRxError", testKNetworkNodeRxError)
	t.Run("KNetworkNodeRxPackets", testKNetworkNodeRxPackets)
	t.Run("KNetworkNodeTxBytes", testKNetworkNodeTxBytes)
	t.Run("KNetworkNodeTxDropped", testKNetworkNodeTxDropped)
	t.Run("KNetworkNodeTxError", testKNetworkNodeTxError)
	t.Run("KNetworkNodeTxPackets", testKNetworkNodeTxPackets)
	t.Run("KNetworkPodRxBytes", testKNetworkPodRxBytes)
	t.Run("KNetworkPodRxDropped", testKNetworkPodRxDropped)
	t.Run("KNetworkPodRxError", testKNetworkPodRxError)
	t.Run("KNetworkPodRxPackets", testKNetworkPodRxPackets)
	t.Run("KNetworkPodTxBytes", testKNetworkPodTxBytes)
	t.Run("KNetworkPodTxDropped", testKNetworkPodTxDropped)
	t.Run("KNetworkPodTxError", testKNetworkPodTxError)
	t.Run("KNetworkPodTxPackets", testKNetworkPodTxPackets)
}
