package types

const MaxInlineBytes = 64 << 10 // 64 KiB

type Envelope struct {
	Type  string `json:"type"`
	From  string `json:"from,omitempty"` // deviceID del emisor
	Hello *Hello `json:"hello,omitempty"`
	Clip  *Clip  `json:"clip,omitempty"`
}

type Hello struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type Clip struct {
	MsgID     string `json:"msg_id,omitempty"`
	Mime      string `json:"mime,omitempty"`
	Size      int    `json:"size,omitempty"`
	Data      []byte `json:"data,omitempty"`
	UploadURL string `json:"upload_url,omitempty"`
}
