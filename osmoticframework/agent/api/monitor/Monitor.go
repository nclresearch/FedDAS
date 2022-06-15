package monitor

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"strconv"
	"time"
)

//The Prometheus container will only deploy on itself. So it is safe to hardcode the address to localhost
//We do not need to open port 9090 on the edge device.
//However, we do need to open ports for the monitoring services (cAdvisor, Node exporter)
const promAddress = "http://localhost:9090/api/v1"

//Queries in a specific point in time.
func restQuery(query string, time time.Time) (*Metric, error) {
	client := resty.New()
	response, err := client.R().
		SetQueryParams(map[string]string{
			"query": query,
			"time":  strconv.FormatInt(time.Unix(), 10),
		}).
		Get(promAddress + "/query")
	if err != nil {
		return nil, err
	}
	err = processStatus(response)
	if err != nil {
		return nil, err
	}
	metric, err := parsePromResponse(response.Body())
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Queries in a specific range in time, and it will return multiple points in time
//Prometheus smallest scale in time is second. Therefore there is no point in using nanosecond precision.
func restQueryRange(query string, from, to time.Time, step time.Duration) (*Metric, error) {
	client := resty.New()
	response, err := client.R().
		SetQueryParams(map[string]string{
			"query": query,
			"from":  strconv.FormatInt(from.UnixNano(), 10),
			"to":    strconv.FormatInt(to.UnixNano(), 10),
			"step":  strconv.Itoa(int(step.Seconds())),
		}).
		Get(promAddress + "/query_range")
	if err != nil {
		return nil, err
	}
	err = processStatus(response)
	if err != nil {
		return nil, err
	}
	metric, err := parsePromResponse(response.Body())
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func processStatus(response *resty.Response) error {
	if response.StatusCode() == 200 {
		return nil
	}
	var errorJson map[string]interface{}
	err := json.Unmarshal(response.Body(), &errorJson)
	if err != nil {
		return err
	}
	//API errors
	errString, ok := errorJson["error"].(string)
	if !ok {
		return errors.New("unknown error")
	}
	return errors.New(errString)
}

func parsePromResponse(response []byte) (*Metric, error) {
	var responseJson map[string]interface{}
	err := json.Unmarshal(response, &responseJson)
	if err != nil {
		return nil, errors.New("unable to parse broken json from prometheus")
	}
	var metric Metric
	resultType, ok := responseJson["data"].(map[string]interface{})["resultType"].(string)
	if !ok {
		return nil, errors.New("cannot parse api response")
	}
	resultRaw := responseJson["data"].(map[string]interface{})["result"]
	metric = Metric{}
	switch resultType {
	case "matrix":
		var result []map[string]interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		metric.Type = MatrixType
		var matrix = make([]Matrix, 0)
		for _, dev := range result {
			devInfoDecoded, ok := dev["metric"].(map[string]interface{})
			//mapstructure has decoded some of the fields in metric to int64 or float64 instead of string.
			//To minimize the unnecessary type assertions, we'll convert all of them to string
			devInfo := make(map[string]string)
			for key, value := range devInfoDecoded {
				switch value.(type) {
				case string:
					devInfo[key] = value.(string)
				case int64:
					devInfo[key] = strconv.FormatInt(value.(int64), 10)
				case float64:
					devInfo[key] = strconv.FormatFloat(value.(float64), 'E', -1, 64)
				}
			}
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			//Assertion must be interface due to mixed types (int and string)
			//JSON structure of matrix
			/*
				"result": [
					{
						"metric": {
							"cpu": "0"
						},
						"values": [
							[
								1574172422,
								"0.028000000000001805"
							],
							[
								1574172423,
								"0.028000000000001805"
							]
						]
					}
				]
			*/
			var scalars [][]interface{}
			err := mapstructure.Decode(dev["values"], &scalars)
			if err != nil {
				return nil, err
			}
			entry := Matrix{Key: devInfo}
			values := make([]Scalar, 0)
			for _, scalar := range scalars {
				value, err := parseScalar(scalar)
				if err != nil {
					return nil, err
				}
				values = append(values, *value)
			}
			entry.Values = values
			matrix = append(matrix, entry)
		}
		metric.Data = matrix
	case "vector":
		var result []map[string]interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		metric.Type = VectorType
		var vectors = make([]Vector, 0)
		for _, dev := range result {
			devInfoDecoded, ok := dev["metric"].(map[string]interface{})
			//mapstructure has decoded some of the fields in metric to int64 or float64 instead of string.
			//To minimize the unnecessary type assertions, we'll convert all of them to string
			devInfo := make(map[string]string)
			for key, value := range devInfoDecoded {
				switch value.(type) {
				case string:
					devInfo[key] = value.(string)
				case int64:
					devInfo[key] = strconv.FormatInt(value.(int64), 10)
				case float64:
					devInfo[key] = strconv.FormatFloat(value.(float64), 'E', -1, 64)
				}
			}
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			//Assertion must be interface due to mixed types (int and string)
			//JSON structure of vectors
			/*
				"result": [
					{
						"metric": {
							"cpu": "4"
						},
						"value": [
							1574172422,
							"0.027999999999999935"
						]
					}
				]
			*/
			//There's only one element in "value" for vectors. So no 2-dimensional arrays
			scalar, ok := dev["value"].([]interface{})
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			value, err := parseScalar(scalar)
			if err != nil {
				return nil, err
			}
			entry := Vector{
				Key:    devInfo,
				Scalar: *value,
			}
			vectors = append(vectors, entry)
		}
		metric.Data = vectors
	case "scalar":
		//Assertion must be interface due to mixed types (int and string)
		//JSON structure of scalars
		/*
			"result": [
				1574172422,
				"0.027999999999999935"
			]
		*/
		var result []interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		metric.Type = ScalarType
		value, err := parseScalar(result)
		if err != nil {
			return nil, err
		}
		metric.Data = *value
	case "string":
		//Unused in Prometheus. So do nothing.
		metric.Type = StringType
	default:
		return nil, errors.New("unknown data type")
	}
	return &metric, nil
}

func parseScalar(scalar []interface{}) (*Scalar, error) {
	var timestamp time.Time
	//Dynamic typing shenanigans from json
	//Sometimes Prometheus returns milliseconds in its timestamp, which makes the timestamp a float
	//Sometimes it does not, which makes it an int
	switch scalar[0].(type) {
	case int64:
		unixTimestamp, _ := scalar[0].(int64)
		timestamp = time.Unix(unixTimestamp, 0)
	case float64:
		unixTimestamp, _ := scalar[0].(float64)
		sec := int64(unixTimestamp)
		//nano seconds
		nsec := int64(unixTimestamp - (float64(sec))*1000000)
		timestamp = time.Unix(sec, nsec)
	default:
		return nil, errors.New("cannot parse time")
	}
	rawValue := scalar[1].(string)
	if rawValue == "NaN" {
		return &Scalar{
			Time:      timestamp,
			Value:     0,
			Undefined: true,
		}, nil
	} else {
		floatValue, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, errors.New("unable to parse scalar: " + rawValue)
		}
		return &Scalar{
			Time:      timestamp,
			Value:     floatValue,
			Undefined: false,
		}, nil
	}
}
