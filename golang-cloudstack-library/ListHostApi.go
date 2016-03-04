package cloudstack

type ListHostParameter struct {
	// lists hosts existing in particular cluster
	ClusterId  ID
	// comma separated list of host details requested, value can be a list of [ min, all, capacity, events, stats]
	Details []string
	// if true, list only hosts dedicated to HA
	HaHost NullBool
	// hypervisor type of host: XenServer,KVM,VMware,Hyperv,BareMetal,Simulator
	Hypervisor NullString
	// List by keyword
	Keyword NullString 
	// the name of the host
	Name NullString
	Page      NullNumber
	PageSize  NullNumber
	// the pod ID
	PodId ID
	// list hosts by resource state. Resource state represents current state determined by admin of host, valule can be one of 
	// [Enabled, Disabled, Unmanaged, PrepareForMaintenance, ErrorInMaintenance, Maintenance, Error]
	ResourceState NullString
	// the state of the host
	State NullString
	// the host type
	Type  NullString
	// lists hosts in the same cluster as this VM and flag hosts with enough CPU/RAm to host this VM
	VirtualmachineId NullString
	ZoneId NullString
}


type Host struct {
	ResourceBase
	Id 	ID `json:"id"`
	AverageLoad	NullNumber  `json:"averageload"`
	Capabilities	NullNumber `json:"Capabilities"`
	ClusterId ID 	`json:"clusterid"`
	ClusterName NullString 	`json:"clustername"`
	ClusterType NullString 	`json:"clustertype"`
	CPUAllocated NullNumber 	`json:"cpuallocated"`
	CPUNumber NullNumber 	`json:"cpunumber"`
	CPUSockerts NullNumber 	`json:"cpusockets"`
	CPUSpeed		NullNumber	`json:"cpuspeed"`
	CPUUsed		NullNumber	`json:"cpuused"`
	CPUWithoverProvisioning NullNumber `json:"cpuwithoverprovisioning"`
	Created NullString 		`json:"created"`
	Details NullString 		`json:"details"`
	Disconnected NullNumber `json:"Disconnected"`
	DiskSizeAllocated NullNumber	`json:"disksizeallocated"`
	DisksizeTotal NullNumber		`json:"disksizetotal"`
	Events NullString	`json:"events"`
	Hahost NullBool	`json:"hahost"`
	HasEnoughCapacity NullBool `json:"hasenoughcapacity"`
	HosTags NullString `json:"hosttag"`
	Hypervisor NullString `json:"hypervisor"`
	HypervisorVersion NullString `json:"hypervisorversion"`
	Ipaddress NullString `json:"ipaddress"`
	IsLocalStorageActive NullBool `json:"islocalstorageactive"`
	LastPinged NullString `json:"lastpinged"`
	Managementserverid NullString `json:"managementserverid"`
	MemoryAllocated NullNumber `json:"memoryallocated"`
	Memorytotal NullNumber `json:"memorytotal"`
	MemoryUsed NullNumber `json:"memoryused"`
	Name NullString `json:"name"`
	NetworkKBbsRead NullNumber`json:"networkkbsread"`
	NetworkKBbsWrite NullNumber `json:"networkkbsread"`
	OscategoryId	 ID `json:"oscategoryid"`
	OscategoryName NullString `json:"oscategoryname"`
	PodId ID `json:"podid"`
	PodName NullString `json:"podname"`
	Removed NullString `json:"removed"`
	ResourceState NullString `json"resourcestate"`
	State NullString `json:"state"`
	SuitableForMigration NullBool `json:"suitableformigration"`
	Type 	NullString `json:"type"`
	Version NullString `json:"version"`
	ZoneId	ID	`json:"zoneid"`		
	ZoneName NullString `json:"zonename"`
	JobId	ID 	`json:"jobid"`
	Jobstatus NullString `json:"jobstatus"`
}

func NewListHostParam() (p *ListHostParameter) {
	p = new(ListHostParameter)
	return p
}

func (c *Client) ListHost(p *ListHostParameter) ([]*Host, error) {
	obj, err := c.Request("listVirtualMachines", convertParamToMap(p))
	if err != nil {
		return nil, err
	}
	return obj.([]*Host), err
}