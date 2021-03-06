/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package recipe

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"
)

// RetryTimeout is an error implementation that is thrown
// when retry fails post timeout
type RetryTimeout struct {
	Err string
}

// Error implements error interface
func (rt *RetryTimeout) Error() string {
	return rt.Err
}

// Retryable helps executing user provided functions as
// conditions in a repeated manner till this condition succeeds
type Retryable struct {
	Message string

	WaitTimeout  time.Duration
	WaitInterval time.Duration
}

// RetryConfig helps in creating an instance of Retryable
type RetryConfig struct {
	WaitTimeout  *time.Duration
	WaitInterval *time.Duration
}

// NewRetry returns a new instance of Retryable
func NewRetry(config RetryConfig) *Retryable {
	var timeout, interval time.Duration
	if config.WaitTimeout != nil {
		timeout = *config.WaitTimeout
	} else {
		timeout = 60 * time.Second
	}
	if config.WaitInterval != nil {
		interval = *config.WaitInterval
	} else {
		interval = 1 * time.Second
	}
	return &Retryable{
		WaitTimeout:  timeout,
		WaitInterval: interval,
	}
}

// Waitf retries this provided function as a condition till
// this condition succeeds.
//
// Clients invoking this method need to return appropriate
// values in the function implementation to let this function
// to be either returned, or exited or retried.
func (r *Retryable) Waitf(
	condition func() (bool, error),
	message string,
	args ...interface{},
) error {
	context := fmt.Sprintf(
		message,
		args...,
	)
	// mark the start time
	start := time.Now()
	for {
		done, err := condition()
		if err == nil && done {
			klog.V(3).Infof(
				"Retryable condition succeeded: %s", context,
			)
			return nil
		}
		if err != nil && done {
			klog.V(3).Infof(
				"Retryable condition completed with error: %s: %s",
				context,
				err,
			)
			return err
		}
		if time.Since(start) > r.WaitTimeout {
			var errmsg = "No errors found"
			if err != nil {
				errmsg = fmt.Sprintf("%+v", err)
			}
			return &RetryTimeout{
				fmt.Sprintf(
					"Retryable condition timed out after %s: %s: %s",
					r.WaitTimeout,
					context,
					errmsg,
				),
			}
		}
		if err != nil {
			// Log error, but keep trying until timeout
			klog.V(3).Infof(
				"Retryable condition has errors: Will retry: %s: %s",
				context,
				err,
			)
		} else {
			klog.V(3).Infof(
				"Waiting for retryable condition to succeed: Will retry: %s",
				context,
			)
		}
		time.Sleep(r.WaitInterval)
	}
}
