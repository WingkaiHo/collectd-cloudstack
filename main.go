// collectd-cloudstack project main.go
package main

import (
	cloudstack "collectd-cloudstack/golang-cloudstack-library"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"bufio"
	"os"
)

var csmgr_host string
var curr_time int64

func get_submit_number_stat_str(host, plugin, plugin_ins, str_type, str_type_ins, value string, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL csmgr_%s/%s-%s/%s-%s %d:%s\n", host, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}

func collect_zone_capacity(c []cloudstack.Capacity) {
	var stat string
	var t int64
	var err error
	var type_total_name, type_used_name, type_used_pect_name string
	f := bufio.NewWriter(os.Stdout)

	for i := range c {
		t, err = c[i].Type.Int64()
		if err == nil || c[i].ZoneName.IsNil() || c[i].CapacityUsed.IsNil() ||
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
		
		stat += get_submit_number_stat_str(csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_total_name, c[i].CapacityTotal.String(), curr_time)
		
		stat += get_submit_number_stat_str(csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_used_name, c[i].CapacityUsed.String(), curr_time)
			
		stat += get_submit_number_stat_str(csmgr_host, "zone", c[i].ZoneName.String(), 
			"guage", type_used_pect_name, c[i].PercentUsed.String(), curr_time)
			
	}
	f.Write([]byte(stat))
	f.Flush()
}

func main() {
	log.SetOutput(ioutil.Discard)

	endpoint, _ := url.Parse("http://172.16.200.22:8080/client/api")
	apikey := "XNqIjXqsNTx5f8Ad55Xfl40nDPnkAgjh-A1C2ITIU4JVSI4kTSEY0GzWYXxkoKxPuzzmph_ZZg3hcesvOzsjNg"
	secretkey := "ilgJ-NyGuDSOblEXK-vb4n9kgjuCHM7t-T3M4hbJT_DxPcky5Dz8ib4zG7gYnNx9QJGgKRfZLdAlU2YZHgZ7OA"
	username := "12345678"
	password := "12345678"

	client, _ := cloudstack.NewClient(endpoint, apikey, secretkey, username, password)

	params := cloudstack.NewListZonesParameter()
	params.Name.Set("VDC2-ZONE1")

	zones, _ := client.ListZones(params)
	b, _ := json.MarshalIndent(zones, "", "    ")

	fmt.Println("Count:", len(zones))
	fmt.Println(string(b))

	//capacity, _ := client
	ListCapacityParam := cloudstack.NewListCapacityParamete()
	capacity, _ := client.ListCapacity(ListCapacityParam)
	//capacity[0].
	c, _ := json.MarshalIndent(capacity, "", "    ")
	fmt.Println("Count:", len(capacity))
	fmt.Println(string(c))
}
