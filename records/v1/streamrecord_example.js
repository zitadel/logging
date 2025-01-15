// The streamRecord has flat properties in snake_case.
// This makes them very easy map them to database columns and use them in queries.
// The fields following fields are potentially set in all stream records.
// - "stream"
// - "version"
// - "observed_time"
// - "instance_id"
// - "trace_id",
// - "span_id"
// The properties "stream" and "version" define what additional fields are potentially available in the record
streamRecord = {
  // Possible values for "stream":
  // - "request_grpc"
  // - "request_http"
  // - "runtime_version",
  // - "runtime_service",
  // - "notification"
  // - "action_targetcall"
  // - "action_trigger_grpc_request"
  // - "action_trigger_grpc_response"
  // - "action_trigger_event"
  // - "action_trigger_function"
  // - "event"
  "stream": "request_http",
  "version": "v1",
  "observed_time": "20250114T162059Z", // ISO 8601 Timestamp
  "trace_id": "123",
  "span_id": "123",
  "instance_id": "1234567890123", // If available
  "org_id": "1234567890123", // If available
  "user_id": "1234567890123", // If available

  // additional static properties if configured that should be present in all records.
  // for example
  // - "region": "US1",
  // - "runtime_service_version": "v2.67.2" // so it is not only available in a single runtime_service record but in all records

  // The "runtime" stream contains normal log records that are written all over the Zitadel code.
  // if stream is runtime_*
  "runtime_severity": "info",
  "runtime_message": "user created",
  // runtime_attributes_* contains additional information passed in the log record
  // these properties have no guaranteed schema.
  "runtime_attributes_userid": "1234567890123",

  // if stream is runtime_service
  // A record is only written once in a runtime lifecycle
  "runtime_service_name": "zitadel",
  "runtime_service_version": "v2.67.2",
  "runtime_service_process": "sdsf321ew6f5", // For example Pod ID

  // runtime_error
  "runtime_error_cause": "user not found by email user@example.com: no rows in result set",
  "runtime_error_stack": "line1\nline2\nline3",
  "runtime_error_i18n_key": "Errors.User.NotFound", // If error is of type ZitadelError
  "runtime_error_type": "InternalError", // If error is of type ZitadelError

  // request*
  "request_is_system_user": false,
  "request_is_authenticated": true,
  "request_latency": "50ms",

  // request_http
  "request_http_protocol": "",
  "request_http_host": "",
  "request_http_port": "",
  "request_http_path": "",
  "request_http_method": "",
  "request_http_status": 200,
  "request_http_referer": "",
  "request_http_user_agent": "",
  "request_http_remote_ip": "",
  "request_http_bytes_received": 1000,
  "request_http_bytes_sent": 1000,

  // request_grpc
  "request_grpc_service": "",
  "request_grpc_method": "",
  "request_grpc_code": "",

  // action_targetcall

  "action_targetcall_target_id": "",
  "action_targetcall_name": "",
  "action_targetcall_protocol": "",
  "action_targetcall_host": "",
  "action_targetcall_port": "",
  "action_targetcall_path": "",
  "action_targetcall_method": "",
  "action_targetcall_status": 200,

  //action_trigger_grpc*
  "action_trigger_grpc_service": "",
  "action_trigger_grpc_method": "",

  // action_trigger_grpc_request
  // (no additional properties)

  // action_trigger_grpc_response
  "action_trigger_grpc_response_code": 200,

  // action_trigger_event
  "action_trigger_event_id": "",

  // action_trigger_function
  "action_trigger_function_name": "",

  // event
  "event_id": "",
  "event_sequence": "",
  "event_position": "",
  "event_type": "",
  "event_data": {},
  "event_editor_user": "",
  "event_version": "",
  "event_aggregate_id": "",
  "event_aggregate_type": "",
  "event_resource_owner": "",
}