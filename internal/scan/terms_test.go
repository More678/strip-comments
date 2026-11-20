package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
)

func TestStripWithTerms(t *testing.T) {
	tb := glossary.NewTable(glossary.Default)
	const idsLookalikes = "var ids = \"host-based IDs; network-based ids; my_host-based IDS; network-based IDS2\"\n" +
		"// IDS, HIDS, NIDS, IPS, host-based IDPS, host-based  IDS, network-based IDS后, 主机型 入侵检测系统, 网络型入侵防御系统\n"
	const retentionLookalikes = "var label = \"my_audit record retention period; audit record retention period2; audit_record_retention_period\"\n" +
		"// audit record retention periods; audit record retention period后; audit record  retention period\n" +
		"// audit record retention policy; log retention period; backup retention period; AU-11; 审计记录保留期限\n"
	const timestampLookalikes = "var timestamp = \"my_timestamps for audit records; timestamps for audit records2; timestamps_for_audit_records\"\n" +
		"// timestamps for audit records后; timestamp for audit records; timestamps for audit record; timestamps  for audit records\n" +
		"// timestamps for log files; timestamp; timestamps; 审计记录时间戳; 时间戳; AU-8\n"
	const eventResultLookalikes = "var result = \"my_audit event result; audit event result2; audit_event_result\"\n" +
		"// audit event results; audit  event result; audit event result后; AuditEventResult; event.outcome\n" +
		"// audit log write result; outcome; result; success; failure; 成功; 失败\n"
	const auditCapacityLookalikes = "var capacity = \"reallocate audit log storage capacity; allocate_audit_log_storage_capacity\"\n" +
		"// allocate audit log storage capacity2; allocate audit log storage capacity后; allocate  audit log storage capacity\n" +
		"// maximum audit log storage capacity; remaining audit log storage capacity; audit log retention period; 容量\n"
	const auditDiskLookalikes = "var msg = \"the audit log is almost full; the disk with the audit log is full; audit log write failed\"\n" +
		"// _the disk with the audit log is almost full; the disk with the audit log is almost full2; the disk with the audit log is almost full后\n" +
		"// the disk with the audit log is almost  full; space_left_action; disk_full_action; disk_error_action\n"
	const auditRotationLookalikes = "var labels = \"backup of audit log files; archival of audit log files; deletion of audit log files\"\n" +
		"// log rotation; audit log retention; logrotate; rotation_of_audit_log_files\n" +
		"// _rotation of audit log files; rotation of audit log files2; rotation of audit log files后; rotation of audit  log files\n"
	const auditCompressionLookalikes = "var labels = \"compression of active audit log files; archival of rotated audit log files; encryption of rotated audit log files\"\n" +
		"// decompression of rotated audit log files; compression of rotated audit log files2; compression of rotated audit log files后\n" +
		"// compression_of_rotated_audit_log_files; compression of  rotated audit log files; --audit-log-compress; gzip\n"
	const auditIntegrityLookalikes = "var status = \"audit log file integrity verified; integrity check failed; not verified\"\n" +
		"// verify_audit_log_file_integrity; _verify audit log file integrity; verify audit log file integrity2; verify audit log file integrity后\n" +
		"// verify  audit log file integrity; enable audit log file integrity validation; verify file integrity; 审计日志文件完整性校验\n"
	const auditReadOnlyLookalikes = "var read = \"provides read access to audit records; provides read-write access to audit records; audit records were read\"\n" +
		"// read-only access to audit records; immutable audit records; provides read-only access to audit information; 只读\n" +
		"// _provides read-only access to audit records; provides read-only access to audit records2; provides read-only access to audit records后\n"
	const auditOrderLookalikes = "var label = \"sort audit records by timestamp; synchronize system clocks; original arrival order of audit records\"\n" +
		"// time ordering of audit records; original time ordering of audit events; original_time_ordering_of_audit_records\n" +
		"// _original time ordering of audit records; original time ordering of audit records2; original time ordering of audit records后; original  time ordering of audit records\n"
	const auditContentLookalikes = "var status = \"the original content of audit records is not preserved; integrity check passed; immutable audit records\"\n" +
		"// _the original content of audit records is preserved; the original content of audit records is preserved2\n" +
		"// the original content of audit records is preserved后; the original content of  audit records is preserved\n"
	const onDemandReportLookalikes = "var label = \"generate scheduled audit reports; collect audit records on demand; generated on-demand audit reports\"\n" +
		"// regenerate on-demand audit reports; generate on-demand audit reports2; generate on-demand audit reports后\n" +
		"// generate_on_demand_audit_reports; generate on-demand audit report; generate on-demand  audit reports\n"
	const auditFilterLookalikes = "var label = \"filter audit records by event type; collect audit records by event type; delete audit records by event type\"\n" +
		"// search and filter audit records by event ID; search and filter audit records before storage; search_and_filter_audit_records_by_event_type\n" +
		"// _search and filter audit records by event type; search and filter audit records by event type2; search and filter audit records by event type后\n"
	const auditTimeRangeLookalikes = "var label = \"collect audit records within a time range; retain audit records for a time range; search query audit records by time range\"\n" +
		"// find_all_audit_records_within_a_time_range; find all audit records within a date range; find all audit records within  a time range\n" +
		"// _find all audit records within a time range; find all audit records within a time range2; find all audit records within a time range后\n"
	const auditActorLookalikes = "var actor = \"search audit records by target account; search audit records by reporting service; collect audit records by actor\"\n" +
		"// research audit records by actor; search audit records by actors; search audit records by actor_id; search audit records by actor后\n" +
		"// search_audit_records_by_actor; searchAuditRecordsByActor; search audit  records by actor\n"
	const auditTargetLookalikes = "var targetResources = \"get audit logs by actor; get audit logs by resource owner; get audit logs for a given resource\"\n" +
		"// forget audit logs containing a given target resource; get audit logs containing a given target resources; get audit logs containing a given target resource_id\n" +
		"// get audit logs containing a given target resource后; get_audit_logs_containing_a_given_target_resource; getAuditLogsByTargetResource\n"
	const correlationLookalikes = "var correlationId = \"Get audit logs by event ID; Get audit logs by resource ID; Get audit logs by user ID\"\n" +
		"// _Get audit logs by correlation ID; Get audit logs by correlation IDs; Get audit logs by correlation ID后\n" +
		"// get audit logs by correlation ID; Get audit logs by correlation id; GetAuditLogsByCorrelationID; correlation_id\n"
	const auditPageLookalikes = "var NextToken = \"get all pages of audit logs; get the first page of audit logs; get the next audit event by EventId\"\n" +
		"// get the next page token; get_the_next_page_of_audit_logs; forget the next page of audit logs; get the next page of audit logs_v2\n" +
		"// get the next page of audit logs后; get the next  page of audit logs\n"
	const auditPageLimitLookalikes = "var pageSize = \"number of audit logs returned on this page; total number of matching audit logs; default number of audit logs per page\"\n" +
		"// maximum number of audit logs per collection cycle; maximum number of audit logs to return per request\n" +
		"// _maximum number of audit logs to return per page; maximum number of audit logs to return per page2; maximum number of audit logs to return per page后\n"
	const auditCSVLookalikes = "var defaultFormat = \"CSV\" // export all audit logs to a CSV file; exported the results of an audit log search to a CSV file\n" +
		"// export the results of an audit log search to a JSON file; export the results of an audit log search to a csv file\n" +
		"// reexport the results of an audit log search to a CSV file; export the results of an audit log search to a CSV file_name; export the results of an audit log search to a CSV file后\n"
	const savedAuditFilterLookalikes = "var label = \"save audit log search results; schedule an audit log search; update a saved audit log filter\"\n" +
		"// Save Filter; saved_filters; recreate a saved audit log filter; create a saved audit log filter_id\n" +
		"// create a saved audit log filters; create a saved audit log filter后; create a saved audit  log filter\n"
	const clearAuditFilterLookalikes = "var label = \"Clear filters; clear all audit log filters; delete a saved audit log filter; delete audit log records\"\n" +
		"// _clear all audit log filters and reload; clear all audit log filters and reload2; clear all audit log filters and reload后\n" +
		"// clearAuditLogFilters; clear_all_audit_log_filters_and_reload; reload previous audit log results\n"
	const auditDetailClause = "when you expand an audit log entry, you see exactly what changed"
	const auditDetailLookalikes = "var label = \"Expand; expand an audit log entry; edit an audit log entry; export an audit log entry\"\n" +
		"// when you expand an audit log entry, you see only a summary; expand all audit log entries\n" +
		"// _" + auditDetailClause + "; " + auditDetailClause + "2; " + auditDetailClause + "后\n"
	const auditValuesClause = "the audit log shows the old and new values"
	const auditValuesLookalikes = "var oldValue, newValue = 1, 2 // old and new values; restore the old value; roll back the change\n" +
		"// the audit log shows only the new value; the audit log does not show the old and new values\n" +
		"// _" + auditValuesClause + "; " + auditValuesClause + "2; " + auditValuesClause + "后\n"
	const copyAuditIDLookalikes = "var CopyAuditEventID = record.id // Copy ID; Copy Audit Record; Copy Audit Event Hash; Export Audit Logs\n" +
		"// Copy audit event ID; Copy Audit Event IDs; _Copy Audit Event ID; Copy Audit Event ID2; Copy Audit Event ID后; event_id\n"
	const auditTimeSubject = "timestamps in the audit logs table and detail view"
	const auditLocalDisplay = auditTimeSubject + " are displayed in your local time zone"
	const auditLocalLookalikes = "var timeZone = \"local time zone\" // change the system time zone; synchronize the system clock\n" +
		"// " + auditTimeSubject + " are stored in your local time zone; " + auditTimeSubject + " are displayed in UTC\n" +
		"// " + auditTimeSubject + " are not displayed in your local time zone; _" + auditLocalDisplay + "; " + auditLocalDisplay + "2; " + auditLocalDisplay + "后\n"
	const occurrenceLabel = "The time the event described by the log entry occurred"
	const receiptLabel = "The time the log entry was received by Logging"
	const auditTimingLookalikes = "var receiveTimestamp = \"receipt time; time written to disk; requestReceivedTimestamp\"\n" +
		"// The time the request was received by Logging; The time the log entry was received by logging; The time the log entry was written to disk\n" +
		"// _" + occurrenceLabel + "; " + occurrenceLabel + "2; _" + receiptLabel + "; " + receiptLabel + "后\n"
	const auditEmptyState = "No audit logs match your filters"
	const auditEmptyLookalikes = "var label = \"No audit logs yet; No audit log entries are accessible; Audit log query failed\"\n" +
		"// No audit logs match all filters; No audit logs match your filter; " + auditEmptyState + "2; " + auditEmptyState + "后\n" +
		"// _" + auditEmptyState + "; No audit logs match  your filters; permission denied; next page pending\n"
	const auditLoadingLookalikes = "var label = \"No audit log entries found; Error loading audit log entries; Loading...\"\n" +
		"// _Loading audit log entries; Loading audit log entries2; Loading audit log entries后; LoadingAuditLogEntries\n"
	const auditRetryLookalikes = "var retryLoad = \"Retry; Retrying audit log query; Audit log retry succeeded; Failed to load audit logs\"\n" +
		"// Automatically retry loading audit logs; RetryLoadingAuditLogs; _Retry loading audit logs; Retry loading audit logs2; Retry loading audit logs后\n"
	const auditTimeoutLookalikes = "var status = \"Audit log query cancelled by user; No matching audit records; failed to query audit logs\"\n" +
		"// audit log query timeout; context.DeadlineExceeded; _audit log query timed out; audit log query timed out2; audit log query timed out后\n"
	const auditCancelLookalikes = "var label = \"Cancel; cancellation requested; CloudTrail Lake query cancelled; CloudTrail Lake query timed out\"\n" +
		"// CancelQuery; CANCELLED; TIMED_OUT; cancel a running CloudTrail Lake query; Cancel a scheduled CloudTrail Lake query\n" +
		"// _Cancel a running CloudTrail Lake query; Cancel a running CloudTrail Lake query2; Cancel a running CloudTrail Lake query后\n"
	const lakeQueryStatusLabel = "Get the status of a CloudTrail Lake query"
	const lakeQueryStatusLookalikes = "var label = \"Get the results of a CloudTrail Lake query; Get an audit event by ID; get_query_status; get_query_results\"\n" +
		"// get the status of a CloudTrail Lake query; _" + lakeQueryStatusLabel + "; " + lakeQueryStatusLabel + "2; " + lakeQueryStatusLabel + "后\n"
	const deliveryCondition = "If the query results were delivered to an S3 bucket, "
	const deliveryPhrase = deliveryCondition + "the response also provides the S3 URI and the delivery status"
	const deliveryLookalikes = "var delivery_status = \"The delivery status of the query results\" // QueryStatus: FINISHED; DeliveryStatus: FAILED; caller read: unknown\n" +
		"// If the query results were not delivered to an S3 bucket, the response also provides the S3 URI and the delivery status\n" +
		"// _" + deliveryPhrase + "; " + deliveryPhrase + "2; " + deliveryPhrase + "后\n"
	const scannedEventsLabel = "The number of events that the query scanned in the event data store"
	const scannedEventsLookalikes = "var EventsScanned = \"The number of events that matched a query; The number of results returned\"\n" +
		"// EventsMatched; QueryResultRows; the number of events that the query scanned in the event data store\n" +
		"// _" + scannedEventsLabel + "; " + scannedEventsLabel + "s; " + scannedEventsLabel + "后\n"
	const scannedBytesLabel = "The total bytes that the query scanned in the event data store"
	const scannedBytesLookalikes = "var BytesScanned = 4096 // EventsScanned; bytes billed; bytes written to the result file\n" +
		"// The total bytes that the query returned; the total bytes that the query scanned in the event data store\n" +
		"// _" + scannedBytesLabel + "; " + scannedBytesLabel + "s; " + scannedBytesLabel + "后\n"
	const queryRuntimeLabel = "The query's run time, in milliseconds"
	const queryRuntimeLookalikes = "var ExecutionTimeInMillis = 18 // The client's wait time, in milliseconds; The query's timeout, in milliseconds\n" +
		"// The query's run time, in seconds; the query's run time, in milliseconds; _" + queryRuntimeLabel + "; " + queryRuntimeLabel + "2; " + queryRuntimeLabel + "后\n"
	const queryCreationLabel = "Lists matching query results, and shows query ID, status, and creation time of each query"
	const queryCreationLookalikes = "var CreationTime = \"The creation time of the query\" // execution start time; audit event time; saved query template creation time\n" +
		"// _" + queryCreationLabel + "; " + queryCreationLabel + "2; " + queryCreationLabel + "后\n"
	const queryStatusFilterLabel = "The status of queries that you want to return in results"
	const queryStatusFilterLookalikes = "var QueryStatus = \"FINISHED\" // The status of audit events that you want to return in results; The status of result rows that you want to return in results\n" +
		"// the status of queries that you want to return in results; _" + queryStatusFilterLabel + "; " + queryStatusFilterLabel + "2; " + queryStatusFilterLabel + "后\n"
	const queryListRangeClause = "to bound a ListQueries request, and limit its results to only those queries run within a specified time period"
	const queryListRangeLookalikes = "var label = \"list audit events within a specified time period; audit event retention period\"\n" +
		"// to bound a ListQueries request, and limit its results to only those events recorded within a specified time period\n" +
		"// _" + queryListRangeClause + "; " + queryListRangeClause + "2; " + queryListRangeClause + "后\n"
	const queryListStoreLabel = "The ARN (or the ID suffix of the ARN) of an event data store on which queries were run"
	const queryListStoreLookalikes = "var label = \"The ID of the query; The URI for the S3 bucket where CloudTrail delivered query results\"\n" +
		"// EventDataStore; QueryId; DeliveryS3Uri; The ARN of an event data store; the ARN (or the ID suffix of the ARN) of an event data store on which queries were run\n" +
		"// _" + queryListStoreLabel + "; " + queryListStoreLabel + "2; " + queryListStoreLabel + "后\n"
	const queryListPageLabel = "NextPage retrieves the next ListQueries page"
	const queryListPageLookalikes = "var help = \"NextPage retrieves the next GetQueryResults page\" // A token you can use to get the next page of results.\n" +
		"// nextPage retrieves the next ListQueries page; _" + queryListPageLabel + "; " + queryListPageLabel + "2; " + queryListPageLabel + "后\n"
	const queryListLimitLabel = "The maximum number of queries to show on a page"
	const queryListLimitLookalikes = "var MaxResults = 5 // The maximum number of query results to display on a single page; total number of query executions\n" +
		"// MaxQueryResults; QueryResultRows; the maximum number of queries to show on a page\n" +
		"// _" + queryListLimitLabel + "; " + queryListLimitLabel + "2; " + queryListLimitLabel + "后\n"
	const queryErrorLabel = "The error message returned if a query failed"
	const queryErrorLookalikes = "var label = \"Error getting query status\" // DescribeQuery: request failed; QueryStatus: unknown; no metadata returned\n" +
		"// The error message returned if a status request failed; the error message returned if a query failed\n" +
		"// _" + queryErrorLabel + "; " + queryErrorLabel + "2; " + queryErrorLabel + "后\n"
	const auditPageTokenLookalikes = "var NextToken = \"Invalid pagination token; Query ID not found; audit query results have expired; no more pages\"\n" +
		"// InvalidNextTokenException; QueryIdNotFoundException; InvalidAuditLogPageToken; no next token\n" +
		"// _invalid audit log page token; invalid audit log page tokens; invalid audit log page token2; invalid audit log page token后; invalid audit  log page token\n"
	const eventReadOnlyLabel = "Information about whether the event is a write event or a read event"
	const eventReadOnlyLookalikes = "var readOnly = false // ReadOnly; read-only access to audit records; read succeeded; write failed\n" +
		"// information about whether the event is a write event or a read event; _" + eventReadOnlyLabel + "; " + eventReadOnlyLabel + "2; " + eventReadOnlyLabel + "后\n"
	const eventNameLookupError = "failed to lookup CloudTrail events by event name"
	const eventNameLookalikes = "var label = \"failed to lookup CloudTrail events by event ID; failed to lookup CloudTrail events by user name\"\n" +
		"// CloudTrail Events by Event Name; LookupAttributeKeyEventName; Failed to lookup CloudTrail events by event name\n" +
		"// _" + eventNameLookupError + "; " + eventNameLookupError + "s; " + eventNameLookupError + "后\n"
	const eventIDLookup = "Retrieves a specific audit event by its ID"
	const eventIDLookalikes = "var eventId = \"Retrieves a specific audit event by its API operation name; Retrieves a specific audit event by its correlation ID\"\n" +
		"// Retrieved a specific audit event by its ID; retrieves a specific audit event by its ID; getEvent(eventId)\n" +
		"// _" + eventIDLookup + "; " + eventIDLookup + "s; " + eventIDLookup + "2; " + eventIDLookup + "后\n"
	const auditResourceTypeFilter = "Parameter for filtering audit logs by resource type"
	const auditResourceTypeLookalikes = "var resourceType = \"Parameter for filtering audit logs by resource name; Parameter for filtering audit logs by event type\"\n" +
		"// resource_type; Parameter for sorting audit logs by resource type; parameter for filtering audit logs by resource type\n" +
		"// _" + auditResourceTypeFilter + "; " + auditResourceTypeFilter + "s; " + auditResourceTypeFilter + "2; " + auditResourceTypeFilter + "后\n"
	const auditSourceIPFilter = "Get audit logs by source IP"
	const auditSourceIPLookalikes = "var label = \"Get audit logs by destination IP; Get audit logs by authenticated principal\"\n" +
		"// GetAuditLogsBySourceIP; get audit logs by source IP; Got audit logs by source IP\n" +
		"// _" + auditSourceIPFilter + "; " + auditSourceIPFilter + "s; " + auditSourceIPFilter + "2; " + auditSourceIPFilter + "后\n"
	const auditUserAgentFilter = "Filter OpenShift API server audit logs by user agent"
	const auditUserAgentLookalikes = "var userAgent = \"Filter OpenShift API server audit logs by authenticated user; Filter OpenShift API server audit logs by client ID\"\n" +
		"// Filter Kubernetes API server audit logs by a certain API version and only output the user agent; User-Agent; filter OpenShift API server audit logs by user agent\n" +
		"// _" + auditUserAgentFilter + "; " + auditUserAgentFilter + "s; " + auditUserAgentFilter + "2; " + auditUserAgentFilter + "后\n"
	const auditHTTPStatusFilter = "Find audit logs by HTTP status code"
	const auditHTTPStatusLookalikes = "var statusCode = \"Find audit logs by status code; Find audit logs by operation outcome; HTTP status code of the audit query\"\n" +
		"// findByStatusCode; Found audit logs by HTTP status code; find audit logs by HTTP status code\n" +
		"// _" + auditHTTPStatusFilter + "; " + auditHTTPStatusFilter + "s; " + auditHTTPStatusFilter + "2; " + auditHTTPStatusFilter + "后\n"
	const auditHTTPMethodFilter = "Find audit logs by HTTP method"
	const auditHTTPMethodLookalikes = "var httpMethod = \"Find audit logs by method; Find audit logs by operation name; HTTP method for fetching audit logs\"\n" +
		"// findByHttpMethodOrderByTimestampDesc; Found audit logs by HTTP method; find audit logs by HTTP method\n" +
		"// _" + auditHTTPMethodFilter + "; " + auditHTTPMethodFilter + "s; " + auditHTTPMethodFilter + "2; " + auditHTTPMethodFilter + "后\n"
	const auditRequestPathFilter = "Find audit logs by request path"
	const auditRequestPathContext = ". Filter path: /books; request target: /books?limit=2; log query: GET /audit; matches: 0\n"
	const auditRequestPathLookalikes = "var label = \"Find audit logs by request target; Find audit logs by query string; Path to retrieve audit logs\"\n" +
		"// findByPath; Find audit logs by path; Found audit logs by request path; find audit logs by request path\n" +
		"// _" + auditRequestPathFilter + "; " + auditRequestPathFilter + "s; " + auditRequestPathFilter + "2; " + auditRequestPathFilter + "后\n"
	const auditDecisionFilter = "auth_decision: Filter by authorization decision"
	const auditDecisionContext = "; recorded decision: ALLOW; response.status_code: 503; reader access: unknown\n"
	const auditDecisionLookalikes = "var auth_decision = \"Filter by authorization decision; auth_decision: Filter by authentication decision; auth_decision: Filter by request outcome\"\n" +
		"var entry = `{\"response\":{\"status_code\":403}}` // decision missing; matches: 0; DENY and NOT_REQUIRED are explicit values\n" +
		"// auth_decision: filter by authorization decision; _" + auditDecisionFilter + "; " + auditDecisionFilter + "s; " + auditDecisionFilter + "后\n"
	const minRequestDurationFilter = "Filter by minimum request duration in milliseconds"
	const maxRequestDurationFilter = "Filter by maximum request duration in milliseconds"
	const requestDurationContext = "; durationMs: 500; query elapsed: 7 ms; timeout: 1000 ms; timedOut: false\n"
	const requestDurationLookalikes = "var label = \"Filter by minimum query duration in milliseconds; Filter by maximum request timeout in milliseconds; Filter by maximum request duration in seconds\"\n" +
		"// minDurationMs; maxDurationMs; milliseconds; filter by minimum request duration in milliseconds\n" +
		"// _" + minRequestDurationFilter + "; " + maxRequestDurationFilter + "2; " + minRequestDurationFilter + "后\n"
	const auditTLSVersionFilter = "This filters all TLS connections with 1.1 and below"
	const auditTLSVersionContext = ". tlsDetails.tlsVersion: TLSv1.1; SupportedVersions: [TLSv1.1, TLSv1.3]; MinVersion: TLSv1.0\n"
	const auditTLSVersionLookalikes = "var entry = `{\"tlsDetails\":{}}` // negotiated version: unknown; SupportedVersions: [TLSv1.3]; MinVersion: TLSv1.2\n" +
		"// This filters all TLS connections with 1.1 and above; This filters all TLS connections with 1.0 and below\n" +
		"// this filters all TLS connections with 1.1 and below; _" + auditTLSVersionFilter + "; " + auditTLSVersionFilter + "2; " + auditTLSVersionFilter + "后\n"
	cases := []struct {
		name      string
		opt       Options
		in, want  string
		wantTerms int
		wantInCom int // how many of the hits were inside a comment
	}{
		{
			name: "code and comment both hit, comment then stripped",
			opt:  Options{Glossary: tb},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 威胁检测 = 1\n", wantTerms: 2, wantInCom: 1,
		},
		{
			// "pentest" -> "threat_detection" grows the buffer by 9 bytes
			// before both comments. Stale spans would cut into the code.
			// The underscore is the known cost of toIdent: the glossary
			// cannot tell a string literal from an identifier.
			name: "length change before comments re-lexes spans",
			opt:  Options{Glossary: tb},
			in:   "s := \"pentest\" // a\nt := 1 // b\n",
			want: "s := \"threat_detection\"\nt := 1\n", wantTerms: 1,
		},
		{
			name: "comments only leaves code alone",
			opt:  Options{Glossary: tb, TermsInComments: true, KeepComments: true},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 渗透 = 1 // 威胁检测\n", wantTerms: 1, wantInCom: 1,
		},
		{
			name: "keep comments rewrites without stripping",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   "// 漏洞利用\nvar x = 1\n",
			want: "// 风险验证\nvar x = 1\n", wantTerms: 1, wantInCom: 1,
		},
		{
			name: "qualified IDS terms in code and kept comments",
			opt:  Options{Glossary: tb, KeepComments: true},
			in: "var 主机型入侵检测系统 = \"host-based IDS\" // 网络型入侵检测系统, network-based IDS\n" +
				"var 网络型入侵检测系统 = \"network-based IDS\" // 主机型入侵检测系统, host-based IDS\n",
			want: "var 基于主机的入侵检测系统 = \"host_based_intrusion_detection_system\" // 基于网络的入侵检测系统, network intrusion detection system\n" +
				"var 基于网络的入侵检测系统 = \"network_intrusion_detection_system\" // 基于主机的入侵检测系统, host-based intrusion detection system\n",
			wantTerms: 8, wantInCom: 4,
		},
		{
			name: "IDS lookalikes remain unchanged",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   idsLookalikes, want: idsLookalikes,
		},
		{
			name:      "audit record retention period in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"audit record retention period\" // Audit record retention period (AU-11)\n",
			want:      "var label = \"retention_period_for_audit_records\" // Retention period for audit records (AU-11)\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit retention boundaries and different concepts",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   retentionLookalikes, want: retentionLookalikes,
		},
		{
			name:      "audit timestamps in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"timestamps for audit records\" // Timestamps for audit records: 2025-01-02T03:04:05Z\n",
			want:      "var label = \"time_stamps_for_audit_records\" // Time stamps for audit records: 2025-01-02T03:04:05Z\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit timestamp boundaries and unrelated timestamps",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   timestampLookalikes, want: timestampLookalikes,
		},
		{
			name:      "audit event result in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"audit event result\" // Audit event result: failure; audit log write result: success\n",
			want:      "var label = \"outcome_of_the_audited_event\" // Outcome of the audited event: failure; audit log write result: success\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit event result boundaries and unrelated outcomes",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   eventResultLookalikes, want: eventResultLookalikes,
		},
		{
			name:      "audit storage allocation in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"allocate audit log storage capacity\" // Allocate audit log storage capacity: 8 GiB; free: 2 GiB; retention: 30 days\n",
			want:      "var label = \"allocate_storage_capacity_for_audit_logs\" // Allocate storage capacity for audit logs: 8 GiB; free: 2 GiB; retention: 30 days\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit storage allocation boundaries and other measurements",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditCapacityLookalikes, want: auditCapacityLookalikes,
		},
		{
			name:      "audit disk near full warning in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var msg = \"the disk with the audit log is almost full\" // The disk with the audit log is almost full (threshold from policy)\n",
			want:      "var msg = \"the_disk_holding_the_audit_log_is_nearly_full\" // The disk holding the audit log is nearly full (threshold from policy)\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit disk warning boundaries and distinct failures",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditDiskLookalikes, want: auditDiskLookalikes,
		},
		{
			name:      "audit log rotation in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"rotation of audit log files\" // Rotation of audit log files: disabled; retention: 14 days\n",
			want:      "var label = \"audit_log_file_rotation\" // Audit log file rotation: disabled; retention: 14 days\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit log rotation boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditRotationLookalikes, want: auditRotationLookalikes,
		},
		{
			name:      "rotated audit log compression in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"compression of rotated audit log files\" // Compression of rotated audit log files: disabled; codec: inherited; retention: unchanged\n",
			want:      "var label = \"compression_of_audit_log_files_that_have_been_rotated\" // Compression of audit log files that have been rotated: disabled; codec: inherited; retention: unchanged\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "rotated audit log compression boundaries and other operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditCompressionLookalikes, want: auditCompressionLookalikes,
		},
		{
			name:      "audit file integrity action in code and a negated comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"verify audit log file integrity\" // Could not verify audit log file integrity; result: unknown\n",
			want:      "var action = \"verify_the_integrity_of_the_audit_log_file\" // Could not verify the integrity of the audit log file; result: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit file integrity boundaries and result labels",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditIntegrityLookalikes, want: auditIntegrityLookalikes,
		},
		{
			name:      "audit read-only access in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var scope = \"provides read-only access to audit records\" // This API provides read-only access to audit records; records read: 0\n",
			want:      "var scope = \"allows_read_only_access_to_audit_records\" // This API allows read-only access to audit records; records read: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit read-only access boundaries and unrelated claims",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditReadOnlyLookalikes, want: auditReadOnlyLookalikes,
		},
		{
			name:      "audit record time order in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"original time ordering of audit records\" // Original time ordering of audit records was not preserved: 12:02, 12:01\n",
			want:      "var label = \"original_chronological_order_of_audit_records\" // Original chronological order of audit records was not preserved: 12:02, 12:01\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit record time order boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditOrderLookalikes, want: auditOrderLookalikes,
		},
		{
			name:      "audit record content preservation in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var objective = \"the original content of audit records is preserved\" // No evidence that the original content of audit records is preserved.\n",
			want:      "var objective = \"the_original_content_of_audit_records_is_kept_intact\" // No evidence that the original content of audit records is kept intact.\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit record content preservation boundaries and other claims",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditContentLookalikes, want: auditContentLookalikes,
		},
		{
			name:      "on-demand audit reports in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"generate on-demand audit reports\" // Cannot generate on-demand audit reports; reports generated: 0\n",
			want:      "var action = \"generate_audit_reports_on_demand\" // Cannot generate audit reports on demand; reports generated: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "on-demand audit reports boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   onDemandReportLookalikes, want: onDemandReportLookalikes,
		},
		{
			name:      "audit record filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"search and filter audit records by event type\" // Cannot search and filter audit records by event type: login; 0 shown, 12 stored\n",
			want:      "var action = \"search_and_filter_audit_log_entries_by_event_type\" // Cannot search and filter audit log entries by event type: login; 0 shown, 12 stored\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit record filter boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditFilterLookalikes, want: auditFilterLookalikes,
		},
		{
			name:      "audit time range query in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"find all audit records within a time range\" // Cannot find all audit records within a time range: [09:00Z, 11:00Z); collection: continuous; retention: 30 days\n",
			want:      "var action = \"find_all_audit_records_in_a_specified_time_range\" // Cannot find all audit records in a specified time range: [09:00Z, 11:00Z); collection: continuous; retention: 30 days\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit time range query boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditTimeRangeLookalikes, want: auditTimeRangeLookalikes,
		},
		{
			name:      "audit actor query in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"search audit records by actor\" // Cannot search audit records by actor; keys: actor_id, target_id, reporter_id\n",
			want:      "var action = \"query_audit_records_by_actor\" // Cannot query audit records by actor; keys: actor_id, target_id, reporter_id\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit actor query boundaries and other roles",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditActorLookalikes, want: auditActorLookalikes,
		},
		{
			name:      "audit target resource query in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"get audit logs containing a given target resource\" // Cannot get audit logs containing a given target resource; fields: targetResources, initiatedBy, owner_id\n",
			want:      "var label = \"retrieve_audit_logs_containing_a_given_target_resource\" // Cannot retrieve audit logs containing a given target resource; fields: targetResources, initiatedBy, owner_id\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit target resource query boundaries and other roles",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditTargetLookalikes, want: auditTargetLookalikes,
		},
		{
			name:      "audit correlation ID query in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Get audit logs by correlation ID\" // Get audit logs by correlation ID; same correlationId, different actors; eventId and resourceId unchanged\n",
			want:      "var label = \"Retrieve_audit_logs_by_correlation_ID\" // Retrieve audit logs by correlation ID; same correlationId, different actors; eventId and resourceId unchanged\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit correlation ID query boundaries and unrelated IDs",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   correlationLookalikes, want: correlationLookalikes,
		},
		{
			name:      "audit next page in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Get the next page of audit logs\" // Cannot get the next page of audit logs; NextToken present; EventId unchanged; remaining: unknown\n",
			want:      "var label = \"Retrieve_the_next_page_of_audit_logs\" // Cannot retrieve the next page of audit logs; NextToken present; EventId unchanged; remaining: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit next page boundaries and distinct lookups",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditPageLookalikes, want: auditPageLookalikes,
		},
		{
			name:      "audit page limit in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Maximum number of audit logs to return per page\" // maximum number of audit logs to return per page: 8; returned: 3; total matches: 19; default: unspecified\n",
			want:      "var label = \"Maximum_number_of_audit_logs_to_return_on_any_one_page\" // maximum number of audit logs to return on any one page: 8; returned: 3; total matches: 19; default: unspecified\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit page limit boundaries and distinct counts",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditPageLimitLookalikes, want: auditPageLimitLookalikes,
		},
		{
			name:      "audit search CSV export in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"export the results of an audit log search to a CSV file\" // Cannot export the results of an audit log search to a CSV file; matched: 3 of 19; rows written: 0; default: NDJSON\n",
			want:      "var action = \"export_audit_log_search_results_to_a_CSV_file\" // Cannot export audit log search results to a CSV file; matched: 3 of 19; rows written: 0; default: NDJSON\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit search CSV export boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditCSVLookalikes, want: auditCSVLookalikes,
		},
		{
			name:      "saved audit log filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"Create a saved audit log filter\" // Cannot create a saved audit log filter; results saved: 0; schedule: off\n",
			want:      "var action = \"Save_a_new_audit_log_filter\" // Cannot save a new audit log filter; results saved: 0; schedule: off\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "saved audit log filter boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   savedAuditFilterLookalikes, want: savedAuditFilterLookalikes,
		},
		{
			name:      "clear audit log filters in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"Clear all audit log filters and reload\" // Cannot clear all audit log filters and reload; saved presets: 2; stored records: 9\n",
			want:      "var action = \"Clear_every_audit_log_filter_and_reload\" // Cannot clear every audit log filter and reload; saved presets: 2; stored records: 9\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "clear audit log filters boundaries and distinct deletions",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   clearAuditFilterLookalikes, want: clearAuditFilterLookalikes,
		},
		{
			name:      "audit entry expansion in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + auditDetailClause + "\" // When you expand an audit log entry, you see exactly what changed; edits: 0; exports: 0\n",
			want:      "var help = \"expanding_an_audit_log_entry_shows_you_exactly_what_changed\" // Expanding an audit log entry shows you exactly what changed; edits: 0; exports: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit entry expansion boundaries and other actions",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditDetailLookalikes, want: auditDetailLookalikes,
		},
		{
			name:      "audit change values in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + auditValuesClause + "\" // Check whether the audit log shows the old and new values; oldValue: not recorded; newValue: 0; rollbacks: 0\n",
			want:      "var help = \"the_audit_log_shows_the_before_and_after_values\" // Check whether the audit log shows the before and after values; oldValue: not recorded; newValue: 0; rollbacks: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit change values boundaries and rollback wording",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditValuesLookalikes, want: auditValuesLookalikes,
		},
		{
			name:      "copy audit event ID in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Copy Audit Event ID\" // Disabled: Copy Audit Event ID; source: record.id; copied: false; exports: 0\n",
			want:      "var label = \"Copy_Audit_Event_Identifier\" // Disabled: Copy Audit Event Identifier; source: record.id; copied: false; exports: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "copy audit event ID boundaries and distinct actions",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   copyAuditIDLookalikes, want: copyAuditIDLookalikes,
		},
		{
			name:      "audit local time display in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditLocalDisplay + "\" // Timestamps in the audit logs table and detail view are displayed in your local time zone; raw: 2025-01-02T03:04:05Z\n",
			want:      "var label = \"timestamps_in_the_audit_logs_table_and_detail_view_are_shown_in_your_local_time_zone\" // Timestamps in the audit logs table and detail view are shown in your local time zone; raw: 2025-01-02T03:04:05Z\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit local time display boundaries and other clock settings",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditLocalLookalikes, want: auditLocalLookalikes,
		},
		{
			name: "audit event and receipt times in code and kept comments",
			opt:  Options{Glossary: tb, KeepComments: true},
			in: "var timestamp = \"" + occurrenceLabel + "\" // " + occurrenceLabel + ": 2025-01-02T03:04:05Z\n" +
				"var receive_timestamp = \"" + receiptLabel + "\" // " + receiptLabel + ": 2025-01-02T03:04:09Z; disk write: unknown\n",
			want: "var timestamp = \"When_the_event_described_by_the_log_entry_occurred\" // When the event described by the log entry occurred: 2025-01-02T03:04:05Z\n" +
				"var receive_timestamp = \"When_the_log_entry_was_received_by_Logging\" // When the log entry was received by Logging: 2025-01-02T03:04:09Z; disk write: unknown\n",
			wantTerms: 4, wantInCom: 2,
		},
		{
			name: "audit event and receipt times boundaries and other stages",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditTimingLookalikes, want: auditTimingLookalikes,
		},
		{
			name:      "audit filtered empty state in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var title = \"" + auditEmptyState + "\" // Result: no audit logs match your filters! Stored records: 42\n",
			want:      "var title = \"No_audit_log_entries_match_your_filters\" // Result: no audit log entries match your filters! Stored records: 42\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit filtered empty state boundaries and other states",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditEmptyLookalikes, want: auditEmptyLookalikes,
		},
		{
			name:      "audit query loading in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Loading audit log entries\" // Loading audit log entries... total: unknown\n",
			want:      "var label = \"Loading_audit_log_records\" // Loading audit log records... total: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query loading boundaries and other states",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditLoadingLookalikes, want: auditLoadingLookalikes,
		},
		{
			name:      "audit manual retry in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Retry loading audit logs\" // Button: Retry loading audit logs; last load: failed; retries started: 0; outcome: unknown\n",
			want:      "var label = \"Try_loading_audit_logs_again\" // Button: Try loading audit logs again; last load: failed; retries started: 0; outcome: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit manual retry boundaries and other states",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditRetryLookalikes, want: auditRetryLookalikes,
		},
		{
			name:      "audit query timeout in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var status = \"audit log query timed out\" // Audit log query timed out; cancelled by user: false; server status: unknown; matches: unknown\n",
			want:      "var status = \"audit_log_query_did_not_return_in_time\" // Audit log query did not return in time; cancelled by user: false; server status: unknown; matches: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query timeout boundaries and distinct outcomes",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditTimeoutLookalikes, want: auditTimeoutLookalikes,
		},
		{
			name:      "audit query cancellation request in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"Cancel a running CloudTrail Lake query\" // Action: Cancel a running CloudTrail Lake query; request sent: true; confirmation: pending; stopped: unknown\n",
			want:      "var label = \"Request_cancellation_of_a_running_CloudTrail_Lake_query\" // Action: Request cancellation of a running CloudTrail Lake query; request sent: true; confirmation: pending; stopped: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query cancellation request boundaries and distinct states",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditCancelLookalikes, want: auditCancelLookalikes,
		},
		{
			name:      "audit query status lookup in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var action = \"" + lakeQueryStatusLabel + "\" // Get the status of a CloudTrail Lake query; lookup: QueryId, not EventId; QueryStatus: RUNNING; rows: not fetched\n",
			want:      "var action = \"Check_the_status_of_a_CloudTrail_Lake_query\" // Check the status of a CloudTrail Lake query; lookup: QueryId, not EventId; QueryStatus: RUNNING; rows: not fetched\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query status lookup boundaries and other identifiers",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   lakeQueryStatusLookalikes, want: lakeQueryStatusLookalikes,
		},
		{
			name: "audit query result delivery in code and a kept comment",
			opt:  Options{Glossary: tb, KeepComments: true},
			in: "var help = \"" + deliveryPhrase + "\" // QueryStatus: FINISHED; DeliveryStatus: PENDING\n" +
				"// " + deliveryPhrase + ". DeliveryStatus: SUCCESS; caller read: unknown\n",
			want: "var help = \"If_the_query_results_were_delivered_to_an_S3_bucket,_the_response_also_includes_the_S3_URI_and_the_delivery_status\" // QueryStatus: FINISHED; DeliveryStatus: PENDING\n" +
				"// " + deliveryCondition + "the response also includes the S3 URI and the delivery status. DeliveryStatus: SUCCESS; caller read: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query result delivery boundaries and distinct states",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   deliveryLookalikes, want: deliveryLookalikes,
		},
		{
			name:      "audit scanned event count in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + scannedEventsLabel + "\" // " + scannedEventsLabel + ". EventsScanned: 24; EventsMatched: 5; rows on this page: 2\n",
			want:      "var label = \"How_many_events_the_query_scanned_in_the_event_data_store\" // How many events the query scanned in the event data store. EventsScanned: 24; EventsMatched: 5; rows on this page: 2\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit scanned event count boundaries and distinct counts",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   scannedEventsLookalikes, want: scannedEventsLookalikes,
		},
		{
			name:      "audit scanned bytes in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + scannedBytesLabel + "\" // " + scannedBytesLabel + ". BytesScanned: 4096; EventsScanned: 24; result file: 128 bytes; charge: unknown\n",
			want:      "var label = \"The_total_number_of_bytes_that_the_query_scanned_in_the_event_data_store\" // The total number of bytes that the query scanned in the event data store. BytesScanned: 4096; EventsScanned: 24; result file: 128 bytes; charge: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit scanned bytes boundaries and distinct measurements",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   scannedBytesLookalikes, want: scannedBytesLookalikes,
		},
		{
			name:      "audit query execution duration in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + queryRuntimeLabel + "\" // " + queryRuntimeLabel + ". ExecutionTimeInMillis: 18; client wait: 31 ms; timeout: 250 ms; status: not checked\n",
			want:      "var label = \"The_duration_of_query_execution_in_milliseconds\" // The duration of query execution in milliseconds. ExecutionTimeInMillis: 18; client wait: 31 ms; timeout: 250 ms; status: not checked\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query execution duration boundaries and other timings",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryRuntimeLookalikes, want: queryRuntimeLookalikes,
		},
		{
			name:      "audit query creation time in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + queryCreationLabel + "\" // " + queryCreationLabel + ". CreationTime: 10:00Z; QueryStatus: QUEUED; execution start: unknown; event time: 09:15Z\n",
			want:      "var label = \"Lists_matching_query_results_and_shows_the_ID,_status,_and_creation_time_for_each_query\" // Lists matching query results and shows the ID, status, and creation time for each query. CreationTime: 10:00Z; QueryStatus: QUEUED; execution start: unknown; event time: 09:15Z\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query creation time boundaries and other timestamps",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryCreationLookalikes, want: queryCreationLookalikes,
		},
		{
			name:      "audit query list status filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + queryStatusFilterLabel + "\" // " + queryStatusFilterLabel + ". QueryStatus filter: FINISHED; matching queries: 0; result rows: not fetched\n",
			want:      "var help = \"The_status_to_match_when_listing_queries\" // The status to match when listing queries. QueryStatus filter: FINISHED; matching queries: 0; result rows: not fetched\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query list status filter boundaries and other filters",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryStatusFilterLookalikes, want: queryStatusFilterLookalikes,
		},
		{
			name:      "audit query list time range in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var endHelp = \"Use with StartTime " + queryListRangeClause + ".\" // Use with EndTime " + queryListRangeClause + ". Runs: [10:00Z, 11:00Z); events: [08:00Z, 09:00Z); retention: unknown\n",
			want:      "var endHelp = \"Use with StartTime to_limit_ListQueries_results_to_queries_run_within_the_specified_time_range.\" // Use with EndTime to limit ListQueries results to queries run within the specified time range. Runs: [10:00Z, 11:00Z); events: [08:00Z, 09:00Z); retention: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query list time range boundaries and other windows",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryListRangeLookalikes, want: queryListRangeLookalikes,
		},
		{
			name:      "audit query list store in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + queryListStoreLabel + "\" // " + queryListStoreLabel + ". Scope: EventDataStore; QueryId: not supplied; DeliveryS3Uri: not requested\n",
			want:      "var help = \"The_ARN_or_its_ID_suffix_for_the_event_data_store_where_the_queries_ran\" // The ARN or its ID suffix for the event data store where the queries ran. Scope: EventDataStore; QueryId: not supplied; DeliveryS3Uri: not requested\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query list store boundaries and other identifiers",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryListStoreLookalikes, want: queryListStoreLookalikes,
		},
		{
			name:      "audit query list next page in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + queryListPageLabel + "\" // " + queryListPageLabel + ". ListQueriesOutput.NextToken: present; QueryId: unset; GetQueryResultsOutput.NextToken: unknown; rows: not fetched\n",
			want:      "var help = \"NextPage_retrieves_the_next_page_from_ListQueries\" // NextPage retrieves the next page from ListQueries. ListQueriesOutput.NextToken: present; QueryId: unset; GetQueryResultsOutput.NextToken: unknown; rows: not fetched\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query list next page boundaries and other paginators",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryListPageLookalikes, want: queryListPageLookalikes,
		},
		{
			name:      "audit query list page limit in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + queryListLimitLabel + "\" // " + queryListLimitLabel + ". MaxResults: 5; queries on this page: 2; total queries: 12; rows for one query: 17\n",
			want:      "var help = \"The_maximum_number_of_queries_to_list_per_page\" // The maximum number of queries to list per page. MaxResults: 5; queries on this page: 2; total queries: 12; rows for one query: 17\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query list page limit boundaries and distinct counts",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryListLimitLookalikes, want: queryListLimitLookalikes,
		},
		{
			name:      "audit query error message in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + queryErrorLabel + "\" // " + queryErrorLabel + ". DescribeQuery: OK; QueryStatus: FAILED; ErrorMessage: not provided\n",
			want:      "var help = \"The_error_message_returned_if_the_query_itself_failed\" // The error message returned if the query itself failed. DescribeQuery: OK; QueryStatus: FAILED; ErrorMessage: not provided\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit query error message boundaries and status request errors",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   queryErrorLookalikes, want: queryErrorLookalikes,
		},
		{
			name:      "invalid audit page token in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var msg = \"Invalid audit log page token.\" // Error: invalid audit log page token; QueryId: not checked; results expired: unknown; more pages: unknown\n",
			want:      "var msg = \"Invalid_audit_log_pagination_token.\" // Error: invalid audit log pagination token; QueryId: not checked; results expired: unknown; more pages: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "invalid audit page token boundaries and other outcomes",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditPageTokenLookalikes, want: auditPageTokenLookalikes,
		},
		{
			name:      "audit event read-only attribute in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + eventReadOnlyLabel + "\" // " + eventReadOnlyLabel + ". ReadOnly: true; operation result: failed; viewer permissions: unknown\n",
			want:      "var help = \"Whether_the_event_is_classified_as_a_write_or_a_read\" // Whether the event is classified as a write or a read. ReadOnly: true; operation result: failed; viewer permissions: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit event read-only attribute boundaries and unrelated permissions",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   eventReadOnlyLookalikes, want: eventReadOnlyLookalikes,
		},
		{
			name:      "audit event name lookup in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var msg = \"" + eventNameLookupError + ": %w\" // " + eventNameLookupError + "; EventName: ListTrails; EventId: unknown; Username: unknown; operation outcome: unknown\n",
			want:      "var msg = \"failed_to_look_up_CloudTrail_events_by_event_name: %w\" // failed to look up CloudTrail events by event name; EventName: ListTrails; EventId: unknown; Username: unknown; operation outcome: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit event name lookup boundaries and other filters",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   eventNameLookalikes, want: eventNameLookalikes,
		},
		{
			name:      "audit event ID lookup in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + eventIDLookup + "\" // " + eventIDLookup + ". eventId: evt-17; found: false; operation outcome: unknown\n",
			want:      "var label = \"Retrieves_a_specific_audit_event_by_its_identifier\" // Retrieves a specific audit event by its identifier. eventId: evt-17; found: false; operation outcome: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit event ID lookup boundaries and other keys",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   eventIDLookalikes, want: eventIDLookalikes,
		},
		{
			name:      "audit resource type filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditResourceTypeFilter + "\" // " + auditResourceTypeFilter + ". Multiple values can be provided. resource_type: group; matches: 0\n",
			want:      "var label = \"Parameter_to_filter_audit_logs_by_resource_type\" // Parameter to filter audit logs by resource type. Multiple values can be provided. resource_type: group; matches: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit resource type filter boundaries and other criteria",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditResourceTypeLookalikes, want: auditResourceTypeLookalikes,
		},
		{
			name:      "audit source IP filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditSourceIPFilter + "\" // " + auditSourceIPFilter + "; sourceIp: 192.0.2.10; destinationIp: 198.51.100.20; actor: unknown; matches: 0\n",
			want:      "var label = \"Retrieve_audit_logs_by_source_IP\" // Retrieve audit logs by source IP; sourceIp: 192.0.2.10; destinationIp: 198.51.100.20; actor: unknown; matches: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit source IP filter boundaries and unrelated lookups",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditSourceIPLookalikes, want: auditSourceIPLookalikes,
		},
		{
			name:      "audit user agent filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditUserAgentFilter + ":\" // " + auditUserAgentFilter + ": userAgent: ExampleClient/1.0; clientVerified: false; actor: unknown; matches: 0\n",
			want:      "var label = \"Filter_OpenShift_API_server_audit_log_entries_by_user_agent:\" // Filter OpenShift API server audit log entries by user agent: userAgent: ExampleClient/1.0; clientVerified: false; actor: unknown; matches: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit user agent filter boundaries and distinct operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditUserAgentLookalikes, want: auditUserAgentLookalikes,
		},
		{
			name:      "audit HTTP status filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditHTTPStatusFilter + "\" // " + auditHTTPStatusFilter + ": filter statusCode: 503; query HTTP: 200; matches: 0; operation outcome: unknown\n",
			want:      "var label = \"Find_audit_log_entries_by_HTTP_status_code\" // Find audit log entries by HTTP status code: filter statusCode: 503; query HTTP: 200; matches: 0; operation outcome: unknown\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit HTTP status filter boundaries and other statuses",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditHTTPStatusLookalikes, want: auditHTTPStatusLookalikes,
		},
		{
			name:      "audit HTTP method filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditHTTPMethodFilter + "\" // " + auditHTTPMethodFilter + ": httpMethod: PATCH; query: GET /audit; operation: UpdateProfile; matches: 0\n",
			want:      "var label = \"Find_audit_log_entries_by_HTTP_method\" // Find audit log entries by HTTP method: httpMethod: PATCH; query: GET /audit; operation: UpdateProfile; matches: 0\n",
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit HTTP method filter boundaries and other operations",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditHTTPMethodLookalikes, want: auditHTTPMethodLookalikes,
		},
		{
			name:      "audit request path filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var label = \"" + auditRequestPathFilter + ".\" // " + auditRequestPathFilter + auditRequestPathContext,
			want:      "var label = \"Find_audit_log_entries_by_request_path.\" // Find audit log entries by request path" + auditRequestPathContext,
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit request path filter boundaries and other lookup keys",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditRequestPathLookalikes, want: auditRequestPathLookalikes,
		},
		{
			name:      "audit authorization decision filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + auditDecisionFilter + "\" // " + auditDecisionFilter + auditDecisionContext,
			want:      "var help = \"auth_decision:_Authorization_decision_to_match\" // auth_decision: Authorization decision to match" + auditDecisionContext,
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit authorization decision filter boundaries and missing decision",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditDecisionLookalikes, want: auditDecisionLookalikes,
		},
		{
			name: "audit request duration filters in code and kept comments",
			opt:  Options{Glossary: tb, KeepComments: true},
			in: "var minHelp = \"" + minRequestDurationFilter + "\" // minDurationMs=100: " + minRequestDurationFilter + "\n" +
				"var maxHelp = \"" + maxRequestDurationFilter + "\" // maxDurationMs=500: " + maxRequestDurationFilter + requestDurationContext,
			want: "var minHelp = \"Filter_by_minimum_request_duration_in_ms\" // minDurationMs=100: Filter by minimum request duration in ms\n" +
				"var maxHelp = \"Filter_by_maximum_request_duration_in_ms\" // maxDurationMs=500: Filter by maximum request duration in ms" + requestDurationContext,
			wantTerms: 4, wantInCom: 2,
		},
		{
			name: "audit request duration filters boundaries and other timings",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   requestDurationLookalikes, want: requestDurationLookalikes,
		},
		{
			name:      "audit TLS version filter in code and a kept comment",
			opt:       Options{Glossary: tb, KeepComments: true},
			in:        "var help = \"" + auditTLSVersionFilter + ".\" // " + auditTLSVersionFilter + auditTLSVersionContext,
			want:      "var help = \"This_filters_all_TLS_connections_using_version_1.1_or_lower.\" // This filters all TLS connections using version 1.1 or lower" + auditTLSVersionContext,
			wantTerms: 2, wantInCom: 1,
		},
		{
			name: "audit TLS version filter boundaries and absent negotiated version",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   auditTLSVersionLookalikes, want: auditTLSVersionLookalikes,
		},
		{
			name: "nil glossary is the old behaviour",
			opt:  Options{},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 渗透 = 1\n",
		},
	}
	goLang := lang.ByName("go")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := Strip([]byte(tc.in), goLang, tc.opt)
			if string(res.Output) != tc.want {
				t.Errorf("output\n got  %q\n want %q", res.Output, tc.want)
			}
			if len(res.Terms) != tc.wantTerms {
				t.Errorf("terms = %d, want %d: %+v", len(res.Terms), tc.wantTerms, res.Terms)
			}
			in := 0
			for _, h := range res.Terms {
				if h.InComment {
					in++
				}
			}
			if in != tc.wantInCom {
				t.Errorf("in-comment hits = %d, want %d", in, tc.wantInCom)
			}
			again := Strip(res.Output, goLang, tc.opt)
			if string(again.Output) != tc.want || len(again.Terms) != 0 {
				t.Errorf("second pass: output %q, hits %+v", again.Output, again.Terms)
			}
		})
	}
}

