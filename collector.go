package main

import (
	_"strconv"
	"strings"
	"sync"
	"time"
	"fmt"
	"github.com/mojodojo101/backhaul_b5_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/soniah/gosnmp"
)

const prefix = "backhaul_b5_"

var (
	
	upDesc *prometheus.Desc
	/*
		upDesc               *prometheus.Desc

		upTimeDesc				 *prometheus.Desc
		cpuUsageOneDesc			 *prometheus.Desc
		cpuUsageTwoDesc			 *prometheus.Desc
		cpuUsageThreeDesc			 *prometheus.Desc
		cpuUsageFourDesc			 *prometheus.Desc
	*/
	//http://backhaul.help.mimosa.co/snmp-oid-reference

	//General Information
	
	/*
		TempMetric = &Metric{
		Name:	"internal_temp",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.1.8.0",
		Type:	"int",
		Help:	"The internal temperature of the device in Celsius",	
		}
		//1.3.6.1.4.1.43356.2.1.2.1.8.0	mimosaInternalTemp.0	INTEGER: 382 C1	Overview > Dashboard > Device Details > Internal Temp or CPU Temp (Local)
	*/

	temperatureDesc	*prometheus.Desc
	
	
	//Performance Information
	
	/*
		ThroughputTxMetric = &Metric{
		Name:	"throughput_tx",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.7.1.0",
		Type:	"int",
		Help:	"The througput transmit rate in kbps",	
		}
		//1.3.6.1.4.1.43356.2.1.2.7.1.0	mimosaPhyTxRate.0	INTEGER: 94081 kbps2	Overview > Dashboard > Performance > Throughput > Tx
	*/
	throughputTxDesc *prometheus.Desc

	/*
		ThroughputRxMetric = &Metric{
		Name:	"throughput_rx",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.7.2.0",
		Type:	"int",
		Help:	"The througput receive rate in kbps",	
		}
		//1.3.6.1.4.1.43356.2.1.2.7.2.0 	mimosaPhyRxRate.0	INTEGER: 76406 kbps2	Overview > Dashboard > Performance > Throughput > Rx
	*/
	throughputRxDesc *prometheus.Desc
	/*
		PerformanceTxMetric = &Metric{
		Name:	"performance_tx",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.7.3.0",
		Type:	"int",
		Help:	"The performance per transmit rate in %",	
		}
		//1.3.6.1.4.1.43356.2.1.2.7.3.0 	mimosaPerTxRate.0	INTEGER: 0.27 %2	Overview > Dashboard > Performance > PER > Tx
	*/
	performanceTxDesc *prometheus.Desc

	/*
	PerformanceRxMetric = &Metric{
		Name:	"performance_rx",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.7.4.0"	,
		Type:	"int",
		Help:	"The performance per receive rate in %",	
		}
		//1.3.6.1.4.1.43356.2.1.2.7.4.0	mimosaPerRxRate.0	INTEGER: 0.73 %2	Overview > Dashboard > Performance > PER > Rx
	*/
	performanceRxDesc *prometheus.Desc

	//Wan information
	/*
		WanUptimeMetric = &Metric{
		Name:	"wan_uptime",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.3.4.0"	,
		Type:	"timeticks",
		Help:	"The Link Uptime of the device",	
		}
		//1.3.6.1.4.1.43356.2.1.2.3.4.0	mimosaWanUpTime.0	Timeticks: (18571300) 2 days, 3:35:13.00	Overview > Dashboard > Link Uptime
	*/
	wanUptimeDesc *prometheus.Desc


	/*
	WanStatusMetric = &Metric{
		Name:	"wan_status",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.3.3.0"	,
		Type:	"integer",
		Help:	"The Wireless Status",	
		}
		//1.3.6.1.4.1.43356.2.1.2.3.3.0	mimosaWanStatus.0	INTEGER: connected(1)	Overview > Dashboard > Wireless Status
	*/
	wanStatusDesc *prometheus.Desc

		
	/*
		SignalStrengthMetric = &Metric{
		Name:	"signal_strength",
		Oid:	"1.3.6.1.4.1.43356.2.1.2.6.7.0"	,
		Type:	"integer",
		Help:	"The target signal strength in dBm",	
		}
		//1.3.6.1.4.1.43356.2.1.2.6.7.0	mimosaTargetSignalStrength.0	INTEGER: -500 dBm1	Overview > Dashboard > Signal Meter > Target
	*/
	signalStrengthDesc *prometheus.Desc
)

