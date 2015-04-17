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

include "zipkinCore.thrift"
include "zipkinDependencies.thrift"

struct Trace {
  1: list<zipkinCore.Span> spans
}

exception QueryException {
  1: string msg
}

struct SpanTimestamp {
  1: string name
  2: i64 start_timestamp
  3: i64 end_timestamp
}

/**
 * This sums up a single Trace to make it easy for a client to get an overview of what happened.
 */
struct TraceSummary {
  1: i64 trace_id                  # the trace
  2: i64 start_timestamp           # start timestamp of the trace, in microseconds
  3: i64 end_timestamp             # end timestamp of the trace, in microseconds
  4: i32 duration_micro            # how long did the entire trace take? in microseconds
  # 5: map<string, i32> service_counts     # which services were involved?
  6: list<zipkinCore.Endpoint> endpoints      # which endpoints were involved?
  7: list<SpanTimestamp> span_timestamps
}

/**
 * A modified version of the Annotation struct that brings in more information
 */
struct TimelineAnnotation {
  1: i64 timestamp                 # microseconds from epoch
  2: string value                  # what happened at the timestamp?
  3: zipkinCore.Endpoint host      # host this happened on
  4: i64 span_id                   # which span does this annotation belong to?
  5: optional i64 parent_id        # parent span id
  6: string service_name           # which service did this annotation happen on?
  7: string span_name              # span name, rpc method for example
}

/**
 * This sums up a single Trace to make it easy for a client to get an overview of what happened.
 */
struct TraceTimeline {
  1: i64 trace_id                          # the trace
  2: i64 root_most_span_id                 # either the true root span or the closest we can find
  6: list<TimelineAnnotation> annotations  # annotations as they happened
  7: list<zipkinCore.BinaryAnnotation> binary_annotations # all the binary annotations
}

/**
 * Returns a combination of trace, summary and timeline.
 */
struct TraceCombo {
  1: Trace trace
  2: optional TraceSummary summary # not set if no spans in trace
  3: optional TraceTimeline timeline # not set if no spans in trace
  4: optional map<i64, i32> span_depths # not set if no spans in trace
}

enum Order { TIMESTAMP_DESC, TIMESTAMP_ASC, DURATION_ASC, DURATION_DESC, NONE }

/**
 * The raw data in our storage might have various problems. How should we adjust it before
 * returning it to the user?
 *
 * Time skew adjuster tries to make sure that even though servers might have slightly
 * different clocks the annotations in the returned data are adjusted so that they are
 * in the correct order.
 */
enum Adjust { NOTHING, TIME_SKEW }

struct QueryRequest {
  1: string service_name
  2: optional string span_name
  3: optional list<string> annotations
  4: optional list<zipkinCore.BinaryAnnotation> binary_annotations
  5: i64 end_ts
  6: i32 limit
  7: Order order
}

struct QueryResponse {
  1: list<i64> trace_ids
  2: i64 start_ts
  3: i64 end_ts
}

