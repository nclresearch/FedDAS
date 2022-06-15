package vars

import (
	"encoding/base64"
	"encoding/json"
	"osmoticframework/controller/types"
	"reflect"
	"testing"
)

func TestGenerateTargetJson(t *testing.T) {
	type args struct {
		agentIds map[string]types.Agent
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test generate",
			args: args{agentIds: map[string]types.Agent{
				"AgentA": {
					InternalIP:    "192.168.1.1",
					DeviceSupport: []string{"gpu"},
					SensorSupport: []string{},
					Containers:    []string{"a5b8965f5a9624471b2ac9e7cb468a8e1b9c561acf20ca56a0ac3377cb6c2159"},
					LastAlive:     1602253967,
				},
				"AgentB": {
					InternalIP:    "192.168.1.2",
					DeviceSupport: []string{},
					SensorSupport: []string{"camera"},
					Containers:    []string{},
					LastAlive:     1602253977,
				},
			}},
			want: "W3sibGFiZWxzIjp7ImpvYiI6IkFnZW50QSJ9LCJ0YXJnZXRzIjpbIjE5Mi4xNjguMS4xOjgwODAiLCIxOTIuMTY4LjEuMTo5MTAwIl19LHsibGFiZWxzIjp7ImpvYiI6IkFnZW50QiJ9LCJ0YXJnZXRzIjpbIjE5Mi4xNjguMS4yOjgwODAiLCIxOTIuMTY4LjEuMjo5MTAwIl19XQ==",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//The output json may not be in an exact order, hence the test would compare with the deserialised json instead
			got := GenerateTargetJson(tt.args.agentIds)
			var targets []TargetJob
			decode, err := base64.StdEncoding.DecodeString(got)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			err = json.Unmarshal(decode, &targets)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if len(targets) != 2 {
				t.Errorf("Target size incorrect. Got %d, Want %d", len(targets), 2)
			}
			if !(targets[0].Labels.Job == "AgentA" || targets[1].Labels.Job == "AgentA") {
				t.Errorf("AgentA does not exist in taeget file")
			}
			if !(targets[0].Labels.Job == "AgentB" || targets[1].Labels.Job == "AgentB") {
				t.Errorf("AgentB does not exist in taeget file")
			}
			agentATargets := []string{"192.168.1.1:8080", "192.168.1.1:9100"}
			agentBTargets := []string{"192.168.1.2:8080", "192.168.1.2:9100"}
			if targets[0].Labels.Job == "AgentA" {
				if !reflect.DeepEqual(targets[0].Target, agentATargets) {
					t.Errorf("A Scarpe target address not correct. Got %s, Want %s", targets[0].Target, agentATargets)
				}
				if !reflect.DeepEqual(targets[1].Target, agentBTargets) {
					t.Errorf("B Scarpe target address not correct. Got %s, Want %s", targets[1].Target, agentBTargets)
				}
			} else {
				if !reflect.DeepEqual(targets[1].Target, agentATargets) {
					t.Errorf("A Scarpe target address not correct. Got %s, Want %s", targets[1].Target, agentATargets)
				}
				if !reflect.DeepEqual(targets[0].Target, agentBTargets) {
					t.Errorf("B Scarpe target address not correct. Got %s, Want %s", targets[0].Target, agentBTargets)
				}
			}
		})
	}
}
