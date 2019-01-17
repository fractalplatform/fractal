package metrics

type Config struct {
	MetricsFlag  bool   `mapstructure:"test-metricsflag"`
	InfluxDBFlag bool   `mapstructure:"test-influxdbflag"`
	Url          string `mapstructure:"test-influxdburl"`
	DataBase     string `mapstructure:"test-influxdbname"`
	UserName     string `mapstructure:"test-influxdbuser"`
	PassWd       string `mapstructure:"test-influxdbpasswd"`
	NameSpace    string `mapstructure:"test-influxdbnamespace"`
}
