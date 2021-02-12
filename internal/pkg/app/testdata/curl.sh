# Copyright 2021 Comcast Cable Communications Management, LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "version"

curl -X GET http://localhost:3000/ears/version | jq .

echo "add routes"

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteAA.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteBB.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteAB.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterChainMatchRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterSplitRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterDeepSplitRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterMatchAllowRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterMatchDenyRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterUnwrapRoute.json | jq .

# idempotency test

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRoute.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteBlankID.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleFilterRoute.json | jq .

# echo "invalid routes"

# curl -X PUT http://localhost:3000/ears/v1/routes/wrongid --data @simpleRoute.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteBadName.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteBadPluginName.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteNoReceiver.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteNoSender.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteNoApp.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteNoOrg.json | jq .

# curl -X POST http://localhost:3000/ears/v1/routes --data @simpleRouteNoUser.json | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/routes | jq .

# curl -X GET http://localhost:3000/ears/v1/routes/r100 | jq .

# curl -X GET http://localhost:3000/ears/v1/routes/foo | jq .

# echo "delete routes"

# curl -X DELETE http://localhost:3000/ears/v1/routes/94d5eff28471968e9bd946bc9db27847  | jq .

# curl -X DELETE http://localhost:3000/ears/v1/routes/r100  | jq .

# curl -X DELETE http://localhost:3000/ears/v1/routes/f100  | jq .

# idempotency test

# curl -X DELETE http://localhost:3000/ears/v1/routes/r100  | jq .

# echo "get routes"

# curl -X GET http://localhost:3000/ears/v1/routes | jq .
