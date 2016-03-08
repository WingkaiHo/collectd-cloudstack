// collectd-cloudstack project main.go
package main

import (
	cloudstack "collectd-cloudstack/golang-cloudstack-library"
//	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"bufio"
	"os"
	"strconv"
	"flag"
	"time"
)

var csmgr_host *string
var api_key	*string
var secret_key *string
var mgr_port *int
var curr_time int64
var user_name string
var password string
var client *cloudstack.Client

var m_host_vm_running map[string]int
var m_host_vm_stopping map[string]int
var m_host_vm_stopped map[string]int
var m_host_vm_starting map[string]int

var m_zone_vm_running map[string]int
var m_zone_vm_stopping map[string]int
var m_zone_vm_stopped map[string]int
var m_zone_vm_starting map[string]int

func get_submit_number_stat_str(host, plugin, plugin_ins, str_type, str_type_ins, value string, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL csmgr_%s/%s-%s/%s-%s %d:%s\n", host, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}

func collect_host_status(client *cloudstack.Client) {
	var stat string 
	
	param := cloudstack.NewListHostParam()
	param.State.Set("Up")
	param.ResourceState.Set("Enabled")
	param.Type.Set("Routing")
	hosts, err := client.ListHost(param)
	f := bufio.NewWriter(os.Stdout)
	
	if err != nil {
		fmt.Errorf("Fail to exectue function listHost err is %s", err.Error())
		return
	}
	
	for i := range hosts {
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"guage", "capabilities", hosts[i].Capabilities.String(), curr_time)
			
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"guage", "cpu_allocated", hosts[i].CPUAllocated.String(), curr_time)
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"guage", "cpu_prov_total", hosts[i].CPUWithoverProvisioning.String(), curr_time)
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"guage", "memory_total", hosts[i].Memorytotal.String(), curr_time)
			
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"guage", "memory_allocated", hosts[i].MemoryAllocated.String(), curr_time)
			
		m_host_vm_running[hosts[i].Name.String()] = 0
		m_host_vm_stopped[hosts[i].Name.String()] = 0
		m_host_vm_stopping[hosts[i].Name.String()] = 0
		m_host_vm_starting[hosts[i].Name.String()] = 0
		
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

/**
 * @brief Collect the number of vm running, stop, stopping starting of zone
*/
func collect_vm_number(client *cloudstack.Client) {
	var stat string
	
	param := cloudstack.NewListVirtualMachinesParameter()
	param.ListAll.Set(true)
	vms, err := client.ListVirtualMachines(param)
	f := bufio.NewWriter(os.Stdout)
	
	if err != nil {
		fmt.Errorf("Fail to execute function ListVirtualMachines err is %s", err.Error())
	}
	
	for i := range vms {
		switch vms[i].State.String() {
			case "Running":
				m_zone_vm_running[vms[i].ZoneName.String()] += 1
				m_host_vm_running[vms[i].HostName.String()] += 1
				break 
			case "Stopped":
				m_zone_vm_stopped[vms[i].ZoneName.String()] += 1
				m_host_vm_stopped[vms[i].HostName.String()] += 1
				break
			case "Starting":
				m_zone_vm_starting[vms[i].ZoneName.String()] += 1
				m_host_vm_starting[vms[i].HostName.String()] += 1
				break
			case "Stopping":
				m_zone_vm_stopping[vms[i].ZoneName.String()] += 1
				m_host_vm_stopping[vms[i].HostName.String()] += 1
				break;
				
		}
	}
	
	for key, running_value := range m_zone_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"guage", "user_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"guage", "user_vms_stopped", strconv.Itoa(m_zone_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"guage", "user_vms_startting", strconv.Itoa(m_zone_vm_starting[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"guage", "user_vms_stopping", strconv.Itoa(m_zone_vm_stopping[key]), curr_time)
	}
	
	for key, running_value := range m_host_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"guage", "user_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"guage", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"guage", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"guage", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

