/*
 * Copyright 2020-present Open Networking Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//Package l2oam Common Logger initialization
package l2oam

import (
	"context"
	"fmt"
	"time"
)

//var logger log.CLogger
var logger l2OAMLogger

// func init() {
// 	// Setup this package so that it's log level can be modified at run time
// 	var err error
// 	logger, err = log.RegisterPackage(log.JSON, log.ErrorLevel, log.Fields{})
// 	if err != nil {
// 		panic(err)
// 	}
// }

func init() {
	logger = l2OAMLogger{}
	logger.Debug(context.Background(), "initialize l2oam logger")
}

//l2OAMLogger ...
type l2OAMLogger struct {
}

//Debug ...
func (l *l2OAMLogger) Debug(ctx context.Context, message string) {
	fmt.Printf("%v:%s\n", time.Now(), message)
}
