package metric

import "time"

//The metric struct.
//Contains the response of a Prometheus API request.
/*
These are 4 types of Prometheus metrics
Matrix refers to metrics that has a range in time. Such as from now to the past 10 minutes.
Vector refers to metrics that happened in a particular point in time.
These two types can contain more than one device. For CPUs, it could be individual cores. Think of it like different lines in a Grafana graph.

Scalars are single element values. You can get a scalar result by adding scalar(vector) to the query.
Note: If the vector in the parameter has more than 1 scalar, the value becomes NaN.

String is unused but defined in the Prometheus API
*/

type MetricType string

const (
	MatrixType MetricType = "matrix"
	VectorType MetricType = "vector"
	ScalarType MetricType = "scalar"
	//Unused type in Prometheus
	StringType MetricType = "string"
)

type PromMetric struct {
	Type MetricType `json:"type"`
	//The Data field can either be one of these three types
	//[]Matrix
	//[]Vector
	//Scalar
	Data interface{} `json:"data"`
}

//Matrix must be used in arrays due to how metrics are structured
type Matrix struct {
	//The key usually defers to the device name, job name, instance name, etc.
	Key map[string]string `json:"key"`
	//The metric data
	Values []Scalar `json:"values"`
}

//Vector must be used in arrays due to how metrics are structured
type Vector struct {
	//The key usually defers to the device name, job name, instance name, etc.
	Key map[string]string `json:"key"`
	//The metric data
	Scalar Scalar `json:"scalar"`
}

type Scalar struct {
	Time time.Time `json:"time"`
	//In Prometheus terms, any numeric value are referred as scalars
	Value float64 `json:"value"`
	//JavaScript shenanigans. Normally the scalar should be a float string, but Prometheus will return "NaN" in certain conditions, or when there's no data.
	Undefined bool `json:"undefined"`
}
