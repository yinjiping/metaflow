package tag

var L7TagMap = GenerateL7TagMap()

func GetL7Tag(name string) (*Tag, bool) {
	tag, ok := L7TagMap[name]
	return tag, ok
}

func GenerateL7TagMap() map[string]*Tag {
	l7TagMap := make(map[string]*Tag)

	// 响应码
	l7TagMap["response_code"] = NewTag(
		"",
		"isNotNull(response_code)",
		"",
		"",
	)
	// 采集点ID
	l7TagMap["tap_type_value_id"] = NewTag(
		"tap_type",
		"",
		"tap_type %s %s",
		"",
	)
	// 采集点
	l7TagMap["tap_type_value"] = NewTag(
		"dictGet(deepflow.tap_type_map, ('name'), toUInt64(tap_type))",
		"",
		"toUInt64(tap_type) IN (SELECT value FROM deepflow.tap_type_map WHERE name %s %s)",
		"toUInt64(tap_type) IN (SELECT value FROM deepflow.tap_type_map WHERE %s(name,%s))",
	)
	// IP类型
	l7TagMap["ip_version"] = NewTag(
		"if(is_ipv4=1, 4, 6)",
		"",
		"is_ipv4 %s %s",
		"",
	)
	// 是否匹配服务
	l7TagMap["include_service"] = NewTag(
		"",
		"",
		"is_key_service %s %s",
		"",
	)
	// ID
	l7TagMap["_id"] = NewTag(
		"",
		"",
		"_id %s %s AND time=bitShiftRight(%s, 32) AND toStartOfHour(time)=toStartOfHourbitShiftRight(%s, 32)",
		"",
	)
	return l7TagMap
}
