package metric

type KNodeNetworkBytes struct {
	Node      string
	Interface string
	Bytes     uint64
}

type KNodeNetworkPackets struct {
	Node      string
	Interface string
	Packets   uint64
}

type KNodeNetworkErrors struct {
	Node      string
	Interface string
	Errors    uint64
}

type KPodNetworkBytes struct {
	Pod       string
	Interface string
	Bytes     uint64
}

type KPodNetworkPackets struct {
	Pod       string
	Interface string
	Packets   uint64
}

type KPodNetworkErrors struct {
	Pod       string
	Interface string
	Errors    uint64
}
