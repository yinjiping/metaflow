package metrics

import (
	"errors"
	"fmt"

	ckcommon "server/querier/engine/clickhouse/common"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("clickhouse.metrics")

const METRICS_OPERATOR_GTE = ">="
const METRICS_OPERATOR_LTE = "<="

var METRICS_OPERATORS = []string{METRICS_OPERATOR_GTE, METRICS_OPERATOR_LTE}

type Metrics struct {
	Index       int    // 索引
	DBField     string // 数据库字段
	DisplayName string // 描述
	Unit        string // 单位
	Type        int    // 指标量类型
	Category    string // 类别
	Condition   string // 聚合过滤
	IsAgg       bool   // 是否为聚合指标量
	Permissions []bool // 指标量的权限控制
}

func (m *Metrics) Replace(metrics *Metrics) {
	m.IsAgg = metrics.IsAgg
	if metrics.DBField != "" {
		m.DBField = metrics.DBField
	}
	if metrics.Condition != "" {
		m.Condition = metrics.Condition
	}
}

func (m *Metrics) SetIsAgg(isAgg bool) *Metrics {
	m.IsAgg = isAgg
	return m
}

func NewMetrics(
	index int, dbField string, displayname string, unit string, metricType int, category string,
	permissions []bool, condition string,
) *Metrics {
	return &Metrics{
		Index:       index,
		DBField:     dbField,
		DisplayName: displayname,
		Unit:        unit,
		Type:        metricType,
		Category:    category,
		Permissions: permissions,
		Condition:   condition,
	}
}

func NewReplaceMetrics(dbField string, condition string) *Metrics {
	return &Metrics{
		DBField:   dbField,
		Condition: condition,
		IsAgg:     true,
	}
}

func GetMetrics(field string, db string, table string) (*Metrics, bool) {
	allMetrics, err := GetMetricsByDBTable(db, table)
	if err != nil {
		return nil, false
	}
	metric, ok := allMetrics[field]
	return metric, ok
}

func GetMetricsByDBTable(db string, table string) (map[string]*Metrics, error) {
	var err error
	switch db {
	case "flow_log":
		switch table {
		case "l4_flow_log":
			return GetL4FlowLogMetrics(), err
		case "l7_flow_log":
			return GetL7FlowLogMetrics(), err
		}
	case "flow_metrics":
		switch table {
		case "vtap_flow_port":
			return GetVtapFlowPortMetrics(), err
		case "vtap_flow_edge_port":
			return GetVtapFlowEdgePortMetrics(), err
		case "vtap_app_port":
			return GetVtapAppPortMetrics(), err
		case "vtap_app_edge_port":
			return GetVtapAppEdgePortMetrics(), err
		}
	case "ext_metrics":
		return GetExtMetrics(db, table)
	}
	return nil, err
}

func GetMetricsDescriptions(db string, table string) (map[string][]interface{}, error) {
	allMetrics, err := GetMetricsByDBTable(db, table)
	if allMetrics == nil || err != nil {
		// TODO: metrics not found
		return nil, err
	}
	columns := []interface{}{
		"name", "is_agg", "display_name", "unit", "type", "category", "operators", "permissions",
	}
	values := make([]interface{}, len(allMetrics))
	for field, metrics := range allMetrics {
		values[metrics.Index] = []interface{}{
			field, metrics.IsAgg, metrics.DisplayName, metrics.Unit, metrics.Type,
			metrics.Category, METRICS_OPERATORS, metrics.Permissions,
		}
	}
	return map[string][]interface{}{
		"columns": columns,
		"values":  values,
	}, nil
}

func LoadMetrics(db string, table string, dbDescription map[string]interface{}) (loadMetrics map[string]*Metrics, err error) {
	tableDate, ok := dbDescription[db]
	if !ok {
		return nil, errors.New(fmt.Sprintf("get metrics failed! db: %s", db))
	}
	if ok {
		metricsData, ok := tableDate.(map[string]interface{})[table]
		if ok {
			loadMetrics = make(map[string]*Metrics)
			for i, metrics := range metricsData.([][]interface{}) {
				if len(metrics) < 7 {
					return nil, errors.New(fmt.Sprintf("get metrics failed! db:%s table:%s metrics:%v", db, table, metrics))
				}
				metricType, ok := METRICS_TYPE_NAME_MAP[metrics[4].(string)]
				if !ok {
					return nil, errors.New(fmt.Sprintf("get metrics type failed! db:%s table:%s metrics:%v", db, table, metrics))
				}
				permissions, err := ckcommon.ParsePermission(metrics[6])
				if err != nil {
					return nil, errors.New(fmt.Sprintf("parse metrics permission failed! db:%s table:%s metrics:%v", db, table, metrics))
				}
				lm := NewMetrics(
					i, metrics[1].(string), metrics[2].(string), metrics[3].(string), metricType,
					metrics[5].(string), permissions, "",
				)
				loadMetrics[metrics[0].(string)] = lm
			}
		} else {
			return nil, errors.New(fmt.Sprintf("get metrics failed! db:%s table:%s", db, table))
		}
	}
	return loadMetrics, nil
}

func MergeMetrics(db string, table string, loadMetrics map[string]*Metrics) error {
	var metrics map[string]*Metrics
	var replaceMetrics map[string]*Metrics
	switch db {
	case "flow_log":
		switch table {
		case "l4_flow_log":
			metrics = L4_FLOW_LOG_METRICS
			replaceMetrics = L4_FLOW_LOG_METRICS_REPLACE
		case "l7_flow_log":
			metrics = L7_FLOW_LOG_METRICS
			replaceMetrics = L7_FLOW_LOG_METRICS_REPLACE
		}
	case "flow_metrics":
		switch table {
		case "vtap_flow_port":
			metrics = VTAP_FLOW_PORT_METRICS
			replaceMetrics = VTAP_FLOW_PORT_METRICS_REPLACE
		case "vtap_flow_edge_port":
			metrics = VTAP_FLOW_EDGE_PORT_METRICS
			replaceMetrics = VTAP_FLOW_EDGE_PORT_METRICS_REPLACE
		case "vtap_app_port":
			metrics = VTAP_APP_PORT_METRICS
			replaceMetrics = VTAP_APP_PORT_METRICS_REPLACE
		case "vtap_app_edge_port":
			metrics = VTAP_APP_EDGE_PORT_METRICS
			replaceMetrics = VTAP_APP_EDGE_PORT_METRICS_REPLACE
		}
	case "ext_metrics":
		metrics = EXT_METRICS
	}
	if metrics == nil {
		return errors.New(fmt.Sprintf("merge metrics failed! db:%s, table:%s", db, table))
	}
	for name, value := range loadMetrics {
		// TAG类型指标量都属于聚合类型
		if value.Type == METRICS_TYPE_TAG {
			value.IsAgg = true
		}
		if rm, ok := replaceMetrics[name]; ok && value.DBField == "" {
			value.Replace(rm)
		}
		metrics[name] = value
	}
	return nil
}
