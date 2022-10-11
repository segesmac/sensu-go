package globalid

import corev2 "github.com/sensu/core/v2"

//
// Roles
//
var roleName = "roles"

// RoleTranslator global ID resource
var RoleTranslator = commonTranslator{
	name:       roleName,
	encodeFunc: standardEncoder(roleName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.Role)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(RoleTranslator) }
