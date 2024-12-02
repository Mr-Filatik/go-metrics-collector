package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Metric struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	PollCount     int64
	RandomValue   float64
}

func main() {

	endpointEnv := os.Getenv("ADDRESS")
	var endpoint *string
	if endpointEnv == "" {
		endpoint = flag.String("a", "localhost:8080", "HTTP server endpoint")
	}
	reportIntervalEnv := os.Getenv("REPORT_INTERVAL")
	rnum, rerr := strconv.ParseInt(reportIntervalEnv, 10, 64)
	var reportInterval *int64
	if reportIntervalEnv == "" || rerr != nil {
		reportInterval = flag.Int64("r", 10, "Report interval")
	} else {
		reportInterval = &rnum
	}
	pollIntervalEnv := os.Getenv("POLL_INTERVAL")
	pnum, perr := strconv.ParseInt(pollIntervalEnv, 10, 64)
	var pollInterval *int64
	if pollIntervalEnv == "" || perr != nil {
		pollInterval = flag.Int64("p", 2, "Poll interval")
	} else {
		pollInterval = &pnum
	}
	flag.Parse()

	metr := Metric{PollCount: 0}
	go Refresh(&metr, *pollInterval)
	Send(&metr, "http://"+*endpoint, *reportInterval)
}

func Refresh(m *Metric, pollInterval int64) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		RefreshMetrics(m)
	}
}

func SendRequest(c *resty.Client, endpoint string, metricType string, metricName string, metricValue string) error {

	address := endpoint + "/update/" + metricType + "/" + metricName + "/" + metricValue

	log.Printf("Response to %v.", address)

	resp, err := c.R().
		SetHeader("Content-Type", " text/plain").
		Post(address)

	if err != nil {
		log.Printf("Error on response: %v.", err.Error())
		return err
	}
	log.Printf("Response is done. StatusCode: %v.", resp.Status())
	return nil
}

