/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package checkconstraints

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructsmem/errs"
)

// Checks value by field constraints. Return error if constraints violated
func Check(fld appdef.IField, value any) (err error) {
	switch fld.DataKind() {
	case appdef.DataKind_string:
		err = checkChars(fld, value.(string))
	case appdef.DataKind_bytes:
		err = checkChars(fld, value.([]byte))
	case appdef.DataKind_int8:
		err = checkNumber(fld, value.(int8))
	case appdef.DataKind_int16:
		err = checkNumber(fld, value.(int16))
	case appdef.DataKind_int32:
		err = checkNumber(fld, value.(int32))
	case appdef.DataKind_int64:
		err = checkNumber(fld, value.(int64))
	case appdef.DataKind_float32:
		err = checkNumber(fld, value.(float32))
	case appdef.DataKind_float64:
		err = checkNumber(fld, value.(float64))
	}
	return err
}

// Checks string, bytes or raw value by field constraints. Return error if constraints violated
type chars interface{ string | []byte }

func checkChars[T chars](fld appdef.IField, value T) (err error) {
	maxLenChecked := false

	for k, c := range fld.Constraints() {
		switch k {
		case appdef.ConstraintKind_MinLen:
			if len(value) < int(c.Value().(uint16)) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
		case appdef.ConstraintKind_MaxLen:
			if len(value) > int(c.Value().(uint16)) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
			maxLenChecked = true
		case appdef.ConstraintKind_Pattern:
			if pat := c.Value().(*regexp.Regexp); pat != nil {
				switch fld.DataKind() {
				case appdef.DataKind_string:
					if !pat.MatchString(string(value)) {
						err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
					}
				case appdef.DataKind_bytes:
					if !pat.Match([]byte(value)) {
						err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
					}
				}
			}
		case appdef.ConstraintKind_Enum:
			if enum, ok := c.Value().([]string); ok {
				if l := len(enum); l > 0 {
					if _, ok := slices.BinarySearch(enum, string(value)); !ok {
						err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
					}
				}
			}
		}
	}

	if !maxLenChecked {
		if len(value) > int(appdef.DefaultFieldMaxLength) {
			err = errors.Join(err, errs.ErrDataConstraintViolation(fld, fmt.Sprintf("default MaxLen: %d", appdef.DefaultFieldMaxLength)))
		}
	}

	return err
}

// Checks value by number field constraints. Return error if constraints violated
type number = interface {
	int8 | int16 | int32 | int64 | float32 | float64
}

func checkNumber[T number](fld appdef.IField, value T) (err error) {
	for k, c := range fld.Constraints() {
		switch k {
		case appdef.ConstraintKind_MinIncl:
			if float64(value) < c.Value().(float64) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
		case appdef.ConstraintKind_MinExcl:
			if float64(value) <= c.Value().(float64) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
		case appdef.ConstraintKind_MaxIncl:
			if float64(value) > c.Value().(float64) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
		case appdef.ConstraintKind_MaxExcl:
			if float64(value) >= c.Value().(float64) {
				err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
			}
		case appdef.ConstraintKind_Enum:
			if enum, ok := c.Value().([]T); ok {
				if l := len(enum); l > 0 {
					if _, ok := sort.Find(l, func(i int) int {
						if value > enum[i] {
							return 1
						}
						if value < enum[i] {
							return -1
						}
						return 0
					}); !ok {
						err = errors.Join(err, errs.ErrDataConstraintViolation(fld, c))
					}
				}
			}
		}
	}

	return err
}
