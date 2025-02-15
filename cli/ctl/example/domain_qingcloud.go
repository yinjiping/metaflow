package example

var YamlDomainQingCloud = []byte(`
# 名称
NAME: qingcloud
# 云平台类型
TYPE: 14
CONFIG:
  # 所属区域标识
  region_uuid: ffffffff-ffff-ffff-ffff-ffffffffffff
  # 资源同步控制器
  controller_ip: 127.0.0.1
  # API 密钥 ID
  # 在青云主页面右上角-API密钥-API密钥管理-API密钥ID
  secret_id: xxxxxxxx
  # API 密钥 KEY
  # API密钥ID对应的API密钥KEY
  secret_key: xxxxxxx
`)
