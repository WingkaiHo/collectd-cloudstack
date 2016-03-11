一.概述
1. collectd-cloudstack 是collectd exec插件，可以使用在collectd5.5.0环境下运行
2. 是Go编写的插件可以执行文件可以在不同版本linux环境下运行

二.插件的安装

1. Collectd安装
	cloudstack-manager#yum install collectd
  
2. 拷贝collectd-cloudstack到/usr/lib64/collectd/
  
3. 修改过collectd配置/etc/collectd.conf
    
     cloudstack-manager#vim /etc/collectd.conf

     #Enable follow plugin
     LoadPlugin exec
     LoadPlugin write_graphite
 
    
	<Plugin exec>
      Exec "root" "/usr/lib64/collectd/collectd" "-host=host name of cloudstack manager" "-apikey=The api key of root user" "-secret=secret key of root user" "-port=8080" "-interval=600" 
	</Plugin>

	<Plugin write_graphite>
    <Node "graphing">
        #The ip of graphite-web
        Host "192.168.0.1"
        Port "2003"
        Protocol "tcp"
        LogSendErrors true
        Prefix "collectd."
        Postfix ""
        StoreRates true
        AlwaysAppendDS false
        EscapeCharacter "_"
    </Node>
   </Plugin>

三. cloudstack资源监控
$cloudstack_mgr_host: cloudstack manager 机器hostname
$zone_name: 你所查询zone名称

1. Zone 资源监控数据库路径
1) CPU资源频率总量，分配虚拟机CPU频率
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.cpu_total [整个zone所有host机器CPU总量,单位是HZ]
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.cpu_used  [整个zone已经分配CPU频率数,单位HZ]	

2) 内存总量，已经分配虚拟机内存数量
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.mem_total [整个zone所有host机器内存总量,单位bytes]
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.mem_total [整个zone已经分配给虚拟机内存总数，单位bytes]

3) 主存储容量，已经分配给虚拟机空间容量
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.primary_storage_total [主存储容量,单位bytes]
	collectd.csmgr-$cloudstack_mgr_host.zone-$zone_name.primary_storage_used [分配的容量，单位bytes]

