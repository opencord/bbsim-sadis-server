/*
 * Copyright 2018-present Open Networking Foundation

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
	"fmt"
	"github.com/opencord/voltha-lib-go/v4/pkg/log"
	"sync"
)

type Store struct {
	olts sync.Map
	onus sync.Map
	bps  sync.Map
}

func NewStore() *Store {
	return &Store{
		olts: sync.Map{},
		onus: sync.Map{},
		bps:  sync.Map{},
	}
}

func (s *Store) addOlt(ctx context.Context, entry SadisOltEntry) {
	logger.Debugw(ctx, "adding-olt", log.Fields{"olt": entry})
	s.olts.Store(entry.ID, entry)
}

func (s *Store) addOnu(ctx context.Context, entry SadisOnuEntryV2) {
	logger.Debugw(ctx, "adding-onu", log.Fields{"onu": entry})
	s.onus.Store(entry.ID, entry)
}

func (s *Store) addBp(ctx context.Context, entry SadisBWPEntry) {
	logger.Debugw(ctx, "adding-bp", log.Fields{"bp": entry})
	s.bps.Store(entry.ID, entry)
}

func (s *Store) getOlt(ctx context.Context, id string) (*SadisOltEntry, error) {
	logger.Debugw(ctx, "getting-olt", log.Fields{"olt": id})
	if entry, ok := s.olts.Load(id); ok {
		e := entry.(SadisOltEntry)
		return &e, nil
	}
	return nil, fmt.Errorf("olt-not-found-in-store")
}

func (s *Store) getOnu(ctx context.Context, id string) (*SadisOnuEntryV2, error) {
	logger.Debugw(ctx, "getting-onu", log.Fields{"onu": id})
	if entry, ok := s.onus.Load(id); ok {
		e := entry.(SadisOnuEntryV2)
		return &e, nil
	}
	return nil, fmt.Errorf("onu-not-found-in-store")
}

func (s *Store) getBp(ctx context.Context, id string) (*SadisBWPEntry, error) {
	logger.Debugw(ctx, "getting-bp", log.Fields{"bp": id})
	if entry, ok := s.bps.Load(id); ok {
		e := entry.(SadisBWPEntry)
		return &e, nil
	}
	return nil, fmt.Errorf("bp-not-found-in-store")
}
