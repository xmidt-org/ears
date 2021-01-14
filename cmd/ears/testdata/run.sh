echo "add routes"

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRoute.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteBlankID.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteBadName.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoReceiver.json | jq .

curl -X PUT http://localhost:3000/ears/v1/routes --data @testdata/simpleRouteNoSender.json | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/routes | jq .

echo "delete routes"

curl -X DELETE http://localhost:3000/ears/v1/routes/f7e3d975d08ae28005ac916dbeb5888e  | jq .

curl -X DELETE http://localhost:3000/ears/v1/routes/r123  | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/routes | jq .
