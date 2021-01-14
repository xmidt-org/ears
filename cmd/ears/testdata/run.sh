echo "add routes"

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRoute.json | jq .

# idempotency test

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRoute.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteBlankID.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleFilterRoute.json | jq .

echo "invalid routes"

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteBadName.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteBadPluginName.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoReceiver.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoSender.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoApp.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoOrg.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoUser.json | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/routes | jq .

curl -X GET http://localhost:3000/ears/v1/routes/r123 | jq .

curl -X GET http://localhost:3000/ears/v1/routes/foo | jq .

echo "delete routes"

curl -X DELETE http://localhost:3000/ears/v1/routes/2abe4d167883f594f7b48ee7e1d247ac  | jq .

curl -X DELETE http://localhost:3000/ears/v1/routes/r123  | jq .

curl -X DELETE http://localhost:3000/ears/v1/routes/f123  | jq .

# idempotency test

curl -X DELETE http://localhost:3000/ears/v1/routes/r123  | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/routes | jq .
