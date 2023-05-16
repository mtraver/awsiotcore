# awsiotcore
AWS IoT Core over MQTT in Go

This package follows https://aws.amazon.com/blogs/iot/use-aws-iot-core-mqtt-broker-with-standard-mqtt-libraries/.

# Requirements

## Amazon CA certs

You'll need Amazon's CA certs listed under "CA certificates for server authentication" [https://docs.aws.amazon.com/iot/latest/developerguide/server-authentication.html](here).

As of 2023-04-22 they are:

> - RSA 2048 bit key: [Amazon Root CA 1](https://www.amazontrust.com/repository/AmazonRootCA1.pem)
> - RSA 4096 bit key: Amazon Root CA 2. Reserved for future use.
> - ECC 256 bit key: [Amazon Root CA 3](https://www.amazontrust.com/repository/AmazonRootCA3.pem)
> - ECC 384 bit key: Amazon Root CA 4. Reserved for future use.

Download these and put them in a .pem file and use it when calling `NewClient`.

## Endpoint URL

You'll need the endpoint of the MQTT broker to connect to. You can find that in the "Device data endpoint" section of the AWS IoT settings page, or you can fetch it using the AWS CLI:

```sh
aws iot describe-endpoint --endpoint-type iot:Data-ATS --query 'endpointAddress' --output text
```

For sending and receiving data from the message broker, use an `iot:Data-ATS` endpoint. See https://docs.aws.amazon.com/iot/latest/developerguide/iot-connect-devices.html#iot-connect-device-endpoints for the various endpoint types.

# MQTT topics

By default telemetry will be sent to `things/{device_id}/telemetry`. Set `TelemetryTopicOverride`
on the `Device` to change that.
