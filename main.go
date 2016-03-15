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
	"strings"
)

var csmgr_host *string
var api_key	*string
var secret_key *string
var mgr_port *int
var curr_time int64
var user_name string
var password string
var client *cloudstack.Client
var interval, debug int
var inited int

// statistics the user vm number in host or zone
var m_host_vm_running map[string]int
var m_host_vm_stopping map[string]int
var m_host_vm_stopped map[string]int
var m_host_vm_starting map[string]int

var m_zone_vm_running map[string]int
var m_zone_vm_stopping map[string]int
var m_zone_vm_stopped map[string]int
var m_zone_vm_starting map[string]int

// statistics the system vm(VR console proxy vm and second storage vm) number in host or zone
var m_host_sys_vm_running map[string]int
var m_host_sys_vm_stopping map[string]int
var m_host_sys_vm_stopped map[string]int
var m_host_sys_vm_starting map[string]int

var m_zone_sys_vm_running map[string]int
var m_zone_sys_vm_stopping map[string]int
var m_zone_sys_vm_stopped map[string]int
var m_zone_sys_vm_starting map[string]int

var m_host_user_vm_ifread_curr	map[string]int64
var m_host_user_vm_ifread_last	map[string]int64

var m_host_user_vm_ifwrite_curr	map[string]int64
var m_host_user_vm_ifwrite_last	map[string]int64	

var m_host_user_vm_diskread_curr map[string]int64
var m_host_user_vm_diskread_last map[string]int64

var m_host_user_vm_diskwrite_curr map[string]int64
var m_host_user_vm_diskwrite_last map[string]int64

var m_host_user_vm_diskioread_curr map[string]int64
var m_host_user_vm_diskioread_last map[string]int64

var m_host_user_vm_diskiowrite_curr map[string]int64
var m_host_user_vm_diskiowrite_last map[string]int64

var m_zone_user_vm_ifread_curr map[string]int64
var m_zone_user_vm_ifread_last map[string]int64

var m_zone_user_vm_ifwrite_curr	map[string]int64
var m_zone_user_vm_ifwrite_last	map[string]int64	

var m_zone_user_vm_diskread_curr map[string]int64
var m_zone_user_vm_diskread_last map[string]int64

var m_zone_user_vm_diskwrite_curr map[string]int64
var m_zone_user_vm_diskwrite_last map[string]int64

var m_zone_user_vm_diskioread_curr map[string]int64
var m_zone_user_vm_diskioread_last map[string]int64

var m_zone_user_vm_diskiowrite_curr map[string]int64
var m_zone_user_vm_diskiowrite_last map[string]int64

