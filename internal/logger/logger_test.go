// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"log"
	"strings"
	"testing"
	
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/output"
	"github.com/saucelabs/sypl/processor"
	"github.com/saucelabs/sypl/shared"
)

func TestSetup(t *testing.T) {
	type args struct {
		lvl       string
		fileLevel string
		filePath  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Should work",
			args: args{
				lvl:       "",
				fileLevel: "",
				filePath:  "",
			},
		},
		{
			name: "Should work",
			args: args{
				lvl:       infoLevel,
				fileLevel: infoLevel,
				filePath:  "-",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Setup(&Options{
				Level:     tt.args.lvl,
				FileLevel: tt.args.fileLevel,
				FilePath:  tt.args.filePath,
			})
			if l == nil {
				t.Errorf("Setup() expect %v to don't be nil", l)
			}

			if retrievedL := Get(); retrievedL == nil {
				t.Errorf("Get() expect %v to don't be nil", l)
			}

			// Should do nothing.
			l = Get()
			if l == nil {
				t.Errorf("Setup() expect %v to don't be nil", l)
			}
		})
	}
}


func TestRedirectStandardLogs(t *testing.T) {
	// set global proxy logger
	proxyLogger = sypl.NewDefault("test", level.Debug)
	defer func(){proxyLogger = nil}()

	// proxyLogger sends output to a buffer
	buffer, outputBuffer := output.SafeBuffer(level.Trace, processor.PrefixBasedOnMask(shared.DefaultTimestampFormat))
	proxyLogger.AddOutputs(outputBuffer)

	// test standard logger before and after redirect
	beforeMsg := "Before redirect"
	afterMsg := "After redirect" 
	log.Println(beforeMsg)
	
	RedirectStandardLogs()
	log.Println(afterMsg)

	bufferStr := buffer.String()

	if strings.Contains(bufferStr, beforeMsg) {
		t.Errorf("%s should not appear in proxy logger: %s", beforeMsg, bufferStr)
	}
	if !strings.Contains(bufferStr, afterMsg) {
		t.Errorf("%s should appear in proxy logger: %s", afterMsg, bufferStr)
	}

}