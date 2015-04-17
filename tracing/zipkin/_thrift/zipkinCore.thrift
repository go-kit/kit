# Copyright 2012 Twitter Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
namespace java com.twitter.zipkin.thriftjava
#@namespace scala com.twitter.zipkin.thriftscala
namespace rb Zipkin

#************** Collection related structs **************

# these are the annotations we always expect to find in a span
const string CLIENT_SEND = "cs"
const string CLIENT_RECV = "cr"
const string SERVER_SEND = "ss"
const string SERVER_RECV = "sr"

# this represents a host and port in a network
struct Endpoint {
  1: i32 ipv4,
  2: i16 port                      # beware that this will give us negative ports. some conversion needed
  3: string service_name           # which service did this operation happen on?
}

# some event took place, either one by the framework or by the user
struct Annotation {
  1: i64 timestamp                 # microseconds from epoch
  2: string value                  # what happened at the timestamp?
  3: optional Endpoint host        # host this happened on
  4: optional i32 duration         # how long did the operation take? microseconds
}

enum AnnotationType { BOOL, BYTES, I16, I32, I64, DOUBLE, STRING }

struct BinaryAnnotation {
  1: string key,
  2: binary value,
  3: AnnotationType annotation_type,
  4: optional Endpoint host
}

struct Span {
  1: i64 trace_id                  # unique trace id, use for all spans in trace
  3: string name,                  # span name, rpc method for example
  4: i64 id,                       # unique span id, only used for this span
  5: optional i64 parent_id,                # parent span id
  6: list<Annotation> annotations, # list of all annotations/events that occured
  8: list<BinaryAnnotation> binary_annotations # any binary annotations
  9: optional bool debug = 0       # if true, we DEMAND that this span passes all samplers
}

