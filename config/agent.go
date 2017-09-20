package config

type AgentConfig struct {
	Name             string
	AllowAnyHostKeys bool
	Targets          []string
	WorkUnits        []string
	PointsPerBatch   int
}
