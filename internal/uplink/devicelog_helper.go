package uplink

// metadataString 从消息元数据中取字符串值（不存在或类型不符时返回空串）
func metadataString(msg *DeviceMessage, key string) string {
	if msg == nil {
		return ""
	}
	v, ok := msg.GetMetadata(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
