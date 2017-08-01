package lib

// --- Struct

type MetricConfig struct {
	Metrics []struct {
		Hostname string `yaml:"hostname"`
		Plugins  []struct {
			PluginName   string `yaml:"plugin_name"`
			PluginOption string `yaml:"plugin_option"`
		} `yaml:"plugins"`
	} `yaml:"metrics"`
}

type CrawlConfig struct {
	Agents []CrawlConfigAgent `yaml:"agents"`
}

type CrawlConfigAgent struct {
	GroupName string   `yaml:"group_name" json:"group_name"`
	IP        string   `yaml:"ip" json:"ip"`
	Hostname  string   `yaml:"hostname" json:"hostname"`
	Port      int      `yaml:"port" json:"port"`
	Proxies   []string `yaml:"proxies" json:"proxies"`
	Disabled  bool     `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

func (conf CrawlConfig) GetEnabledAgents() []CrawlConfigAgent {
	enabledAgents := make([]CrawlConfigAgent, 0, len(conf.Agents))
	for _, a := range conf.Agents {
		if !a.Disabled {
			enabledAgents = append(enabledAgents, a)
		}
	}
	return enabledAgents
}

type InventoryCommandsConfig struct {
	Linux map[string][]struct {
		Command       string `yaml:"command"`
		CommandOption string `yaml:"command_option"`
	} `yaml:"linux"`
}
