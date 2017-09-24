package api

const (
	getConfigSuffix       = "/daemon/%s/config"
	postDataSuffix        = "/daemon/%s/pushdata/%s?format=%s"
	checkinSuffix         = "/daemon/%s/checkin"
	configProcessedSuffix = "/daemon/%s/config/processed"
	hostnameParam         = "hostname"
	tokenParam            = "token"
	versionParam          = "version"
	currentMillisParam    = "currentMillis"
	bytesLeftParam        = "bytesLeftForBuffer"
	bytesPerMinParam      = "bytesPerMinuteForBuffer"
	currentQueueSizeParam = "currentQueueSize"
	localParam            = "local"
	pushParam             = "push"
	ephemeralParam        = "ephemeral"
	contentType           = "Content-Type"
	textPlain             = "text/plain"
	applicationJSON       = "application/json"

	NotAcceptableStatusCode = 406
	FormatGraphiteV2        = "graphite_v2"
	GraphiteBlockWorkUnit   = "12b37289-90b2-4b98-963f-75a27110b8da"
)
