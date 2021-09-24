# Event Async Routing Service (EARS)

[![Build Status](https://github.com/xmidt-org/ears/workflows/CI/badge.svg)](https://github.com/xmidt-org/ears/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/ears/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/ears?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/ears)](https://goreportcard.com/report/github.com/xmidt-org/ears)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/ears/blob/main/LICENSE)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_ears&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_ears)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/ears.svg)](CHANGELOG.md)
[![GoDoc](https://godoc.org/github.com/xmidt-org/ears?status.svg)](https://godoc.org/github.com/xmidt-org/ears)

## Summary

A simple scalable routing service to usher events from an input plugin (for example, Kafka) to an output plugin (for example, AWS SQS).
As an event passes through EARS, it may be filtered or transformed depending on the configuration details of a route and the 
event payload. Routes can be dynamically added and removed using a simple REST API and modifications to the routing table are 
quickly synchronized across an EARS cluster.

EARS is designed to eventually replace EEL, offering new features such as quotas and rate limiting as well as highly
dynamic routes while still supporting filtering and transformation capabilities similar to EEL. 

EARS comes with a set of standard plugins to support some of the most common message protocols including Webhook, Kafka, SQS, 
Kinesis etc. but also makes the development of third party plugins easy.  

Our Kanban Board can be found [here](https://github.com/orgs/xmidt-org/projects/3).

## User Guide

* [Routes](userguide/routes.md)
* [EARS API](userguide/api.md)
* [EARS vs. EEL](userguide/eel.md)
* [Simple Examples](userguide/examples.md)
* [Debug Strategies](userguide/debug.md)
* [Filter Plugin Reference](userguide/filters.md)
* [Receiver Plugins Reference](userguide/receivers.md)
* [Sender Plugins Reference](userguide/senders.md)
* [Plugin Developer Guide](userguide/plugindev.md)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
