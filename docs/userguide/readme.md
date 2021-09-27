# EARS User Guide

The EARS service operates in the same problem domain as [EEL](https://github.com/Comcast/eel) but follows a more 
flexible and overall superior design approach. As such, EARS is an asynchronous 
event _filtering_, _transformation_ and _routing_
service. The main idea is for EARS to offer a simple REST API that allows self-service-style
on-boarding of new event source for any type of service concerned with asynchronous event processing.
Users interact with EARS by creating, modifying and removing routes using the EARS REST API. 
Internally EARS manages a global routing table and any changes to the routing table take effect 
immediately.

## Table of Contents

* [Routes](routes.md)
* [EARS API](api.md)
* [EARS vs. EEL](eel.md)  
* [Simple Examples](examples.md)
* [Debug Strategies](debug.md)
* [Filter Plugin Reference](filters.md)
* [Receiver Plugins Reference](receivers.md)
* [Sender Plugins Reference](senders.md)
* [Plugin Developer Guide](plugindev.md)
