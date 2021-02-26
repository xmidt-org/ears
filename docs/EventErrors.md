## Context refactoring

https://github.com/xmidt-org/ears/issues/51

TLDR: We don't need application context

Code in question:
* [RoutingTableManager](https://github.com/xmidt-org/ears/pull/87/files#diff-17ef71a9dd52d72538310afabec1694e60c82e50f5a0ecfc3f5b028156f6cf3e)
* [Receiver Interface](https://github.com/xmidt-org/ears/pull/87/files#diff-4d6bbf527e27eab751626d1b5f04edec2926b061ea936550756744cb24023cde)
* Test cases can no longer stop receivers with context. Instead, need to use `receiver.StopReceiving(ctx)`. [An example](https://github.com/xmidt-org/ears/pull/87/files#diff-1a8c34f76e657abdaa043a78d4057decc9caed9360284c894217eb118d1d32c8)

### Problem 1

[nil exception](https://github.com/xmidt-org/ears/blob/issues51/pkg/route/route_test.go#L111)

```
--- FAIL: TestErrorCases (0.00s)
panic: ReceiverMock.StopReceivingFunc: method is nil but Receiver.StopReceiving was just called [recovered]
	panic: ReceiverMock.StopReceivingFunc: method is nil but Receiver.StopReceiving was just called

goroutine 9 [running]:
testing.tRunner.func1.1(0x1378fa0, 0x1426cc0)
	/usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1072 +0x46a
testing.tRunner.func1(0xc0001a2300)
	/usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1075 +0x636
panic(0x1378fa0, 0x1426cc0)
	/usr/local/Cellar/go/1.15.6/libexec/src/runtime/panic.go:975 +0x47a
github.com/xmidt-org/ears/pkg/receiver.(*ReceiverMock).StopReceiving(0xc000190460, 0x142efa0, 0xc000018078, 0xc000053390, 0x1512401)
	/Users/mchian000/src/ears/pkg/receiver/testing_mock.go:264 +0x275
github.com/xmidt-org/ears/pkg/route_test.TestErrorCases(0xc0001a2300)
	/Users/mchian000/src/ears/pkg/route/route_test.go:69 +0x543
testing.tRunner(0xc0001a2300, 0x13ea490)
	/usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1123 +0x203
created by testing.(*T).Run
	/usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1168 +0x5bc
```

### Problem 2

[Race condition](https://github.com/xmidt-org/ears/blob/main/internal/pkg/plugin/manager.go#L211)

```
WARNING: DATA RACE
Write at 0x00c000422c30 by goroutine 10:
  runtime.mapdelete_faststr()
      /usr/local/Cellar/go/1.15.6/libexec/src/runtime/map_faststr.go:297 +0x0
  github.com/xmidt-org/ears/internal/pkg/plugin.(*manager).stopReceiving()
      /Users/mchian000/src/ears/internal/pkg/plugin/manager.go:256 +0x1d2
  github.com/xmidt-org/ears/internal/pkg/plugin.(*receiver).StopReceiving()
      /Users/mchian000/src/ears/internal/pkg/plugin/receiver.go:81 +0xaf
  github.com/xmidt-org/ears/pkg/route.(*Route).Stop()
      /Users/mchian000/src/ears/pkg/route/route.go:79 +0x106
  github.com/xmidt-org/ears/internal/pkg/app.(*DefaultRoutingTableManager).RemoveRoute()
      /Users/mchian000/src/ears/internal/pkg/app/routingTableManager.go:60 +0x244
  github.com/xmidt-org/ears/internal/pkg/app.(*APIManager).removeRouteHandler()
      /Users/mchian000/src/ears/internal/pkg/app/handlers_v1.go:102 +0x181
  github.com/xmidt-org/ears/internal/pkg/app.(*APIManager).removeRouteHandler-fm()
      /Users/mchian000/src/ears/internal/pkg/app/handlers_v1.go:98 +0x64
  net/http.HandlerFunc.ServeHTTP()
      /usr/local/Cellar/go/1.15.6/libexec/src/net/http/server.go:2042 +0x51
  github.com/gorilla/mux.(*Router).ServeHTTP()
      /Users/mchian000/go/pkg/mod/github.com/gorilla/mux@v1.8.0/mux.go:210 +0x132
  github.com/xmidt-org/ears/internal/pkg/app.TestRouteTable.func1()
      /Users/mchian000/src/ears/internal/pkg/app/handlers_v1_test.go:186 +0x1184
  testing.tRunner()
      /usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1123 +0x202

Previous read at 0x00c000422c30 by goroutine 13:
  runtime.mapiternext()
      /usr/local/Cellar/go/1.15.6/libexec/src/runtime/map.go:846 +0x0
  github.com/xmidt-org/ears/internal/pkg/plugin.(*manager).next()
      /Users/mchian000/src/ears/internal/pkg/plugin/manager.go:208 +0x269
  github.com/xmidt-org/ears/internal/pkg/plugin.(*manager).RegisterReceiver.func1.1()
      /Users/mchian000/src/ears/internal/pkg/plugin/manager.go:148 +0x6b
  github.com/xmidt-org/ears/pkg/plugins/debug.(*Receiver).Trigger()
      /Users/mchian000/src/ears/pkg/plugins/debug/receiver.go:164 +0xcf
  github.com/xmidt-org/ears/pkg/plugins/debug.(*Receiver).Receive.func1()
      /Users/mchian000/src/ears/pkg/plugins/debug/receiver.go:82 +0x567

Goroutine 10 (running) created at:
  testing.(*T).Run()
      /usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1168 +0x5bb
  github.com/xmidt-org/ears/internal/pkg/app.TestRouteTable()
      /Users/mchian000/src/ears/internal/pkg/app/handlers_v1_test.go:104 +0x6db
  testing.tRunner()
      /usr/local/Cellar/go/1.15.6/libexec/src/testing/testing.go:1123 +0x202

Goroutine 13 (finished) created at:
  github.com/xmidt-org/ears/pkg/plugins/debug.(*Receiver).Receive()
      /Users/mchian000/src/ears/pkg/plugins/debug/receiver.go:48 +0x129
  github.com/xmidt-org/ears/internal/pkg/plugin.(*manager).RegisterReceiver.func1()
      /Users/mchian000/src/ears/internal/pkg/plugin/manager.go:147 +0xd4

```

## Receiver Event Error Handling

Currently, receiver gets event errors in one of two ways:
* From the return value of `next` 
* From listening for `Nack` from downstream

[Debug receiver example](https://github.com/xmidt-org/ears/blob/issues51/pkg/plugins/debug/receiver.go)

<strong>Do we need both?</strong>
Can we just live with `Nack` as the main mechanism to convey event errors?

Why? Performance. The asynchronous nature of `Nack` means that we no longer need to block and wait for results as the events travel down the routes. The two part where we wait for downstream results are:
* Receivers fanning out to multiple routes [plugin manager](https://github.com/xmidt-org/ears/blob/issues51/internal/pkg/plugin/manager.go#L207-L226)
* Multiple events fanning out to a sender [route](https://github.com/xmidt-org/ears/blob/issues51/internal/pkg/plugin/manager.go#L207-L226)

The combining effects on codes above is that every event from receiver will have to wait for all downstream senders to respond before the receiver can move on to the next event (ie, synchronous).

## Other (Minor?) Issues
* https://github.com/xmidt-org/ears/blob/issues51/internal/pkg/plugin/manager.go#L207-L226
* https://github.com/xmidt-org/ears/blob/issues51/internal/pkg/app/routingTableManager.go#L144-L147
