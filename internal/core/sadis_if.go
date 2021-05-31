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

// TODO this should be imported from github.com/opencord/bbsim
// but to do that we need to move them from the "internal" module

type SadisConfig struct {
	Sadis            SadisEntries            `json:"sadis"`
	BandwidthProfile BandwidthProfileEntries `json:"bandwidthprofile"`
}

type SadisEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []*SadisEntry    `json:"entries,omitempty"`
	//Entries     []interface{}    `json:"entries,omitempty"`
}

type BandwidthProfileEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []*SadisBWPEntry `json:"entries,omitempty"`
}

type SadisIntegration struct {
	URL   string `json:"url,omitempty"`
	Cache struct {
		Enabled bool   `json:"enabled"`
		MaxSize int    `json:"maxsize"`
		TTL     string `json:"ttl"`
	} `json:"cache"`
}

type SadisEntry struct {
	// common
	ID string `json:"id"`
	// olt
	HardwareIdentifier string `json:"hardwareIdentifier"`
	IPAddress          string `json:"ipAddress"`
	NasID              string `json:"nasId"`
	UplinkPort         int    `json:"uplinkPort"`
	// onu
	NasPortID  string        `json:"nasPortId"`
	CircuitID  string        `json:"circuitId"`
	RemoteID   string        `json:"remoteId"`
	UniTagList []SadisUniTag `json:"uniTagList"`
}

type SadisOltEntry struct {
	ID                 string `json:"id"`
	HardwareIdentifier string `json:"hardwareIdentifier"`
	IPAddress          string `json:"ipAddress"`
	NasID              string `json:"nasId"`
	UplinkPort         int    `json:"uplinkPort"`
}

type SadisOnuEntryV2 struct {
	ID         string        `json:"id"`
	NasPortID  string        `json:"nasPortId"`
	CircuitID  string        `json:"circuitId"`
	RemoteID   string        `json:"remoteId"`
	UniTagList []SadisUniTag `json:"uniTagList"` // this can be SadisUniTagAtt, SadisUniTagDt
}

type SadisUniTag struct {
	UniTagMatch                int    `json:"uniTagMatch,omitempty"`
	PonCTag                    int    `json:"ponCTag,omitempty"`
	PonSTag                    int    `json:"ponSTag,omitempty"`
	TechnologyProfileID        int    `json:"technologyProfileId,omitempty"`
	UpstreamBandwidthProfile   string `json:"upstreamBandwidthProfile,omitempty"`
	DownstreamBandwidthProfile string `json:"downstreamBandwidthProfile,omitempty"`
	IsDhcpRequired             bool   `json:"isDhcpRequired,omitempty"`
	IsIgmpRequired             bool   `json:"isIgmpRequired,omitempty"`
	ConfiguredMacAddress       string `json:"configuredMacAddress,omitempty"`
	UsPonCTagPriority          uint8  `json:"usPonCTagPriority,omitempty"`
	UsPonSTagPriority          uint8  `json:"usPonSTagPriority,omitempty"`
	DsPonCTagPriority          uint8  `json:"dsPonCTagPriority,omitempty"`
	DsPonSTagPriority          uint8  `json:"dsPonSTagPriority,omitempty"`
	ServiceName                string `json:"serviceName,omitempty"`
}

// SADIS BandwithProfile Entry
type SadisBWPEntry struct {
	ID  string `json:"id"`
	CBS int    `json:"cbs"`
	CIR int    `json:"cir"`
	// MEF attributes
	AIR int    `json:"air,omitempty"`
	EBS int    `json:"ebs,omitempty"`
	EIR int    `json:"eir,omitempty"`
	// IETF attributes
	GIR int    `json:"gir,omitempty"`
	PIR int    `json:"pir,omitempty"`
	PBS int    `json:"pbs,omitempty"`
}
