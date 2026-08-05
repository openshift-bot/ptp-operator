package event

import (
	"fmt"
	exports "github.com/redhat-cne/ptp-listener-exports"
	sdktypes "github.com/redhat-cne/sdk-go/pkg/types"
	"reflect"
	"testing"
	"time"
)

func Test_compositeEventKey(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		dataType string
		want     string
	}{
		{
			name:     "both resource and dataType",
			resource: "/cluster/node/master0/CLOCK_REALTIME",
			dataType: "notification",
			want:     "/cluster/node/master0/CLOCK_REALTIME/notification",
		},
		{
			name:     "empty resource falls back to dataType",
			resource: "",
			dataType: "metric",
			want:     "metric",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compositeEventKey(tt.resource, tt.dataType)
			if got != tt.want {
				t.Errorf("compositeEventKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValueByDataType(t *testing.T) {
	values := exports.StoredEventValues{
		"/cluster/node/master0/CLOCK_REALTIME/notification": "LOCKED",
		"/cluster/node/master0/CLOCK_REALTIME/metric":       float64(-4),
	}

	t.Run("finds notification by suffix", func(t *testing.T) {
		val, ok := ValueByDataType(values, "notification")
		if !ok {
			t.Fatal("expected to find notification")
		}
		if val != "LOCKED" {
			t.Errorf("got %v, want LOCKED", val)
		}
	})

	t.Run("finds metric by suffix", func(t *testing.T) {
		val, ok := ValueByDataType(values, "metric")
		if !ok {
			t.Fatal("expected to find metric")
		}
		if val != float64(-4) {
			t.Errorf("got %v, want -4", val)
		}
	})

	t.Run("returns false for missing dataType", func(t *testing.T) {
		_, ok := ValueByDataType(values, "nonexistent")
		if ok {
			t.Error("expected not found")
		}
	})

	t.Run("works with legacy empty-resource keys", func(t *testing.T) {
		legacy := exports.StoredEventValues{
			"notification": "FREERUN",
		}
		val, ok := ValueByDataType(legacy, "notification")
		if !ok {
			t.Fatal("expected to find notification")
		}
		if val != "FREERUN" {
			t.Errorf("got %v, want FREERUN", val)
		}
	})
}

func Test_createStoredEvent(t *testing.T) {
	type args struct {
		data []byte
	}
	aTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", "2023-05-10T20:25:54.453849118Z")
	ts := sdktypes.Timestamp{Time: aTime}
	fmt.Print(err)
	tests := []struct {
		name             string
		args             args
		wantAStoredEvent exports.StoredEvent
		wantAType        string
		wantErr          bool
	}{
		{args: args{data: []byte("{\\\"specversion\\\":\\\"0.3\\\",\\\"id\\\":\\\"f078751e-2712-4176-a66b-e8f6fc8f13ff\\\",\\\"source\\\":\\\"/cluster/node/master0/sync/sync-status/os-clock-sync-state\\\",\\\"type\\\":\\\"event.sync.sync-status.os-clock-sync-state-change\\\",\\\"subject\\\":\\\"/cluster/node/master0/sync/sync-status/os-clock-sync-state\\\",\\\"datacontenttype\\\":\\\"application/json\\\",\\\"time\\\":\\\"2023-05-10T20:25:54.453849118Z\\\",\\\"data\\\":{\\\"version\\\":\\\"v1\\\",\\\"values\\\":[{\\\"resource\\\":\\\"/cluster/node/master0/CLOCK_REALTIME\\\",\\\"dataType\\\":\\\"notification\\\",\\\"valueType\\\":\\\"enumeration\\\",\\\"value\\\":\\\"LOCKED\\\"},{\\\"resource\\\":\\\"/cluster/node/master0/CLOCK_REALTIME\\\",\\\"dataType\\\":\\\"metric\\\",\\\"valueType\\\":\\\"decimal64.3\\\",\\\"value\\\":\\\"-4\\\"}]}}")},
			name:             "ok",
			wantAStoredEvent: exports.StoredEvent{exports.EventTimeStamp: interface{}(&ts), exports.EventType: interface{}("event.sync.sync-status.os-clock-sync-state-change"), exports.EventSource: "/cluster/node/master0/sync/sync-status/os-clock-sync-state", exports.EventValues: exports.StoredEventValues{"/cluster/node/master0/CLOCK_REALTIME/notification": "LOCKED", "/cluster/node/master0/CLOCK_REALTIME/metric": float64(-4)}},
			wantAType:        "event.sync.sync-status.os-clock-sync-state-change",
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAStoredEvent, gotAType, err := createStoredEvent(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStoredEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAStoredEvent, tt.wantAStoredEvent) {
				t.Errorf("createStoredEvent() gotAStoredEvent = %v, want %v", gotAStoredEvent, tt.wantAStoredEvent)
			}
			if gotAType != tt.wantAType {
				t.Errorf("createStoredEvent() gotAType = %v, want %v", gotAType, tt.wantAType)
			}
		})
	}
}
