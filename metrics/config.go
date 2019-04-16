package metrics

type Config struct {
	MetricsFlag  bool   `mapstructure:"metrics"`
	InfluxDBFlag bool   `mapstructure:"influxdb"`
	URL          string `mapstructure:"influxdburl"`
	DataBase     string `mapstructure:"influxdbname"`
	UserName     string `mapstructure:"influxdbuser"`
	PassWd       string `mapstructure:"influxdbpasswd"`
	NameSpace    string `mapstructure:"influxdbnamespace"`
}
