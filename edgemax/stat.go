package edgemax

// SystemStat is a stat which contains system statistics for an EdgeMAX device.
type SystemStat struct {
	CPU    string `json:"cpu"`
	Uptime string `json:"uptime"`
	Mem    string `json:"mem"`
}

// DPIStat contains Deep Packet Inspection stats from an EdgeMAX device.
type DPIStat map[string]map[string]struct {
	RXBytes string `json:"rx_bytes"`
	TXBytes string `json:"tx_bytes"`
}

// InterfaceStat contains network interface data transmission statistics.
type InterfacesStat map[string]struct {
	Mac   string `json:"mac"`
	Stats struct {
		RXBytes string `json:"rx_bytes"`
		TXBytes string `json:"tx_bytes"`
	} `json:"stats"`
}
