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

echo "dev"

curl -X PUT https://dev-us-west-2.gears.comcast.com/ears/v1/orgs/comcast/applications/myapp/config --data-binary @quota.json | jq .

echo "version"

curl -X GET http://localhost:3000/ears/version | jq .

echo "add quota"

curl -X PUT http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/config --data-binary @quota.json | jq .

echo "add fragments"

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/fragments --data-binary @debugSender.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/fragments --data-binary @debugFoobarReceiver.json | jq .

echo "add routes"

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteAA.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteBB.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteAB.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterChainMatchRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterSplitRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterDeepSplitRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterMatchAllowRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterMatchDenyRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterUnwrapRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @sqsReceiverRoute.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @sqsSenderRoute.json | jq .

# update route

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @update1.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @update2.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @update3.json | jq .

curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @update4.json | jq .

# idempotency test

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRoute.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteBlankID.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleFilterRoute.json | jq .

# echo "invalid routes"

# curl -X PUT http://localhost:3000/ears/v1/routes/orgs/comcast/applications/myapp/wrongid --data-binary @simpleRoute.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteBadName.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteBadPluginName.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteNoReceiver.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteNoSender.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteNoApp.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteNoOrg.json | jq .

# curl -X POST http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes --data-binary @simpleRouteNoUser.json | jq .

echo "get routes"

curl -X GET http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes | jq .

# curl -X GET http://localhost:3000/ears/v1/orgs/comcast/routes/r100 | jq .

# curl -X GET http://localhost:3000/ears/v1/orgs/comcast/routes/foo | jq .

# echo "delete routes"

# curl -X DELETE http://localhost:3000/ears/v1/orgs/comcast/routes/94d5eff28471968e9bd946bc9db27847  | jq .

# curl -X DELETE http://localhost:3000/ears/v1/orgs/comcast/routes/r100  | jq .

# curl -X DELETE http://localhost:3000/ears/v1/orgs/comcast/routes/f100  | jq .

# idempotency test

# curl -X DELETE http://localhost:3000/ears/v1/orgs/comcast/routes/r100  | jq .

# echo "get routes"

# curl -X GET http://localhost:3000/ears/v1/orgs/comcast/applications/myapp/routes | jq .
