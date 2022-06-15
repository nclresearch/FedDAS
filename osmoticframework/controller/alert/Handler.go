package alert

import (
	"github.com/mitchellh/mapstructure"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
)

func ParseAlert(jsonMsg map[string]interface{}) {
	if jsonMsg["type"] == nil {
		return
	}
	switch jsonMsg["type"] {
	case ContainerCrash:
		var crashReport types.ContainerCrashReport
		err := mapstructure.Decode(jsonMsg["contents"], &crashReport)
		if err != nil {
			log.Error.Println("Cannot decode container crash report")
			log.Error.Println(err)
			//Ignore if cannot decode
			return
		}
		ContainerCrash <- crashReport
	default:
		return
	}
}

func DBErrorHandler(dbErr types.DatabaseErrorReport) {
	log.Error.Println("From query: " + dbErr.Query)
	log.Error.Printf("Query arguments: #%v\n", dbErr.QueryArgs)
	log.Error.Printf("SQL Error response: %s\n", dbErr.Error)
}