func collect_zone_capacity(client *cloudstack.Client) {
	var stat string
	var t int64
	var err error
	var type_total_name, type_used_name, type_used_pect_name string
	var c []*cloudstack.Capacity
	
	f := bufio.NewWriter(os.Stdout)
	param := cloudstack.NewListCapacityParamete()
	c, err = client.ListCapacity(param)
	
	if err != nil {
		fmt.Errorf("Fail to exectue function ListCapacity err is %s", err.Error())
	}

	for i := range c {
		t, err = c[i].Type.Int64()
		if err != nil || c[i].ZoneName.IsNil() || c[i].CapacityUsed.IsNil() ||
			c[i].CapacityTotal.IsNil() {
			continue
		}

		switch (t) {
		case 0:
			type_total_name = "mem_total"
			type_used_name = "mem_used"
			type_used_pect_name = "mem_used_percent"
			break
		case 1:
			type_total_name = "cpu_total"
			type_used_name = "cpu_used"
			type_used_pect_name = "cpu_used_percent"
			break
		case 2:
			type_total_name = "primary_storage_total"
			type_used_name = "primary_storage_used"
			type_used_pect_name = "primary_storage_percent"
			break
		case 4:
			type_total_name = "virtual_network_public_ip_total"
			type_used_name = "virtual_network_public_ip_used"
			type_used_pect_name = "virtual_network_public_ip_used_percent"
			break
		case 5:
			type_total_name = "private_ip_total"
			type_used_name = "private_ip_used"
			type_used_pect_name = "private_ip_used_percent"
			break
		case 6:
			type_total_name = "privat_total"
			type_used_name = "private_ip_used"
			type_used_pect_name = "private_ip_used_percent"
		case 7:
			type_total_name = "vlan_total"
			type_used_name = "vlan_used"
			type_used_pect_name = "vlan_used_percent"
			break;
		case 8:
			type_total_name = "direct_attached_public_ip_total"
			type_used_name = "direct_attached_public_ip_used"
			type_used_pect_name = "direct_attached_public_ip_used_percent"
		default:
			continue
		}
		
		stat += get_submit_number_stat_str(*csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_total_name, c[i].CapacityTotal.String(), curr_time)
		
		stat += get_submit_number_stat_str(*csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_used_name, c[i].CapacityUsed.String(), curr_time)
			
		stat += get_submit_number_stat_str(*csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_used_pect_name, c[i].PercentUsed.String(), curr_time)
		
		m_zone_vm_running[c[i].ZoneName.String()] = 0
		m_zone_vm_starting[c[i].ZoneName.String()] = 0
		m_zone_vm_stopped[c[i].ZoneName.String()] = 0
		m_zone_vm_stopping[c[i].ZoneName.String()] = 0
			
	}
	f.Write([]byte(stat))
	f.Flush()
}

func main() {
	log.SetOutput(ioutil.Discard)
	
	var err error
	csmgr_host = flag.String("host", "localhost", "The hostname of cloudstack manager.")
	api_key = flag.String("apikey", "", "API key from an account on the root level.")
	secret_key = flag.String("secret", "", "Associated API Secret from the account")
	mgr_port = flag.Int("port", 8080, "The port of cloudstack manager of access. Default is 8080")
	user_name = "admin"
	password = "password"
	
	flag.Parse()
	
	if *api_key == "" {
		fmt.Errorf("Please use -apikey to set api key from an account on the root level.")
	}
	
	if *secret_key == "" {
		fmt.Errorf("Please use -secret to set secretkey to associated API Secret from the account")
	}
	
	endpoint, err := url.Parse("http://"+*csmgr_host+":"+strconv.Itoa(*mgr_port)+"/client/api")
	if err != nil {
		fmt.Errorf("Fail to parse the url.")
	}
	
	client, err = cloudstack.NewClient(endpoint, *api_key, *secret_key, user_name, password)
	if err != nil {
		fmt.Errorf("Fail to create the cloudstack client instance.")
	}
	
	m_host_vm_running = make(map[string]int)
	m_host_vm_stopping = make(map[string]int)
	m_host_vm_stopped = make(map[string]int)
	m_host_vm_starting = make(map[string]int)
	
	m_zone_vm_running = make(map[string]int)
	m_zone_vm_stopping = make(map[string]int)
	m_zone_vm_stopped = make(map[string]int)
	m_zone_vm_starting = make(map[string]int)
	
	curr_time = time.Now().Unix();
	collect_zone_capacity(client)
	collect_host_status(client)
	collect_vm_number(client)
}
