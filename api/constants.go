package api

const GET_CONFIG_SUFFIX = "/daemon/%s/config"
const POST_DATA_SUFFIX = "/daemon/%s/pushdata/%s?format=%s"
const CHECKIN_SUFFIX = "/daemon/%s/checkin"
const CONFIG_PROCESSED_SUFFIX = "/daemon/%s/config/processed"
const HOSTNAME = "hostname"
const TOKEN = "token"
const VERSION = "version"
const CURRENT_MILLIS = "currentMillis"
const BYTES_LEFT = "bytesLeftForBuffer"
const BYTES_PER_MIN = "bytesPerMinuteForBuffer"
const CURR_QUEUE_SIZE = "currentQueueSize"
const LOCAL = "local"
const PUSH = "push"
const EPHEMERAL = "ephemeral"
const CONTENT_TYPE = "Content-Type"
const TEXT_PLAIN = "text/plain"
const APPLICATION_JSON = "application/json"
const NOT_ACCEPTABLE_STATUS_CODE = 406

const FORMAT_GRAPHITE_V2 = "graphite_v2"
const GRAPHITE_BLOCK_WORK_UNIT = "12b37289-90b2-4b98-963f-75a27110b8da"
