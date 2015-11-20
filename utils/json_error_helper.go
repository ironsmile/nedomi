package utils

import (
	"encoding/json"

	"github.com/MStoykov/jsonutils"
)

// ShowContextOfJSONError is utility function that wrap *json.SyntaxError in an
// error which shows the context of the error.
// If the error provided is not *json.SyntaxError it's returned as it is.
func ShowContextOfJSONError(err error, jsonContents []byte) error {
	var synErr, ok = err.(*json.SyntaxError)
	if !ok { // nothing to do
		return err
	}

	return jsonutils.NewSyntaxError(synErr, jsonContents, 2)
}
