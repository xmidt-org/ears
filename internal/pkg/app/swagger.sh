# SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
# SPDX-License-Identifier: Apache-2.0

# reading

# https://goswagger.io/faq/faq_spec.html
# https://medium.com/@pedram.esmaeeli/generate-swagger-specification-from-go-source-code-648615f7b9d9

# generate swagger yaml

swagger generate spec -o ./swagger.yaml --scan-models

# serve html documentation

swagger serve -F=swagger swagger.yaml