func init() {
	l := []string{"target"}
	
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", l, nil)
	temperatureDesc		=prometheus.NewDesc(prefix+"internal_temp","The internal temperature of the device in Celsius",l,nil)
	throughputTxDesc 	=prometheus.NewDesc(prefix+"throughput_tx","The througput transmit rate in kbps",l,nil)
	throughputRxDesc 	=prometheus.NewDesc(prefix+"throughput_rx","The througput receive rate in kbps",l,nil)
	performanceTxDesc 	=prometheus.NewDesc(prefix+"performance_tx","The performance per transmit rate in %",l,nil)
	performanceRxDesc 	=prometheus.NewDesc(prefix+"performance_rx","The performance per receive rate in %",l,nil)
	wanUptimeDesc 		=prometheus.NewDesc(prefix+"wan_uptime","The Link Uptime of the device",l,nil)
	wanStatusDesc 		=prometheus.NewDesc(prefix+"wan_status","The Wireless Status",l,nil)
	signalStrengthDesc  =prometheus.NewDesc(prefix+"signal_strength","The target signal strength in dBm",l,nil)
}

type backhaulB5Collector struct {
	cfg *config.Config
}

func newBackhaulB5Collector(cfg *config.Config) *backhaulB5Collector {
	return &backhaulB5Collector{cfg}
}

func (c backhaulB5Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc

	ch <- temperatureDesc		
	ch <- throughputTxDesc 		
	ch <- throughputRxDesc 	
	ch <- performanceTxDesc 		
	ch <- performanceRxDesc 	
	ch <- wanUptimeDesc 			
	ch <- wanStatusDesc 		
	ch <- signalStrengthDesc  	


	
}
func (c backhaulB5Collector) collectTarget(target string, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	snmp := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: *snmpCommunity,
		Version:   gosnmp.Version1,
		Timeout:   time.Duration(2) * time.Second,
	}
	err := snmp.Connect()
	if err != nil {
		log.Infof("Connect() err: %v\n", err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, target)
		return
	}
	defer snmp.Conn.Close()

	oids := []string{"1.3.6.1.4.1.43356.2.1.2.1.8.0","1.3.6.1.4.1.43356.2.1.2.7.1.0","1.3.6.1.4.1.43356.2.1.2.7.2.0"}
	oids = append(oids,"1.3.6.1.4.1.43356.2.1.2.7.3.0","1.3.6.1.4.1.43356.2.1.2.7.4.0","1.3.6.1.4.1.43356.2.1.2.3.4.0","1.3.6.1.4.1.43356.2.1.2.3.3.0","1.3.6.1.4.1.43356.2.1.2.6.7.0") 
	result, err2 := snmp.Get(oids)
	if err2 != nil {
		log.Infof("Get() err: %v from %s\n", err2, target)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, target)
		return
	}

	for _, variable := range result.Variables {
		if variable.Value == nil {
			continue
		}
		
		switch variable.Name[1:] {
		case oids[0]:
			ch <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[1]:
			ch <- prometheus.MustNewConstMetric(throughputTxDesc, prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[2]:
			ch <- prometheus.MustNewConstMetric(throughputRxDesc, prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[3]:
			ch <- prometheus.MustNewConstMetric(performanceTxDesc , prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[4]:
			ch <- prometheus.MustNewConstMetric(performanceRxDesc , prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[5]:
			ch <- prometheus.MustNewConstMetric(wanUptimeDesc , prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[6]:
			ch <- prometheus.MustNewConstMetric(wanStatusDesc, prometheus.GaugeValue, float64(variable.Value.(int)), target)
		case oids[7]:
			ch <- prometheus.MustNewConstMetric(signalStrengthDesc , prometheus.GaugeValue, float64(variable.Value.(int)), target)
	}
}

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, target)
}

func (c backhaulB5Collector) Collect(ch chan<- prometheus.Metric) {
	targets := strings.Split(*snmpTargets, ",")
	targets = append(targets, c.cfg.Targets...)
	wg := &sync.WaitGroup{}
	fmt.Printf("targets:%v\n",targets)
	for _, target := range targets {
		if target == "" {
			continue
		}
		wg.Add(1)
		go c.collectTarget(target, ch, wg)
	}

	wg.Wait()
}