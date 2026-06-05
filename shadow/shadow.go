// package shadow implements an AWS IoT Core Device Shadow service client over MQTT.
package shadow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type shadowOperation string

const (
	operationGet    shadowOperation = "get"
	operationUpdate                 = "update"
	operationDelete                 = "delete"
)

type result[T any] struct {
	resp *AcceptedResponse[T]
	err  error
}

// Client allows interaction with the AWS IoT Core Device Shadow service.
//
// See the [AWS IoT Device Shadow service] documentation.
//
// [AWS IoT Device Shadow service]: https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html
type Client[T any] struct {
	mqtt          mqtt.Client
	thingName     string
	shadowName    string
	pending       sync.Map
	updateHandler UpdateHandler[T]
}

// NewClient creates a new shadow client.
//
// If shadowName is the empty string then the client will interact with the
// unnamed (classic) shadow.
//
// See the [AWS IoT Device Shadow service] documentation.
//
// [AWS IoT Device Shadow service]: https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html
func NewClient[T any](mqttClient mqtt.Client, thingName, shadowName string, updateHandler UpdateHandler[T]) (*Client[T], error) {
	sc := &Client[T]{
		mqtt:          mqttClient,
		thingName:     thingName,
		shadowName:    shadowName,
		updateHandler: updateHandler,
	}

	prefix := sc.TopicPrefix()
	subs := map[string]byte{
		prefix + "/get/accepted":     1,
		prefix + "/get/rejected":     1,
		prefix + "/update/accepted":  1,
		prefix + "/update/rejected":  1,
		prefix + "/delete/accepted":  1,
		prefix + "/delete/rejected":  1,
		prefix + "/update/delta":     1,
		prefix + "/update/documents": 1,
	}

	token := mqttClient.SubscribeMultiple(subs, sc.handleMessage)
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("shadow subscriptions failed: %w", token.Error())
	}

	return sc, nil
}

func (sc *Client[T]) TopicPrefix() string {
	if sc.shadowName == "" {
		return fmt.Sprintf("$aws/things/%s/shadow", sc.thingName)
	}

	return fmt.Sprintf("$aws/things/%s/shadow/name/%s", sc.thingName, sc.shadowName)
}

// Get gets the shadow's current state.
func (sc *Client[T]) Get(ctx context.Context) (*AcceptedResponse[T], error) {
	return sc.request(ctx, operationGet, nil)
}

// ReportState reports the device's current state.
func (sc *Client[T]) ReportState(ctx context.Context, state T) (*AcceptedResponse[T], error) {
	return sc.update(ctx, Request[T]{
		State: DesiredReported[T]{
			Reported: state,
		},
	})
}

// UpdateDesiredState sets the shadow's desired state.
func (sc *Client[T]) UpdateDesiredState(ctx context.Context, state T) (*AcceptedResponse[T], error) {
	return sc.update(ctx, Request[T]{
		State: DesiredReported[T]{
			Desired: state,
		},
	})
}

func (sc *Client[T]) Delete(ctx context.Context) error {
	_, err := sc.request(ctx, operationDelete, nil)
	return err
}

func (sc *Client[T]) topic(operation shadowOperation) string {
	return fmt.Sprintf("%s/%s", sc.TopicPrefix(), operation)
}

func (sc *Client[T]) update(ctx context.Context, state Request[T]) (*AcceptedResponse[T], error) {
	return sc.request(ctx, operationUpdate, &state)
}

func (sc *Client[T]) request(ctx context.Context, operation shadowOperation, request *Request[T]) (*AcceptedResponse[T], error) {
	clientToken := fmt.Sprintf("%s-%d", operation, time.Now().UnixNano())

	// If the request is nil then we'll initialize an empty one
	// so that we can send a client token.
	if request == nil {
		request = &Request[T]{}
	}
	request.ClientToken = clientToken

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	ch := make(chan result[T], 1)
	sc.pending.Store(clientToken, ch)
	defer sc.pending.Delete(clientToken)

	topic := sc.topic(operation)
	if token := sc.mqtt.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("publish failed: %w", token.Error())
	}

	select {
	case res := <-ch:
		return res.resp, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("shadow %s: %w", operation, ctx.Err())
	}
}

func (sc *Client[T]) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()

	if strings.HasSuffix(topic, "/update/delta") {
		if sc.updateHandler != nil {
			var delta DeltaResponse[T]
			if err := json.Unmarshal(msg.Payload(), &delta); err != nil {
				log.Printf("Failed to unmarshal delta: %v", err)
				return
			}

			sc.updateHandler.HandleShadowUpdateDelta(&delta)
		}

		return
	}

	if strings.HasSuffix(topic, "/update/documents") {
		if sc.updateHandler != nil {
			var documents DocumentsResponse[T]
			if err := json.Unmarshal(msg.Payload(), &documents); err != nil {
				log.Printf("Failed to unmarshal documents: %v", err)
				return
			}

			sc.updateHandler.HandleShadowUpdateDocuments(&documents)
		}

		return
	}

	var envelope struct {
		ClientToken string `json:"clientToken"`
	}
	if err := json.Unmarshal(msg.Payload(), &envelope); err != nil {
		return
	}

	val, ok := sc.pending.LoadAndDelete(envelope.ClientToken)
	if !ok {
		return
	}
	ch := val.(chan result[T])

	if strings.HasSuffix(topic, "/accepted") {
		var resp AcceptedResponse[T]
		if err := json.Unmarshal(msg.Payload(), &resp); err != nil {
			ch <- result[T]{err: err}
			return
		}

		ch <- result[T]{resp: &resp}
	} else if strings.HasSuffix(topic, "/rejected") {
		var errResp ErrorResponse
		if err := json.Unmarshal(msg.Payload(), &errResp); err != nil {
			ch <- result[T]{err: err}
			return
		}

		if errResp.Code == 404 {
			ch <- result[T]{err: fmt.Errorf("%w: %s", ErrNotFound, errResp.Message)}
		} else {
			ch <- result[T]{err: &errResp}
		}
	} else {
		ch <- result[T]{err: fmt.Errorf("received message on unknown topic: %q", topic)}
	}
}
