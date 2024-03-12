# Changelog
test
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- refactoring tracing for events and APIs
- basic kinesis plugin
- addRoute api to return error if tenant does not exist
- deleteTenant api to return error if tenant has routes
- various fixes for unit tests

## [v0.3.0]
- Added basic metrics and tracing
- Added secret managements
- Fix bugs in quota manager
- Fix bugs in sqs plugins

## [v0.2.0]

A working version of EARS including:
- event handling and acknowledgements
- plugin framework
- routing table manager and filter chain
- routing configuration REST API  
- github actions, normalize analysis tools, Dockerfiles and Makefiles. [#60](https://github.com/xmidt-org/ears/pull/60)
- debug, sqs, kafka, and http sender/reciever plugins
- various filter plugins
- multi-tenant support
- tenant quota and rate limiting


## [v0.1.0]

* Initial creation

[Unreleased]: https://github.com/xmidt-org/ears/compare/v0.2.0..HEAD
[v0.2.0]: https://github.com/xmidt-org/ears/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/xmidt-org/ears/compare/aab401079d1826abe069f3f7e8516371a440bafc...v0.1.0
