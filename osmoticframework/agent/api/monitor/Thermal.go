package monitor

import "time"

func Thermals(time time.Time) (*Metric, error) {
	const query = "node_thermal_zone_temp"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