service ZipkinQuery {

    #************** Index lookups **************

    QueryResponse getTraceIds(1: QueryRequest request) throws (1: QueryException qe);

    /**
     * Fetch trace ids by service and span name.
     * Gets "limit" number of entries from before the "end_ts".
     *
     * Span name is optional.
     * Timestamps are in microseconds.
     */
    list<i64> getTraceIdsBySpanName(1: string service_name, 2: string span_name,
        4: i64 end_ts, 5: i32 limit, 6: Order order) throws (1: QueryException qe);

    /**
     * Fetch trace ids by service name.
     * Gets "limit" number of entries from before the "end_ts".
     *
     * Timestamps are in microseconds.
     */
    list<i64> getTraceIdsByServiceName(1: string service_name,
        3: i64 end_ts, 4: i32 limit, 5: Order order) throws (1: QueryException qe);

    /**
     * Fetch trace ids with a particular annotation.
     * Gets "limit" number of entries from before the "end_ts".
     *
     * When requesting based on time based annotations only pass in the first parameter, "annotation" and leave out
     * the second "value". If looking for a key-value binary annotation provide both, "annotation" is then the
     * key in the key-value.
     *
     * Timestamps are in microseconds.
     */
    list<i64> getTraceIdsByAnnotation(1: string service_name, 2: string annotation, 3: binary value,
        5: i64 end_ts, 6: i32 limit, 7: Order order) throws (1: QueryException qe);


    #************** Fetch traces from id **************

    /**
     * Get the traces that are in the database from the given list of trace ids.
     */

    set<i64> tracesExist(1: list<i64> trace_ids) throws (1: QueryException qe);

    /**
     * Get the full traces associated with the given trace ids.
     *
     * Second argument is a list of methods of adjusting the trace
     * data before returning it. Can be empty.
     */
    list<Trace> getTracesByIds(1: list<i64> trace_ids, 2: list<Adjust> adjust) throws (1: QueryException qe);

    /**
     * Get the trace timelines associated with the given trace ids.
     * This is a convenience method for users that just want to know
     * the annotations and the (assumed) order they happened in.
     *
     * Second argument is a list of methods of adjusting the trace
     * data before returning it. Can be empty.
     *
     * Note that if one of the trace ids does not have any data associated with it, it will not be
     * represented in the output list.
     */
    list<TraceTimeline> getTraceTimelinesByIds(1: list<i64> trace_ids, 2: list<Adjust> adjust) throws (1: QueryException qe);

    /**
     * Fetch trace summaries for the given trace ids.
     *
     * Second argument is a list of methods of adjusting the trace
     * data before returning it. Can be empty.
     *
     * Note that if one of the trace ids does not have any data associated with it, it will not be
     * represented in the output list.
     */
    list<TraceSummary> getTraceSummariesByIds(1: list<i64> trace_ids, 2: list<Adjust> adjust) throws (1: QueryException qe);

    /**
     * Not content with just one of traces, summaries or timelines? Want it all? This is the method for you.
     */
    list<TraceCombo> getTraceCombosByIds(1: list<i64> trace_ids, 2: list<Adjust> adjust) throws (1: QueryException qe);

    #************** Misc metadata **************

    /**
     * Fetch all the service names we have seen from now all the way back to the set ttl.
     */
    set<string> getServiceNames() throws (1: QueryException qe);

    /**
     * Get all the seen span names for a particular service, from now back until the set ttl.
     */
    set<string> getSpanNames(1: string service_name) throws (1: QueryException qe);

    #************** TTL related **************

    /**
     * Change the TTL of a trace. If we find an interesting trace we want to keep around for further
     * investigation.
     */
    void setTraceTimeToLive(1: i64 trace_id, 2: i32 ttl_seconds) throws (1: QueryException qe);

    /**
     * Get the TTL in seconds of a specific trace.
     */
    i32 getTraceTimeToLive(1: i64 trace_id) throws (1: QueryException qe);

    /**
     * Get the data ttl. This is the number of seconds we keep the data around before deleting it.
     */
    i32 getDataTimeToLive() throws (1: QueryException qe);

    /**
     * Get an aggregate representation of all services paired with every service they call in to.
     * This includes information on call counts and mean/stdDev/etc of call durations.  The two arguments
     * specify epoch time in microseconds. The end time is optional and defaults to one day after the
     * start time.
     */
    zipkinDependencies.Dependencies getDependencies(1: optional i64 start_time, 2: optional i64 end_time) throws (1: QueryException qe);

    list<string> getTopAnnotations(1: string service_name) throws (1: QueryException qe);
    list<string> getTopKeyValueAnnotations(1: string service_name) throws (1: QueryException qe);

    /**
     * Given a time stamp, server service name, and rpc name, fetch all of the client services calling in paired
     * with the lists of every span duration (list<i64>) from the server to client. The lists of span durations
     * include information on call counts and mean/stdDev/etc of call durations.
     *
     * The three arguments specify epoch time in microseconds, server side service name and rpc name. The return maps
     * contains the key - client_service_name and value - list<span_durations>.
     */
     map<string, list<i64>> getSpanDurations(1: i64 time_stamp, 2: string service_name, 3: string rpc_name);

    /**
     * Given a time stamp, server service name, and rpc name, fetch all of the client services calling in paired
     * with the lists of every trace Ids (list<i64>) from the server to client.
     *
     * The three arguments specify epoch time in microseconds, server side service name and rpc name. The return maps
     * contains the key - client_service_name and value - list<trace_id>.
     */
     map<string, list<i64>> getServiceNamesToTraceIds(1: i64 time_stamp, 2: string service_name, 3: string rpc_name);
}