func Send(m *Metric, endpoint string, reportInterval int64) {
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)

		client := resty.New()
		SendRequest(client, endpoint, "gauge", "Alloc", strconv.FormatFloat(m.Alloc, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "BuckHashSys", strconv.FormatFloat(m.BuckHashSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "Frees", strconv.FormatFloat(m.Frees, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "GCCPUFraction", strconv.FormatFloat(m.GCCPUFraction, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "GCSys", strconv.FormatFloat(m.GCSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapAlloc", strconv.FormatFloat(m.HeapAlloc, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapIdle", strconv.FormatFloat(m.HeapIdle, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapInuse", strconv.FormatFloat(m.HeapInuse, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapObjects", strconv.FormatFloat(m.HeapObjects, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapReleased", strconv.FormatFloat(m.HeapReleased, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "HeapSys", strconv.FormatFloat(m.HeapSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "LastGC", strconv.FormatFloat(m.LastGC, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "Lookups", strconv.FormatFloat(m.Lookups, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "MCacheInuse", strconv.FormatFloat(m.MCacheInuse, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "MCacheSys", strconv.FormatFloat(m.MCacheSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "MSpanInuse", strconv.FormatFloat(m.MSpanInuse, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "MSpanSys", strconv.FormatFloat(m.MSpanSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "Mallocs", strconv.FormatFloat(m.Mallocs, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "NextGC", strconv.FormatFloat(m.NextGC, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "NumForcedGC", strconv.FormatFloat(m.NumForcedGC, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "NumGC", strconv.FormatFloat(m.NumGC, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "OtherSys", strconv.FormatFloat(m.OtherSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "PauseTotalNs", strconv.FormatFloat(m.PauseTotalNs, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "StackInuse", strconv.FormatFloat(m.StackInuse, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "StackSys", strconv.FormatFloat(m.StackSys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "Sys", strconv.FormatFloat(m.Sys, 'f', -1, 64))
		SendRequest(client, endpoint, "gauge", "TotalAlloc", strconv.FormatFloat(m.TotalAlloc, 'f', -1, 64))
		SendRequest(client, endpoint, "counter", "PollCount", strconv.FormatInt(m.PollCount, 10))
		m.PollCount = 0
		SendRequest(client, endpoint, "gauge", "RandomValue", strconv.FormatFloat(m.RandomValue, 'f', -1, 64))

		log.Printf("Send metrics (Errors count: %v).", 0) //hardcode =(

		//log.Printf("Alloc: %v.", m.Alloc)
		//log.Printf("BuckHashSys: %v.", m.BuckHashSys)
		//log.Printf("Frees: %v.", m.Frees)
		//log.Printf("GCCPUFraction: %v.", m.GCCPUFraction)
		//log.Printf("GCSys: %v.", m.GCSys)
		//log.Printf("HeapAlloc: %v.", m.HeapAlloc)
		//log.Printf("HeapIdle: %v.", m.HeapIdle)
		//log.Printf("HeapInuse: %v.", m.HeapInuse)
		//log.Printf("HeapObjects: %v.", m.HeapObjects)
		//log.Printf("HeapReleased: %v.", m.HeapReleased)
		//log.Printf("HeapSys: %v.", m.HeapSys)
		//log.Printf("LastGC: %v.", m.LastGC)
		//log.Printf("Lookups: %v.", m.Lookups)
		//log.Printf("MCacheInuse: %v.", m.MCacheInuse)
		//log.Printf("MCacheSys: %v.", m.MCacheSys)
		//log.Printf("MSpanInuse: %v.", m.MSpanInuse)
		//log.Printf("MSpanSys: %v.", m.MSpanSys)
		//log.Printf("Mallocs: %v.", m.Mallocs)
		//log.Printf("NextGC: %v.", m.NextGC)
		//log.Printf("NumForcedGC: %v.", m.NumForcedGC)
		//log.Printf("NumGC: %v.", m.NumGC)
		//log.Printf("OtherSys: %v.", m.OtherSys)
		//log.Printf("PauseTotalNs: %v.", m.PauseTotalNs)
		//log.Printf("StackInuse: %v.", m.StackInuse)
		//log.Printf("StackSys: %v.", m.StackSys)
		//log.Printf("Sys: %v.", m.Sys)
		//log.Printf("TotalAlloc: %v.", m.TotalAlloc)
		//log.Printf("PollCount: %v.", m.PollCount) //counter
		//log.Printf("RandomValue: %v.", m.RandomValue)
	}
}

func RefreshMetrics(m *Metric) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.Alloc = float64(mem.Alloc)
	m.BuckHashSys = float64(mem.BuckHashSys)
	m.Frees = float64(mem.Frees)
	m.GCCPUFraction = mem.GCCPUFraction
	m.GCSys = float64(mem.GCSys)
	m.HeapAlloc = float64(mem.HeapAlloc)
	m.HeapIdle = float64(mem.HeapIdle)
	m.HeapInuse = float64(mem.HeapInuse)
	m.HeapObjects = float64(mem.HeapObjects)
	m.HeapReleased = float64(mem.HeapReleased)
	m.HeapSys = float64(mem.HeapSys)
	m.LastGC = float64(mem.LastGC)
	m.Lookups = float64(mem.Lookups)
	m.MCacheInuse = float64(mem.MCacheInuse)
	m.MCacheSys = float64(mem.MCacheSys)
	m.MSpanInuse = float64(mem.MSpanInuse)
	m.MSpanSys = float64(mem.MSpanSys)
	m.Mallocs = float64(mem.Mallocs)
	m.NextGC = float64(mem.NextGC)
	m.NumForcedGC = float64(mem.NumForcedGC)
	m.NumGC = float64(mem.NumGC)
	m.OtherSys = float64(mem.OtherSys)
	m.PauseTotalNs = float64(mem.PauseTotalNs)
	m.StackInuse = float64(mem.StackInuse)
	m.StackSys = float64(mem.MSpanSys)
	m.Sys = float64(mem.Sys)
	m.TotalAlloc = float64(mem.TotalAlloc)

	m.PollCount++
	m.RandomValue = rand.Float64()

	log.Printf("Update metrics.")
}
