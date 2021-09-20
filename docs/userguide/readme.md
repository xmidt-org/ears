# EARS User Guide

The EARS service operates in the same problem domain as EEL but follows a more flexible and overall
superior design approach. As such, EARS is an asynchronous event _filtering_, _transformation_ and _routing_
service. The main idea is for EARS to offer a simple REST API that will allow self-service-style
onboarding of new event source for any type of service concerned with asynchronous event processing.
User interact with EARS by creating, modifying and removing routes using the EARS REST API. 
Internally EARS manages a global routing table and any changes to the routing table take effect 
immediately.

## Table of Contents

* Routes
* EARS API
* EARS vs. EEL  
* Simple Examples
* Debug Strategies
* Filter Plugin Reference
* Receiver Plugins Reference
* Sender Plugins Reference
* Plugin Developer Guide
