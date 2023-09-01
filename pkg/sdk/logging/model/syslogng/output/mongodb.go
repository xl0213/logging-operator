// Copyright © 2023 Cisco Systems, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

// +name:"MongoDB"
// +weight:"200"
type _hugoMongoDB interface{} //nolint:deadcode,unused

// +docName:"Sending messages from a local network to an MongoDB database"
//
// ## Prerequisites
//
// ## Example
//
// {{< highlight yaml >}}
// apiVersion: logging.banzaicloud.io/v1beta1
// kind: SyslogNGOutput
// metadata:
//
//	name: mongodb
//	namespace: default
//
// spec:
//
//	mongodb:
//	  collection: syslog
//	  uri: mongodb://127.0.0.1:27017/syslog?wtimeoutMS=60000&socketTimeoutMS=60000&connectTimeoutMS=60000
//	  value_pairs: scope("selected-macros" "nv-pairs")
//
// {{</ highlight >}}
type _docMongoDB interface{} //nolint:deadcode,unused

// +name:"MongoDB Destination"
// +url:"https://www.syslog-ng.com/technical-documents/doc/syslog-ng-open-source-edition/3.17/administration-guide/38#TOPIC-989689"
// +description:"Sending messages into MongoDB Server"
// +status:"Testing"
type _metaMongoDB interface{} //nolint:deadcode,unused

// +kubebuilder:object:generate=true
type MongoDB struct {
	//  The name of the MongoDB collection where the log messages are stored (collections are similar to SQL tables). Note that the name of the collection must not start with a dollar sign ($), and that it may contain dot (.) characters.
	Collection string `json:"collection"`
	//  If set to yes, syslog-ng OSE cannot lose logs in case of reload/restart, unreachable destination or syslog-ng OSE crash. This solution provides a slower, but reliable disk-buffer option.
	Compaction bool `json:"compaction"`
	// Defines the folder where the disk-buffer files are stored.
	Dir string `json:"dir,omitempty"`
	// This option enables putting outgoing messages into the disk buffer of the destination to avoid message loss in case of a system failure on the destination side. For details, see the [Syslog-ng DiskBuffer options](../disk_buffer/). (default: false)
	DiskBuffer *DiskBuffer `json:"disk_buffer,omitempty"`
	// Defines the folder where the disk-buffer files are stored. (default: "mongodb://127.0.0.1:27017/syslog?wtimeoutMS=60000&socketTimeoutMS=60000&connectTimeoutMS=60000")
	Uri string `json:"uri,omitempty"`
	// Creates structured name-value pairs from the data and metadata of the log message. (default: "scope("selected-macros" "nv-pairs")")
	ValuePairs ValuePairs `json:"value_pairs,omitempty"`
	// Batching parameters
	Batch `json:",inline"`
	// Bulk operation related options
	Bulk `json:",inline"`
}

// +kubebuilder:object:generate=true
// Bulk operation related options
type Bulk struct {
	// Enables bulk insert mode. If disabled, each messages is inserted individually. (default: yes)
	// See [syslog-ng docs] https://axoflow.com/docs/axosyslog-core/chapter-destinations/configuring-destinations-mongodb/reference-destination-mongodb/#mongodb-option-bulk
	Bulk *bool `json:"bulk,omitempty"`
	// If set to yes, it disables MongoDB bulk operations validation mode.
	BulkByPassValidation *bool `json:"bulk_bypass_validation,omitempty"`
}

// +kubebuilder:object:generate=true
// TODO move this to a common module once it is used in more places
type ValuePairs struct {
	Scope   RawString `json:"scope,omitempty"`
	Exclude RawString `json:"exclude,omitempty"`
	Key     RawString `json:"key,omitempty"`
	Pair    RawString `json:"pair,omitempty"`
}

// +kubebuilder:object:generate=true
type RawString struct {
	String string `json:"raw_string,omitempty"`
}