# EARS vs. EEL

EARS plays in the same problem domain as [EEL](https://github.com/Comcast/eel): They both deal with 
asynchronous event processing and they both address the three key problems of event filtering, payload 
transformation and event routing. EARS can be thought of as EEL reimagined and the intention is for EARS 
to gradually replace EEL. Here are some of the key differences between the two:

| EARS | EEL | 
| --------------- | --------------- | 
| Dynamically add / remove routes via API call. | Manually manage handler files, restart service. |
| Pluggable persistence layer. DynamoDB, In-Memory and other implementations. | Only file based handlers. |
| Configurable event throughput quotas and rate limiting. | Assign worker pools to tenants. No configurable rate limiting. |
| Event acknowledgement and configurable retry policies. | Fire and forget. |
| Emphasis on traces and metrics for observability. Logs mainly for errors only. | Emphasis on logging. | 
| Modular routes: Receiver -> Filter -> Filter -> ... -> Sender | Monolithic handlers: Single complex transformation with deeply nested function calls. Sometimes requires more than one handler to get the job done. |
| Dynamic routing table. | Static set of handlers. |
| Two dimensional tenant management with org ID and app ID | One dimensional tenant management.
