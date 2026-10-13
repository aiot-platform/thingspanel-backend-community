package downlink

import (
	"project/internal/devicelog"
)

// recordDeviceLog 记录一条下行设备日志（记录标准化后的 JSON 数据，而非脚本编码后的字节）。
// tenant_id 由 devicelog 写库时按 device_id 补齐。
func (h *Handler) recordDeviceLog(msg *Message, status, errMsg string) {
	if msg == nil || msg.DeviceID == "" {
		return
	}

	logType := devicelog.TypeCommand
	switch msg.Type {
	case MessageTypeTelemetry:
		logType = devicelog.TypeTelemetryControl
	case MessageTypeAttributeSet:
		logType = devicelog.TypeAttributeSet
	case MessageTypeAttributeGet:
		logType = devicelog.TypeAttributeGet
	}

	devicelog.Record(devicelog.Entry{
		DeviceID:  msg.DeviceID,
		MsgType:   logType,
		Direction: devicelog.DirectionDown,
		MessageID: msg.MessageID,
		Status:    status,
		Error:     errMsg,
		Payload:   msg.Data,
	})
}
