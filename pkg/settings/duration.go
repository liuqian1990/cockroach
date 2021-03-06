// Copyright 2017 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package settings

import (
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// DurationSetting is the interface of a setting variable that will be
// updated automatically when the corresponding cluster-wide setting
// of type "duration" is updated.
type DurationSetting struct {
	common
	defaultValue time.Duration
	v            int64
	validateFn   func(time.Duration) error
}

var _ Setting = &DurationSetting{}

// Get retrieves the duration value in the setting.
func (d *DurationSetting) Get() time.Duration {
	return time.Duration(atomic.LoadInt64(&d.v))
}

func (d *DurationSetting) String() string {
	return EncodeDuration(d.Get())
}

// Typ returns the short (1 char) string denoting the type of setting.
func (*DurationSetting) Typ() string {
	return "d"
}

// Validate that a value conforms with the validation function.
func (d *DurationSetting) Validate(v time.Duration) error {
	if d.validateFn != nil {
		if err := d.validateFn(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DurationSetting) set(v time.Duration) error {
	if err := d.Validate(v); err != nil {
		return err
	}
	if v := int64(v); atomic.SwapInt64(&d.v, v) != v {
		d.changed()
	}
	return nil
}

func (d *DurationSetting) setToDefault() {
	if err := d.set(d.defaultValue); err != nil {
		panic(err)
	}
}

// RegisterDurationSetting defines a new setting with type duration.
func RegisterDurationSetting(key, desc string, defaultValue time.Duration) *DurationSetting {
	return RegisterValidatedDurationSetting(key, desc, defaultValue, nil)
}

// RegisterNonNegativeDurationSetting defines a new setting with type duration.
func RegisterNonNegativeDurationSetting(
	key, desc string, defaultValue time.Duration,
) *DurationSetting {
	return RegisterValidatedDurationSetting(key, desc, defaultValue, func(v time.Duration) error {
		if v < 0 {
			return errors.Errorf("cannot set %s to a negative duration: %s", key, v)
		}
		return nil
	})
}

// RegisterValidatedDurationSetting defines a new setting with type duration.
func RegisterValidatedDurationSetting(
	key, desc string, defaultValue time.Duration, validateFn func(time.Duration) error,
) *DurationSetting {
	if validateFn != nil {
		if err := validateFn(defaultValue); err != nil {
			panic(errors.Wrap(err, "invalid default"))
		}
	}
	setting := &DurationSetting{
		defaultValue: defaultValue,
		validateFn:   validateFn,
	}
	register(key, desc, setting)
	return setting
}

// TestingSetDuration returns a mock, unregistered string setting for testing.
// See TestingSetBool for more details.
func TestingSetDuration(s **DurationSetting, v time.Duration) func() {
	saved := *s
	*s = &DurationSetting{v: int64(v)}
	return func() {
		*s = saved
	}
}

// TestingDuration returns a one off, unregistered duration setting for test use
// only.
func TestingDuration(v time.Duration) *DurationSetting {
	return &DurationSetting{v: int64(v)}
}

// OnChange registers a callback to be called when the setting changes.
func (d *DurationSetting) OnChange(fn func()) *DurationSetting {
	d.setOnChange(fn)
	return d
}
