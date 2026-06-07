// package shadow implements an AWS IoT Core Device Shadow service client over MQTT.
package shadow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
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

// Client implements interaction with the AWS IoT Core Device Shadow service.
//
// T is the state that will be managed by the AWS IoT Core Device Shadow service.
// It must be marshallable into a JSON object.
//
// See the [AWS IoT Device Shadow service] documentation.
//
// [AWS IoT Device Shadow service]: https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html
type Client[T any] struct {
	mqtt          *autopaho.ConnectionManager
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
// T is the state that will be managed by the AWS IoT Core Device Shadow service.
// It must be marshallable into a JSON object.
//
// See the [AWS IoT Device Shadow service] documentation.
//
// [AWS IoT Device Shadow service]: https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html
func NewClient[T any](connectionManager *autopaho.ConnectionManager, thingName, shadowName string, updateHandler UpdateHandler[T]) *Client[T] {
	sc := &Client[T]{
		mqtt:          connectionManager,
		thingName:     thingName,
		shadowName:    shadowName,
		updateHandler: updateHandler,
	}

	connectionManager.AddOnPublishReceived(sc.onPublishReceived)

	return sc
}

func (sc *Client[T]) TopicPrefix() string {
	if sc.shadowName == "" {
		return fmt.Sprintf("$aws/things/%s/shadow", sc.thingName)
	}

	return fmt.Sprintf("$aws/things/%s/shadow/name/%s", sc.thingName, sc.shadowName)
}

// OnConnectionUp should be called in your [autopaho.ClientConfig] OnConnectionUp function
// in order to (re)subscribe to the AWS IoT Core Device Shadow service MQTT topics.
func (sc *Client[T]) OnConnectionUp(ctx context.Context) error {
	prefix := sc.TopicPrefix()
	if _, err := sc.mqtt.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: prefix + "/get/accepted", QoS: 1},
			{Topic: prefix + "/get/rejected", QoS: 1},
			{Topic: prefix + "/update/accepted", QoS: 1},
			{Topic: prefix + "/update/rejected", QoS: 1},
			{Topic: prefix + "/delete/accepted", QoS: 1},
			{Topic: prefix + "/delete/rejected", QoS: 1},
			{Topic: prefix + "/update/delta", QoS: 1},
			{Topic: prefix + "/update/documents", QoS: 1},
		},
	}); err != nil {
		return fmt.Errorf("shadow subscriptions failed: %w", err)
	}

	return nil
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

// Delete deletes the shadow.
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

	if _, err := sc.mqtt.Publish(ctx, &paho.Publish{
		Topic:   sc.topic(operation),
		Payload: payload,
		QoS:     1,
	}); err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	select {
	case res := <-ch:
		return res.resp, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("shadow %s: %w", operation, ctx.Err())
	}
}

func (sc *Client[T]) onPublishReceived(pr autopaho.PublishReceived) (bool, error) {
	if !strings.HasPrefix(pr.Packet.Topic, sc.TopicPrefix()) {
		// Not our message to process, pass it on.
		return false, nil
	}

	// Handle the message in a separate goroutine so that we don't block processing of new messages.
	go sc.handleMessage(pr)

	return true, nil
}

func (sc *Client[T]) handleMessage(pr autopaho.PublishReceived) (bool, error) {
	topic := pr.Packet.Topic
	payload := pr.Packet.Payload

	if strings.HasSuffix(topic, "/update/delta") {
		if sc.updateHandler != nil {
			var delta DeltaResponse[T]
			if err := json.Unmarshal(payload, &delta); err != nil {
				return true, fmt.Errorf("failed to unmarshal shadow state delta: %w", err)
			}

			sc.updateHandler.HandleShadowUpdateDelta(&delta)
		}

		return true, nil
	}

	if strings.HasSuffix(topic, "/update/documents") {
		if sc.updateHandler != nil {
			var documents DocumentsResponse[T]
			if err := json.Unmarshal(payload, &documents); err != nil {
				return true, fmt.Errorf("failed to unmarshal shadow state documents: %w", err)
			}

			sc.updateHandler.HandleShadowUpdateDocuments(&documents)
		}

		return true, nil
	}

	var envelope struct {
		ClientToken string `json:"clientToken"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return true, err
	}

	val, ok := sc.pending.LoadAndDelete(envelope.ClientToken)
	if !ok {
		// TODO(mtraver) Would it be useful to surface an error in this case? I don't
		// think it should be normal for this to happen. It indicates a client token
		// not being set or received properly or a message type that we don't support.
		return true, nil
	}
	ch := val.(chan result[T])

	if strings.HasSuffix(topic, "/accepted") {
		var resp AcceptedResponse[T]
		if err := json.Unmarshal(payload, &resp); err != nil {
			ch <- result[T]{err: err}
			return true, err
		}

		ch <- result[T]{resp: &resp}
	} else if strings.HasSuffix(topic, "/rejected") {
		var errResp ErrorResponse
		if err := json.Unmarshal(payload, &errResp); err != nil {
			ch <- result[T]{err: err}
			return true, err
		}

		if errResp.Code == 404 {
			ch <- result[T]{err: fmt.Errorf("%w: %s", ErrNotFound, errResp.Message)}
		} else {
			ch <- result[T]{err: &errResp}
		}
	} else {
		err := fmt.Errorf("received message on unknown topic: %q", topic)
		ch <- result[T]{err: err}
		return true, err
	}

	return true, nil
}
