# A Filter Chain Use Case

## Input Event

```
{
  "Type": "Notification",
  "MessageId": "e684b540-9101-530b-9e85-ed705c063207",
  "TopicArn": "topicarn_ACCOUNT",
  "Message": "ewogICJoZWFkZXIiOiB7CiAgICAidHlwZSI6ICJBQ0NPVU5UX1BST0RVQ1QiLAogICAgImV2ZW50IjogIkNSRUFURSIsCiAgICAicGFydG5lciI6ICJQYXJ0bmVyIiwKICAgICJ0cmFja2luZ0lkIjogIjEyMyIsCiAgICAidGltZXN0YW1wIjogIk5vdiAxOSwgMjAyMCA3OjQzOjAyIFBNIiwKICAgICJ0aW1lc3RhbXBNcyI6IDE2MDU4MTQ5ODI1MDMsCiAgICAic2NoZW1hIjogIjEuMCIsCiAgICAiYm9keUVuY3J5cHRlZCI6IGZhbHNlLAogICAgInN1Y2Nlc3MiOiB0cnVlCiAgfSwKICAiYm9keSI6IHsKICAgICJhY2NvdW50IjogewogICAgICAiaWQiOiAiMTIzIiwKICAgICAgInNvdXJjZSI6ICJBQkMiLAogICAgICAic291cmNlSWQiOiAiMTIzLklNUyIsCiAgICAgICJiaWxsaW5nQWNjb3VudElkIjogIjEyMyIsCiAgICAgICJzdGF0dXMiOiAiQWN0aXZlIiwKICAgICAgImVuYWJsZWQiOiB0cnVlCiAgICB9LAogICAgImFkZGVkQWNjb3VudFByb2R1Y3RzIjogWwogICAgICB7CiAgICAgICAgInNvdXJjZSI6ICJBQkMiLAogICAgICAgICJhY2NvdW50UHJvZHVjdElkIjogImEuYiIKICAgICAgfQogICAgXSwKICAgICJyZW1vdmVkQWNjb3VudFByb2R1Y3RzIjogW10sCiAgICAiYWNjb3VudFByb2R1Y3RzIjogWwogICAgICB7CiAgICAgICAgInNvdXJjZSI6ICJBQkMiLAogICAgICAgICJhY2NvdW50UHJvZHVjdElkIjogImEtYiIKICAgICAgfSwKICAgICAgewogICAgICAgICJzb3VyY2UiOiAiQUJDIiwKICAgICAgICAiYWNjb3VudFByb2R1Y3RJZCI6ICJhLmIiCiAgICAgIH0sCiAgICAgIHsKICAgICAgICAic291cmNlIjogIkFCQyIsCiAgICAgICAgImFjY291bnRQcm9kdWN0SWQiOiAiYS5iIgogICAgICB9LAogICAgICB7CiAgICAgICAgInNvdXJjZSI6ICJERUYiLAogICAgICAgICJhY2NvdW50UHJvZHVjdElkIjogImEuYiIKICAgICAgfSwKICAgICAgewogICAgICAgICJzb3VyY2UiOiAiQUJDIiwKICAgICAgICAiYWNjb3VudFByb2R1Y3RJZCI6ICJhLmIiCiAgICAgIH0sCiAgICAgIHsKICAgICAgICAic291cmNlIjogIkFCQyIsCiAgICAgICAgImFjY291bnRQcm9kdWN0SWQiOiAiYS5iIgogICAgICB9LAogICAgICB7CiAgICAgICAgInNvdXJjZSI6ICJBQkMiLAogICAgICAgICJhY2NvdW50UHJvZHVjdElkIjogImEuYiIKICAgICAgfSwKICAgICAgewogICAgICAgICJzb3VyY2UiOiAiR0hJIiwKICAgICAgICAiYWNjb3VudFByb2R1Y3RJZCI6ICJhLmIiCiAgICAgIH0KICAgIF0KICB9Cn0=",
  "Timestamp": "2020-11-19T19:43:03.236Z",
  "SignatureVersion": "1",
  "Signature": "signature",
  "SigningCertURL": "http://123.pem",
  "UnsubscribeURL": "http://unsubscribe",
  "MessageAttributes": {
    "header.partner": {
      "Type": "String",
      "Value": "Partner"
    },
    "header.event": {
      "Type": "String",
      "Value": "CREATE"
    }
  }
}
```

## Decoded Input Event

```
{
  "Type": "Notification",
  "MessageId": "e684b540-9101-530b-9e85-ed705c063207",
  "TopicArn": "topicarn_ACCOUNT",
  "Message": {
    "header": {
      "type": "ACCOUNT_PRODUCT",
      "event": "CREATE",
      "partner": "Partner",
      "trackingId": "123",
      "timestamp": "Nov 19, 2020 7:43:02 PM",
      "timestampMs": 1605814982503,
      "schema": "1.0",
      "bodyEncrypted": false,
      "success": true
    },
    "body": {
      "account": {
        "id": "123",
        "source": "ABC",
        "sourceId": "123.IMS",
        "billingAccountId": "123",
        "status": "Active",
        "enabled": true
      },
      "addedAccountProducts": [
        {
          "source": "ABC",
          "accountProductId": "a.b"
        }
      ],
      "removedAccountProducts": [],
      "accountProducts": [
        {
          "source": "ABC",
          "accountProductId": "a-b"
        },
        {
          "source": "ABC",
          "accountProductId": "a.b"
        },
        {
          "source": "ABC",
          "accountProductId": "a.b"
        },
        {
          "source": "DEF",
          "accountProductId": "a.b"
        },
        {
          "source": "ABC",
          "accountProductId": "a.b"
        },
        {
          "source": "ABC",
          "accountProductId": "a.b"
        },
        {
          "source": "ABC",
          "accountProductId": "a.b"
        },
        {
          "source": "GHI",
          "accountProductId": "a.b"
        }
      ]
    }
  },
  "Timestamp": "2020-11-19T19:43:03.236Z",
  "SignatureVersion": "1",
  "Signature": "signature",
  "SigningCertURL": "http://123.pem",
  "UnsubscribeURL": "http://unsubscribe",
  "MessageAttributes": {
    "header.partner": {
      "Type": "String",
      "Value": "Partner"
    },
    "header.event": {
      "Type": "String",
      "Value": "CREATE"
    }
  }
}
```

## Output Event

```
{
  "message": {
    "op": "process",
    "payload": {
      "body": {
        "addedAccountProducts": [
          {
            "accountProductId": "a.b",
            "source": "ABC"
          }
        ],
        "removedAccountProducts": []
      },
      "metadata": {
        "event": "CREATE",
        "id": "123",
        "timestamp": 1605814982503,
        "type": "ACCOUNT_PRODUCT"
      }
    },
    "uses": "xbo_account"
  },
  "to": {
    "app": "myapp",
    "location": "123"
  },
  "tx": {
    "traceId": "123-456-789-000"
  }
}
```


## Filter Rules

* only consider events from a topicArn ending in `_ACCOUNT` or `_ACCOUNTPRODUCT`
* only consider events with correct partner header and the header `type=ACCOUNT_PRODUCT`
* only forward the event if there is at least one account removed or added
