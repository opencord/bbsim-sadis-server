/*
 * Copyright 2018-2024 Open Networking Foundation (ONF) and the ONF Contributors

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"context"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"gotest.tools/assert"
	"testing"
)

func init() {
	SetupLogger(log.ErrorLevel, "console")
}

func Test_getBp(t *testing.T) {
	store := NewStore()

	bp := SadisBWPEntry{
		ID:  "test-bp",
		AIR: 20,
	}

	store.bps.Store(bp.ID, bp)

	ctx := context.TODO()
	loaded, err := store.getBp(ctx, bp.ID)

	assert.NilError(t, err)
	assert.Equal(t, loaded.ID, bp.ID)
	assert.Equal(t, loaded.AIR, bp.AIR)
}
