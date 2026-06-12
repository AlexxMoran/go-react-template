// Package api embeds the OpenAPI specification (the source of truth for the HTTP
// contract) so the server can serve it and the contract test can validate
// against it.
package api

import _ "embed"

// Spec is the raw OpenAPI 3 document (api/openapi.yaml).
//
//go:embed openapi.yaml
var Spec []byte