// TestTermsUnsafeKeepsOriginal covers a replacement that opens a literal it
// does not close. The file must come back untouched: Output equals the input
// whenever Unsafe is set without Force, so a caller writing it back cannot
// save a file whose structure the replacement just broke.
func TestTermsUnsafeKeepsOriginal(t *testing.T) {
	// The replacement injects a backtick in code position, which opens a Go
	// raw string. Those are multiline, so it runs to EOF unterminated.
	tb := glossary.NewTable([]glossary.Rule{{From: "QUOTE", To: "`"}})
	in := "var s = QUOTE\n// c\n"
	res := Strip([]byte(in), lang.ByName("go"), Options{Glossary: tb})

	if !res.Unsafe {
		t.Errorf("Unsafe = false, want true: %q", res.Output)
	}
	if string(res.Output) != in {
		t.Errorf("output = %q, want the input %q", res.Output, in)
	}
	if len(res.Terms) != 1 {
		t.Errorf("terms = %+v, want the hit to still be reported", res.Terms)
	}
	if res.Reason == "" {
		t.Error("Reason is empty, want it to name the cause")
	}
}

// TestTermsForceStillStrips is the counterpart: with Force the caller has
// said it accepts the risk, so the rewrite goes through.
func TestTermsForceStillStrips(t *testing.T) {
	tb := glossary.NewTable([]glossary.Rule{{From: "QUOTE", To: "`"}})
	in := "var s = QUOTE\n// c\n"
	res := Strip([]byte(in), lang.ByName("go"), Options{Glossary: tb, Force: true})
	if string(res.Output) == in {
		t.Errorf("output unchanged with Force: %q", res.Output)
	}
}
