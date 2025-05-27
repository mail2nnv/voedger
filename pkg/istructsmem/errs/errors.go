/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 */

package errs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
)

// Enriches error with additional information.
//
// argOrMsg is any value to be added to the error message, and args are additional values to be added to the error message.
// Spaces are added between args.
//
// If argOrMsg is a string contains `%` and args is not empty, then argOrMsg is treated as a format string
func Enrich(err error, argOrMsg any, args ...any) error {
	var enrich string
	if msg, ok := argOrMsg.(string); ok && len(args) > 0 && strings.Contains(msg, "%") {
		enrich = fmt.Sprintf(msg, args...)
	} else {
		enrich = fmt.Sprint(argOrMsg)
		for i := range args {
			enrich += " " + fmt.Sprint(args[i])
		}
	}
	return fmt.Errorf("%w: %s", err, enrich)
}

var ErrorEventNotValidError = errors.New("event is not valid")

func ErrorEventNotValid(argOrMsg any, args ...any) error {
	return Enrich(ErrorEventNotValidError, argOrMsg, args...)
}

var ErrNameMissedError = errors.New("name is empty")

func ErrNameMissed(argOrMsg any, args ...any) error {
	return Enrich(ErrNameMissedError, argOrMsg, args...)
}

var ErrOutOfBoundsError = errors.New("out of bounds")

func ErrOutOfBounds(argOrMsg any, args ...any) error {
	return Enrich(ErrOutOfBoundsError, argOrMsg, args...)
}

var ErrWrongTypeError = errors.New("wrong type")

func ErrWrongType(argOrMsg any, args ...any) error {
	return Enrich(ErrWrongTypeError, argOrMsg, args...)
}

var ErrNameNotFoundError = errors.New("name not found")

func ErrNameNotFound(argOrMsg any, args ...any) error {
	return Enrich(ErrNameNotFoundError, argOrMsg, args...)
}

func ErrFieldNotFound(name string, typ interface{}) error {
	return Enrich(ErrNameNotFoundError, "field «%s» in %v", name, typ)
}

func ErrTypedFieldNotFound(t, f string, typ interface{}) error {
	return Enrich(ErrNameNotFoundError, "%s-field «%s» in %v", t, f, typ)
}

func ErrContainerNotFound(name string, typ interface{}) error {
	return Enrich(ErrNameNotFoundError, "container «%s» in %v", name, typ)
}

// name should  be string or any Stringer interface (e.g. QName)
func ErrTypeNotFound(name interface{}) error {
	return Enrich(ErrNameNotFoundError, "type «%v»", name)
}

// name should  be string or any Stringer interface (e.g. QName)
func ErrViewNotFound(name interface{}) error {
	return Enrich(ErrNameNotFoundError, "view «%v»", name)
}

var ErrInvalidNameError = errors.New("name not valid")

func ErrInvalidName(argOrMsg any, args ...any) error {
	return Enrich(ErrInvalidNameError, argOrMsg, args...)
}

var ErrIDNotFoundError = errors.New("ID not found")

func ErrIDNotFound(argOrMsg any, args ...any) error {
	return Enrich(ErrIDNotFoundError, argOrMsg, args...)
}

func ErrRefIDNotFound(t interface{}, f string, id istructs.RecordID) error {
	return ErrIDNotFound("%v field «%s» refers to unknown ID «%d»", t, f, id)
}

var ErrRecordNotFound = errors.New("record cannot be found")

var ErrMinOccursViolationError = errors.New("minimum occurs violated")

func ErrMinOccursViolated(t interface{}, n string, o, minO appdef.Occurs) error {
	return Enrich(ErrMinOccursViolationError, "%v container «%s» has not enough occurrences (%d, minimum %d)", t, n, o, minO)
}

var ErrMaxOccursViolationError = errors.New("maximum occurs violated")

func ErrMaxOccursViolated(t interface{}, n string, o, maxO appdef.Occurs) error {
	return Enrich(ErrMaxOccursViolationError, "%v container «%s» has too many occurrences (%d, maximum %d)", t, n, o, maxO)
}

var ErrFieldIsEmptyError = errors.New("field is empty")

// name should  be string or any Stringer interface (e.g. IField)
func ErrFieldIsEmpty(name interface{}) error {
	return Enrich(ErrFieldIsEmptyError, "%v", name)
}