func get_submit_number_stat_str(host, plugin, plugin_ins, str_type, str_type_ins, value string, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL csmgr_%s/%s-%s/%s-%s %d:%s\n", host, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}

func extrace_precent_value(v string)(float64) {
	num_end := strings.Index(v, "%")
	if num_end > 0  {
		v = string(v[0:num_end-1])
	}
	
	result, err:= strconv.ParseFloat(v, 64)
	if err != nil {
		return 0.0
	}
	
	return result
}

func collect_sys_vm_number(client *cloudstack.Client) {
	var stat string
	var vmstat string
	f := bufio.NewWriter(os.Stdout)
	
	router_param := cloudstack.NewListRouterParam()
	router_param.Listall.Set("true")
	routers, routers_err := client.ListRouter(router_param)
	
	// statistics the number of router in zone and host
	if routers_err == nil {
		for i := range routers {
			switch routers[i].State.String() {
			case "Running":
				m_zone_sys_vm_running[routers[i].ZoneName.String()] += 1
				m_host_sys_vm_running[routers[i].HostName.String()] += 1
				vmstat = "3"
				break 
			case "Starting":
				m_zone_sys_vm_starting[routers[i].ZoneName.String()] += 1
				m_host_sys_vm_starting[routers[i].HostName.String()] += 1
				vmstat = "2"
				break
			case "Stopping":
				m_zone_sys_vm_stopping[routers[i].ZoneName.String()] += 1
				m_host_sys_vm_stopping[routers[i].HostName.String()] += 1
				vmstat = "1"
				break;
			case "Stopped":
				m_zone_sys_vm_stopped[routers[i].ZoneName.String()] += 1
				m_host_sys_vm_stopped[routers[i].HostName.String()] += 1
				vmstat = "0"
				break
			case "Error":
				vmstat = "-1"
				break
			default:
				continue
			}
			stat += get_submit_number_stat_str(*csmgr_host, "systemvm", routers[i].Name.String(),  
					"gauge", "status", vmstat, curr_time)
		}
	} else {
		fmt.Errorf("Fail to execute function listRouters err is %s\n", routers_err.Error())
	}
	
	// statistics the number of second storage, console proxy in host and zone
	sysvms_param := cloudstack.NewListSystemVmsParam()
	sysvms, sysvms_err := client.ListSystemVms(sysvms_param)
	if sysvms_param != nil {
		for i := range sysvms {
			switch sysvms[i].State.String() {
			case "Running":
				m_zone_sys_vm_running[sysvms[i].ZoneName.String()] += 1
				m_host_sys_vm_running[sysvms[i].HostName.String()] += 1
				stat += get_submit_number_stat_str(*csmgr_host, "systemvm", sysvms[i].Name.String(),  
					"gauge", "status", "3", curr_time)
				break
			case "Starting":
				m_zone_sys_vm_starting[sysvms[i].ZoneName.String()] += 1
				m_host_sys_vm_starting[sysvms[i].HostName.String()] += 1
				stat += get_submit_number_stat_str(*csmgr_host, "systemvm", sysvms[i].Name.String(),  
					"gauge", "status", "2", curr_time)
				break 
			case "Stopping":
				m_zone_sys_vm_stopping[sysvms[i].ZoneName.String()] += 1
				m_host_sys_vm_stopping[sysvms[i].HostName.String()] += 1
				stat += get_submit_number_stat_str(*csmgr_host, "systemvm", sysvms[i].Name.String(),  
					"gauge", "status", "1", curr_time)
				break
			case "Stopped":
				m_zone_sys_vm_stopped[sysvms[i].ZoneName.String()] += 1
				m_host_sys_vm_stopped[sysvms[i].HostName.String()] += 1
				stat += get_submit_number_stat_str(*csmgr_host, "systemvm", sysvms[i].Name.String(),  
					"gauge", "status", "0", curr_time)
				break
			case "Error":
				stat += get_submit_number_stat_str(*csmgr_host, "systemvm", sysvms[i].Name.String(),  
					"gauge", "status", "-1", curr_time)
				break;
			}
		}
	} else {
		fmt.Errorf("Fail to execute function listSystemVms err is %s\n", sysvms_err.Error())
	}
	
	
	for key, running_value := range m_zone_sys_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_sys_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_sys_vms_stopped", strconv.Itoa(m_zone_sys_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_sys_vms_startting", strconv.Itoa(m_zone_sys_vm_starting[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_sys_vms_stopping", strconv.Itoa(m_zone_sys_vm_stopping[key]), curr_time)
	}
	
	for key, running_value := range m_host_sys_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_sys_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_sys_vms_stopped", strconv.Itoa(m_host_sys_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_sys_vms_startting", strconv.Itoa(m_host_sys_vm_starting[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_sys_vms_stopping", strconv.Itoa(m_host_sys_vm_stopping[key]), curr_time)
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

func collect_host_status(client *cloudstack.Client) {
	var stat string 
	var isha int
	param := cloudstack.NewListHostParam()
	param.State.Set("Up")
	param.ResourceState.Set("Enabled")
	param.Type.Set("Routing")
	hosts, err := client.ListHost(param)
	f := bufio.NewWriter(os.Stdout)
	
	if err != nil {
		fmt.Errorf("Fail to execute function listHost err is %s\n", err.Error())
		return
	}
	
	for i := range hosts {
		// find the character '%' in CPUAllocated string
		str := hosts[i].CPUAllocated.String()
		num_end := strings.Index(str, "%")
		if num_end > 0  {
			str = string(str[0:num_end-1])
		}
		
		cpu_allocated_percent, parse_err:= strconv.ParseFloat(str, 64)
		if parse_err == nil {
			cpu_allocated_percent /= 100.0
			stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
				"gauge", "cpu_allocated_precent", hosts[i].CPUAllocated.String(), curr_time)
		}
			
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"gauge", "cpu_prov_total", hosts[i].CPUWithoverProvisioning.String(), curr_time)
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"gauge", "memory_total", hosts[i].Memorytotal.String(), curr_time)
			
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"gauge", "memory_allocated", hosts[i].MemoryAllocated.String(), curr_time)
			
		
		if hosts[i].Hahost.Bool() == true {
			isha = 1
		} else {
			isha = 0
		}
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", hosts[i].Name.String(),  
			"gauge", "hahost", strconv.Itoa(isha), curr_time)
			
		m_host_vm_running[hosts[i].Name.String()] = 0
		m_host_vm_stopped[hosts[i].Name.String()] = 0
		m_host_vm_stopping[hosts[i].Name.String()] = 0
		m_host_vm_starting[hosts[i].Name.String()] = 0
		
		m_host_sys_vm_running[hosts[i].Name.String()] = 0
		m_host_sys_vm_stopped[hosts[i].Name.String()] = 0
		m_host_sys_vm_stopping[hosts[i].Name.String()] = 0
		m_host_sys_vm_starting[hosts[i].Name.String()] = 0
		
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

/**
 * @brief Collect the number of vm running, stop, stopping starting of zone
*/
func collect_user_vm_number(client *cloudstack.Client) {
	var stat string
	var ifread, ifwrite int64
	var diskread, diskwrite int64 
	var diskioread, diskiowrite int64
	//var cpuused float64
	var err error
	
	param := cloudstack.NewListVirtualMachinesParameter()
	param.ListAll.Set(true)
	vms, err := client.ListVirtualMachines(param)
	f := bufio.NewWriter(os.Stdout)
	
	if err != nil {
		fmt.Errorf("Fail to execute function ListVirtualMachines err is %s\n", err.Error())
		return
	}
	
	for i := range vms {
		switch vms[i].State.String() {
			case "Running":
				m_zone_vm_running[vms[i].ZoneName.String()] += 1
				m_host_vm_running[vms[i].HostName.String()] += 1
				/*if vms[i].CpuUsed.IsNil() == false {
					cpuused += extrace_precent_value(vms[i].CpuUsed.String())
				}*/
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
		
		ifread, err = vms[i].NetworkKbsRead.Int64()
		if err == nil {
			ifread *= 1024
			m_host_user_vm_ifread_curr[vms[i].HostName.String()] += ifread
			m_zone_user_vm_ifread_curr[vms[i].ZoneName.String()] += ifread
		}
		
		ifwrite, err = vms[i].NetworkKbsWrite.Int64()
		if err == nil {
			ifwrite *= 1024
			m_host_user_vm_ifwrite_curr[vms[i].HostName.String()] += ifwrite
			m_zone_user_vm_ifwrite_curr[vms[i].ZoneName.String()] += ifwrite
		}
		
		diskread, err = vms[i].DiskKbsRead.Int64()
		if err == nil {
			diskread *= 1024
			m_host_user_vm_diskread_curr[vms[i].HostName.String()] += diskread
			m_zone_user_vm_diskread_curr[vms[i].ZoneName.String()] += diskread
		}
		
		diskwrite, err = vms[i].DiskKbsWrite.Int64()
		if err == nil {
			diskread *= 1024
			m_host_user_vm_diskwrite_curr[vms[i].HostName.String()] += diskwrite
			m_zone_user_vm_diskwrite_curr[vms[i].ZoneName.String()] += diskwrite
		}
		
		diskioread, err = vms[i].DiskIoRead.Int64()
		if err == nil {
			m_host_user_vm_diskioread_curr[vms[i].HostName.String()] += diskioread
			m_zone_user_vm_diskioread_curr[vms[i].ZoneName.String()] += diskioread
		}
		
		diskiowrite, err = vms[i].DiskIoWrite.Int64()
		if err == nil {
			m_host_user_vm_diskiowrite_curr[vms[i].HostName.String()] += diskiowrite
			m_zone_user_vm_diskiowrite_curr[vms[i].ZoneName.String()] += diskiowrite
		}
	}
	
	for key, running_value := range m_zone_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_stopped", strconv.Itoa(m_zone_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_startting", strconv.Itoa(m_zone_vm_starting[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_stopping", strconv.Itoa(m_zone_vm_stopping[key]), curr_time)
		
		if inited == 0 {
			continue
		}
		
		if m_zone_user_vm_ifread_curr[key] >=  m_zone_user_vm_ifread_last[key] {
			ifread = m_zone_user_vm_ifread_curr[key] - m_zone_user_vm_ifread_last[key]
		} else {
			ifread = 0
		}
		
		if m_zone_user_vm_ifwrite_curr[key] >=  m_zone_user_vm_ifwrite_last[key] {
			ifwrite = m_zone_user_vm_ifwrite_curr[key] - m_zone_user_vm_ifwrite_last[key]
		} else {
			ifwrite = 0
		}
		
		if m_zone_user_vm_diskread_curr[key] >= m_zone_user_vm_diskread_last[key] {
			diskread = m_zone_user_vm_diskread_curr[key] - m_zone_user_vm_diskread_last[key]
		} else {
			diskread = 0
		}
		
		if m_zone_user_vm_diskwrite_curr[key] >= m_zone_user_vm_diskwrite_last[key] {
			diskwrite = m_zone_user_vm_diskwrite_curr[key] - m_zone_user_vm_diskwrite_last[key]
		} else {
			diskwrite = 0
		}
		
		if m_zone_user_vm_diskioread_curr[key] >= m_zone_user_vm_diskioread_last[key] {
			diskioread = m_zone_user_vm_diskioread_curr[key] - m_zone_user_vm_diskioread_last[key]
		} else {
			diskioread = 0
		}
		
		if m_zone_user_vm_diskiowrite_curr[key] >= m_zone_user_vm_diskiowrite_last[key] {
			diskiowrite = m_zone_user_vm_diskiowrite_curr[key] - m_zone_user_vm_diskiowrite_last[key]
		} else {
			diskiowrite = 0
		}
		
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_network_rx", strconv.FormatInt(ifread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_network_tx", strconv.FormatInt(ifwrite, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_disk_read", strconv.FormatInt(diskread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_disk_write", strconv.FormatInt(diskwrite, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_disk_ioread", strconv.FormatInt(diskioread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "zone", key, 
			"gauge", "user_vms_disk_iowrite", strconv.FormatInt(diskiowrite, 10), curr_time)
	}
	
	for key, running_value := range m_host_vm_running {
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_running", strconv.Itoa(running_value), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_stopped", strconv.Itoa(m_host_vm_stopped[key]), curr_time)
			
		if inited == 0 {
			continue
		}
		
		if m_host_user_vm_ifread_curr[key] >=  m_host_user_vm_ifread_last[key] {
			ifread = m_host_user_vm_ifread_curr[key] - m_host_user_vm_ifread_last[key]
		} else {
			ifread = 0
		}
		
		if m_host_user_vm_ifwrite_curr[key] >=  m_host_user_vm_ifwrite_last[key] {
			ifwrite = m_host_user_vm_ifwrite_curr[key] - m_host_user_vm_ifwrite_last[key]
		} else {
			ifwrite = 0
		}
		
		if m_host_user_vm_diskread_curr[key] >= m_host_user_vm_diskread_last[key] {
			diskread = m_host_user_vm_diskread_curr[key] - m_host_user_vm_diskread_last[key]
		} else {
			diskread = 0
		}
		
		if m_host_user_vm_diskwrite_curr[key] >= m_host_user_vm_diskwrite_last[key] {
			diskwrite = m_host_user_vm_diskwrite_curr[key] - m_host_user_vm_diskwrite_last[key]
		} else {
			diskwrite = 0
		}
		
		if m_host_user_vm_diskioread_curr[key] >= m_host_user_vm_diskioread_last[key] {
			diskioread = m_host_user_vm_diskioread_curr[key] - m_host_user_vm_diskioread_last[key]
		} else {
			diskioread = 0
		}
		
		if m_host_user_vm_diskiowrite_curr[key] >= m_host_user_vm_diskiowrite_last[key] {
			diskiowrite = m_host_user_vm_diskiowrite_curr[key] - m_host_user_vm_diskiowrite_last[key]
		} else {
			diskiowrite = 0
		}
		
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_network_rx", strconv.FormatInt(ifread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_network_tx", strconv.FormatInt(ifwrite, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_disk_read", strconv.FormatInt(diskread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_disk_write", strconv.FormatInt(diskwrite, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_disk_ioread", strconv.FormatInt(diskioread, 10), curr_time)
		stat += get_submit_number_stat_str(*csmgr_host, "host", key, 
			"gauge", "user_vms_disk_iowrite", strconv.FormatInt(diskiowrite, 10), curr_time)
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

func collect_zone_capacity(client *cloudstack.Client) {
	var stat string
	var t int64
	var err error
	var type_total_name, type_used_name string
	var c []*cloudstack.Capacity
	
	f := bufio.NewWriter(os.Stdout)
	param := cloudstack.NewListCapacityParamete()
	c, err = client.ListCapacity(param)
	
	if err != nil {
		fmt.Errorf("Fail to exectue function ListCapacity err is %s\n", err.Error())
		return
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
			break
		case 1:
			type_total_name = "cpu_total"
			type_used_name = "cpu_used"
			break
		case 2:
			type_total_name = "primary_storage_total"
			type_used_name = "primary_storage_used"
			break
		case 4:
			type_total_name = "virtual_network_public_ip_total"
			type_used_name = "virtual_network_public_ip_used"
			break
		case 5:
			type_total_name = "private_ip_total"
			type_used_name = "private_ip_used"
			break
		case 6:
			type_total_name = "privat_total"
			type_used_name = "private_ip_used"
		case 7:
			type_total_name = "vlan_total"
			type_used_name = "vlan_used"
			break;
		case 8:
			type_total_name = "direct_attached_public_ip_total"
			type_used_name = "direct_attached_public_ip_used"
		default:
			continue
		}
		
		stat += get_submit_number_stat_str(*csmgr_host, "zone", c[i].ZoneName.String(), 
			"gauge", type_total_name, c[i].CapacityTotal.String(), curr_time)
		
		stat += get_submit_number_stat_str(*csmgr_host, "zone", c[i].ZoneName.String(), 
			"gauge", type_used_name, c[i].CapacityUsed.String(), curr_time)
		
		m_zone_vm_running[c[i].ZoneName.String()] = 0
		m_zone_vm_starting[c[i].ZoneName.String()] = 0
		m_zone_vm_stopped[c[i].ZoneName.String()] = 0
		m_zone_vm_stopping[c[i].ZoneName.String()] = 0
		
		m_zone_sys_vm_running[c[i].ZoneName.String()] = 0
		m_zone_sys_vm_starting[c[i].ZoneName.String()] = 0
		m_zone_sys_vm_stopped[c[i].ZoneName.String()] = 0
		m_zone_sys_vm_stopping[c[i].ZoneName.String()] = 0
		
		m_host_user_vm_ifread_last[c[i].ZoneName.String()] = m_host_user_vm_ifread_curr[c[i].ZoneName.String()] 
		m_host_user_vm_ifread_curr[c[i].ZoneName.String()] = 0

		m_host_user_vm_ifwrite_last[c[i].ZoneName.String()] = m_host_user_vm_ifwrite_curr[c[i].ZoneName.String()]
		m_host_user_vm_ifwrite_curr[c[i].ZoneName.String()] = 0

		m_host_user_vm_diskread_last[c[i].ZoneName.String()] = m_host_user_vm_diskread_curr[c[i].ZoneName.String()] 
		m_host_user_vm_diskread_curr[c[i].ZoneName.String()] = 0
		
		m_host_user_vm_diskwrite_last[c[i].ZoneName.String()] = m_host_user_vm_diskwrite_curr[c[i].ZoneName.String()]
		m_host_user_vm_diskwrite_curr[c[i].ZoneName.String()] = 0

		m_host_user_vm_diskioread_last[c[i].ZoneName.String()] = m_host_user_vm_diskioread_curr[c[i].ZoneName.String()] 
		m_host_user_vm_diskioread_curr[c[i].ZoneName.String()] = 0 

		m_host_user_vm_diskiowrite_last[c[i].ZoneName.String()] = m_host_user_vm_diskiowrite_curr[c[i].ZoneName.String()]
		m_host_user_vm_diskiowrite_curr[c[i].ZoneName.String()] = 0
			
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
	flag.IntVar(&interval, "interval", 300, "The scan interval of cloudstack resource")
	flag.IntVar(&debug, "debug", 0, "If debug mode running.")
	
	user_name = "admin"
	password = "password"
	
	flag.Parse()
	
	if interval < 20 {
		fmt.Errorf("The scan interval must more then 20 second \n")
		return
	}
	
	if *api_key == "" {
		fmt.Errorf("Please use -apikey to set api key from an account on the root level.\n")
		return
	}
	
	if *secret_key == "" {
		fmt.Errorf("Please use -secret to set secretkey to associated API Secret from the account\n")
		return
	}
	
	endpoint, err := url.Parse("http://"+*csmgr_host+":"+strconv.Itoa(*mgr_port)+"/client/api")
	if err != nil {
		fmt.Errorf("Fail to parse the url.\n")
		return
	}
	
	client, err = cloudstack.NewClient(endpoint, *api_key, *secret_key, user_name, password)
	if err != nil {
		fmt.Errorf("Fail to create the cloudstack client instance.\n")
		return
	}
	
	m_host_vm_running = make(map[string]int)
	m_host_vm_stopping = make(map[string]int)
	m_host_vm_stopped = make(map[string]int)
	m_host_vm_starting = make(map[string]int)
	
	m_zone_vm_running = make(map[string]int)
	m_zone_vm_stopping = make(map[string]int)
	m_zone_vm_stopped = make(map[string]int)
	m_zone_vm_starting = make(map[string]int)
	
	m_host_sys_vm_running = make(map[string]int)
	m_host_sys_vm_stopping = make(map[string]int)
	m_host_sys_vm_stopped = make(map[string]int)
	m_host_sys_vm_starting = make(map[string]int)
	
	m_zone_sys_vm_running = make(map[string]int)
	m_zone_sys_vm_stopping = make(map[string]int)
	m_zone_sys_vm_stopped = make(map[string]int)
	m_zone_sys_vm_starting = make(map[string]int)
	
	
	m_host_user_vm_ifread_curr = make(map[string]int64)
	m_host_user_vm_ifread_last = make(map[string]int64)

	m_host_user_vm_ifwrite_curr = make(map[string]int64)
	m_host_user_vm_ifwrite_last = make(map[string]int64)

	m_host_user_vm_diskread_curr = make(map[string]int64)
	m_host_user_vm_diskread_last = make(map[string]int64)

	m_host_user_vm_diskwrite_curr = make(map[string]int64)
	m_host_user_vm_diskwrite_last = make(map[string]int64)

	m_host_user_vm_diskioread_curr = make(map[string]int64)
	m_host_user_vm_diskioread_last = make(map[string]int64)

	m_host_user_vm_diskiowrite_curr = make(map[string]int64)
	m_host_user_vm_diskiowrite_last = make(map[string]int64)
		
	m_zone_user_vm_ifread_curr = make(map[string]int64)
	m_zone_user_vm_ifread_last = make(map[string]int64)

	m_zone_user_vm_ifwrite_curr = make(map[string]int64)
	m_zone_user_vm_ifwrite_last = make(map[string]int64)

	m_zone_user_vm_diskread_curr = make(map[string]int64)
	m_zone_user_vm_diskread_last = make(map[string]int64)

	m_zone_user_vm_diskwrite_curr = make(map[string]int64)
	m_zone_user_vm_diskwrite_last = make(map[string]int64)

	m_zone_user_vm_diskioread_curr = make(map[string]int64)
	m_zone_user_vm_diskioread_last = make(map[string]int64)

	m_zone_user_vm_diskiowrite_curr = make(map[string]int64)
	m_zone_user_vm_diskiowrite_last = make(map[string]int64)
	
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	inited = 0
	go func() {
		for t := range ticker.C {
			if debug == 1 {
				fmt.Println("DEBUG", time.Now(), " - ", t)
			}
			
			curr_time = time.Now().Unix();
			collect_zone_capacity(client)
			collect_host_status(client)
			collect_user_vm_number(client)
			collect_sys_vm_number(client)
			inited = 1
		}
	}()
	// run for a year - as collectd will restart it
	time.Sleep(time.Second * 86400 * 365 * 100)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}
