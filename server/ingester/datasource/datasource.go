package datasource

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("datasource")

const (
	DATASOURCE_PORT = 20106
)

type DatasourceManager struct {
	ckAddrs        []string // 需要修改数据源的clickhouse地址, 支持多个
	user           string
	password       string
	readTimeout    int
	replicaEnabled bool
	ckdbS3Enabled  bool
	ckdbS3Volume   string
	ckdbS3TTLTimes int

	server *http.Server
}

func NewDatasourceManager(ckAddrs []string, user, password string, readTimeout int, replicaEnabled, ckdbS3Enabled bool, ckdbS3Volume string, ckdbS3TTLTimes int) *DatasourceManager {
	return &DatasourceManager{
		ckAddrs:        ckAddrs,
		user:           user,
		password:       password,
		readTimeout:    readTimeout,
		replicaEnabled: replicaEnabled,
		ckdbS3Enabled:  ckdbS3Enabled,
		ckdbS3Volume:   ckdbS3Volume,
		ckdbS3TTLTimes: ckdbS3TTLTimes,
		server: &http.Server{
			Addr:    ":" + strconv.Itoa(DATASOURCE_PORT),
			Handler: mux.NewRouter(),
		},
	}
}

type JsonResp struct {
	OptStatus   string `json:"OPT_STATUS"`
	Description string `json:"DESCRIPTION,omitempty"`
}

func respSuccess(w http.ResponseWriter) {
	resp, _ := json.Marshal(JsonResp{
		OptStatus: "SUCCESS",
	})
	w.Write(resp)
	log.Info("resp success")
}

func respFailed(w http.ResponseWriter, desc string) {
	resp, _ := json.Marshal(JsonResp{
		OptStatus:   "FAILED",
		Description: desc,
	})
	w.Write(resp)
	log.Warningf("resp failed: %s", desc)
}

type AddBody struct {
	BaseRP       string `json:"base-rp"`
	DB           string `json:"db"`
	Interval     int    `json:"interval"`
	Name         string `json:"name"`
	Duration     int    `json:"retention-time"`
	SummableOP   string `json:"summable-metrics-op"`
	UnsummableOP string `json:"unsummable-metrics-op"`
}

type ModBody struct {
	DB       string `json:"db"`
	Name     string `json:"name"`
	Duration int    `json:"retention-time"`
}

type DelBody struct {
	DB   string `json:"db"`
	Name string `json:"name"`
}

func (m *DatasourceManager) rpAdd(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read body err, %v", err)
		respFailed(w, err.Error())
		return
	}
	var b AddBody
	if err = json.Unmarshal(body, &b); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		respFailed(w, err.Error())
		return
	}
	log.Infof("receive rpadd request: %+v", b)

	err = m.Handle(b.DB, "add", b.BaseRP, b.Name, b.SummableOP, b.UnsummableOP, b.Interval, b.Duration)
	if err != nil {
		respFailed(w, err.Error())
		return
	}
	respSuccess(w)
}

func (m *DatasourceManager) rpMod(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read body err, %v", err)
		respFailed(w, err.Error())
		return
	}
	var b ModBody
	if err = json.Unmarshal(body, &b); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		respFailed(w, err.Error())
		return
	}
	log.Infof("receive rpmod request: %+v", b)

	err = m.Handle(b.DB, "mod", "", b.Name, "", "", 0, b.Duration)
	if err != nil {
		respFailed(w, err.Error())
		return
	}

	respSuccess(w)
}

func (m *DatasourceManager) rpDel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read body err, %v", err)
		respFailed(w, err.Error())
		return
	}
	var b ModBody
	if err = json.Unmarshal(body, &b); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		respFailed(w, err.Error())
		return
	}
	log.Infof("receive rpdel request: %+v", b)

	err = m.Handle(b.DB, "del", "", b.Name, "", "", 0, 0)
	if err != nil {
		respFailed(w, err.Error())
		return
	}
	respSuccess(w)
}

func (m *DatasourceManager) RegisterHandlers() {
	router := m.server.Handler.(*mux.Router)
	router.HandleFunc("/v1/rpadd/", m.rpAdd).Methods("POST")
	router.HandleFunc("/v1/rpmod/", m.rpMod).Methods("PATCH")
	router.HandleFunc("/v1/rpdel/", m.rpDel).Methods("DELETE")
}

func (m *DatasourceManager) Start() {
	m.RegisterHandlers()

	go func() {
		if err := m.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe() failed: %v", err)
		}
	}()
	log.Info("datasource manager started")
}

func (m *DatasourceManager) Close() error {
	if m.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

	if err := m.server.Shutdown(ctx); err != nil {
		log.Errorf("Shutdown() failed: %v", err)
		return err
	}
	cancel()

	log.Info("datasource manager stopped")
	return nil
}
