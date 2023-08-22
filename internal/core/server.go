/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"net/http"
	"sync"
)

type Server struct {
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) StartSadisServer(wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()

	addr := "0.0.0.0:8080"

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/subscribers/{ID}", s.serveEntry)
	router.HandleFunc("/profiles/{ID}", s.serveBWPEntry)

	logger.Fatal(ctx, http.ListenAndServe(addr, router))
}

func (s Server) serveEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]

	ctx := context.TODO()
	logger.Debugw(ctx, "received-sadis-entry-request", log.Fields{"id": id})

	w.Header().Set("Content-Type", "application/json")

	if olt, err := s.store.getOlt(r.Context(), id); err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(olt)
		logger.Infow(ctx, "responded-to-sadis-olt-entry-request", log.Fields{"id": id})
		return
	}

	if onu, err := s.store.getOnu(r.Context(), id); err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(onu)
		logger.Infow(ctx, "responded-to-sadis-onu-entry-request", log.Fields{"id": id})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	msg := make(map[string]interface{})
	msg["statusCode"] = http.StatusNotFound
	msg["message"] = fmt.Sprintf("Entry with ID %s not found.", id)
	_ = json.NewEncoder(w).Encode(msg)

	logger.Warnw(ctx, "sadis-entry-not-found", log.Fields{"id": id})
}

func (s Server) serveBWPEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]

	ctx := context.TODO()
	logger.Debugw(ctx, "received-sadis-bandwidthprofile-request", log.Fields{"id": id})

	w.Header().Set("Content-Type", "application/json")

	if bp, err := s.store.getBp(r.Context(), id); err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(bp)
		logger.Infow(ctx, "responded-to-sadis-bandwidthprofile-request", log.Fields{"id": id})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	msg := make(map[string]interface{})
	msg["statusCode"] = http.StatusNotFound
	msg["message"] = fmt.Sprintf("BandwidthProfile with ID %s not found.", id)
	_ = json.NewEncoder(w).Encode(msg)

	logger.Warnw(ctx, "sadis-bandwidthprofile-not-found", log.Fields{"id": id})
}
