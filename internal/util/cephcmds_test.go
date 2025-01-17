/*
Copyright 2021 The Ceph-CSI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecCommandWithTimeout(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx     context.Context
		program string
		timeout time.Duration
		args    []string
	}
	tests := []struct {
		name        string
		args        args
		stdout      string
		expectedErr error
		wantErr     bool
	}{
		{
			name: "echo hello",
			args: args{
				ctx:     context.TODO(),
				program: "echo",
				timeout: time.Second,
				args:    []string{"hello"},
			},
			stdout:      "hello\n",
			expectedErr: nil,
			wantErr:     false,
		},
		{
			name: "sleep with timeout",
			args: args{
				ctx:     context.TODO(),
				program: "sleep",
				timeout: time.Second,
				args:    []string{"3"},
			},
			stdout:      "",
			expectedErr: context.DeadlineExceeded,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stdout, _, err := ExecCommandWithTimeout(tt.args.ctx,
				tt.args.timeout,
				tt.args.program,
				tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommandWithTimeout() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("ExecCommandWithTimeout() error expected got = %v, want %v", err, tt.expectedErr)
			}

			if stdout != tt.stdout {
				t.Errorf("ExecCommandWithTimeout() got = %v, want %v", stdout, tt.stdout)
			}
		})
	}
}