func ErrFieldMissed(t, name interface{}) error {
	return Enrich(ErrFieldIsEmptyError, "%v %v", t, name)
}

var ErrInvalidVerificationKindError = errors.New("invalid verification kind")

func ErrInvalidVerificationKind(t, f interface{}, k appdef.VerificationKind) error {
	return Enrich(ErrInvalidVerificationKindError, "%s for %v «%v»", k.TrimString(), t, f)
}

var ErrCUDsMissedError = errors.New("CUDs are missed")

// event should be string or any Stringer interface (e.g. IEvent)
func ErrCUDsMissed(event interface{}) error {
	return Enrich(ErrCUDsMissedError, "%v", event)
}

var ErrRawRecordIDRequiredError = errors.New("raw record ID required")

func ErrRawRecordIDRequired(row, fld interface{}, id istructs.RecordID) error {
	return Enrich(ErrRawRecordIDRequiredError, "%v %v: id «%d» is not raw", row, fld, id)
}

var ErrUnexpectedRawRecordIDError = errors.New("unexpected raw record ID")

func ErrUnexpectedRawRecordID(rec, fld interface{}, id istructs.RecordID) error {
	return Enrich(ErrUnexpectedRawRecordIDError, "%v %v: id «%d» should not be raw", rec, fld, id)
}

var ErrRecordIDUniqueViolationError = errors.New("record ID duplicates")

func ErrRecordIDUniqueViolation(id istructs.RecordID, rec, dupe interface{}) error {
	return Enrich(ErrRecordIDUniqueViolationError, "id «%d» used by %v and %v", id, rec, dupe)
}

// name should  be string or any Stringer interface (e.g. QName)
func ErrSingletonViolation(name interface{}) error {
	return Enrich(ErrRecordIDUniqueViolationError, "singleton «%v» violation", name)
}

var ErrWrongRecordIDError = errors.New("wrong record ID")

func ErrWrongRecordID(argOrMsg any, args ...any) error {
	return Enrich(ErrWrongRecordIDError, argOrMsg, args...)
}

func ErrWrongRecordIDTarget(t, f interface{}, id istructs.RecordID, target interface{}) error {
	return Enrich(ErrWrongRecordIDError, "%v %v refers to record ID «%d» that has wrong target «%s»", t, f, id, target)
}

var ErrUnableToUpdateSystemFieldError = errors.New("unable to update system field")

func ErrUnableToUpdateSystemField(t, f interface{}) error {
	return Enrich(ErrUnableToUpdateSystemFieldError, "%v %v", t, f)
}

var ErrAbstractTypeError = errors.New("abstract type")

func ErrAbstractType(argOrMsg any, args ...any) error {
	return Enrich(ErrAbstractTypeError, argOrMsg, args...)
}

var ErrUnexpectedTypeError = errors.New("unexpected type")

func ErrUnexpectedType(argOrMsg any, args ...any) error {
	return Enrich(ErrUnexpectedTypeError, argOrMsg, args...)
}

var ErrUnknownCodecError = errors.New("unknown codec")

func ErrUnknownCodec(argOrMsg any, args ...any) error {
	return Enrich(ErrUnknownCodecError, argOrMsg, args...)
}

var ErrWrongFieldTypeError = errors.New("wrong field type")

func ErrWrongFieldType(argOrMsg any, args ...any) error {
	return Enrich(ErrWrongFieldTypeError, argOrMsg, args...)
}

var ErrDataConstraintViolationError = errors.New("data constraint violation")

func ErrDataConstraintViolation(field, constraint interface{}) error {
	return Enrich(ErrDataConstraintViolationError, "%v: %v", field, constraint)
}

var ErrNumAppWorkspacesNotSetError = errors.New("NumAppWorkspaces is not set")

func ErrNumAppWorkspacesNotSet(app interface{}) error {
	return Enrich(ErrNumAppWorkspacesNotSetError, app)
}

var ErrCorruptedData = errors.New("corrupted data")

// ValidateError: an interface for describing errors that occurred during validation
//   - methods:
//     — Code(): returns error code, see ECode_××× constants
type ValidateError interface {
	error
	Code() int
}

var ErrSequencesViolation = errors.New("sequences violation")
