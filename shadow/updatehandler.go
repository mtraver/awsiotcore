package shadow

// An UpdateHandler handles messages published to the [/update/delta] and [/update/documents] topics.
//
// [/update/delta]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-mqtt.html#update-delta-pub-sub-topic
// [/update/documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-mqtt.html#update-documents-pub-sub-topic
type UpdateHandler[T any] interface {
	HandleShadowUpdateDelta(delta *DeltaResponse[T])
	HandleShadowUpdateDocuments(documents *DocumentsResponse[T])
}
