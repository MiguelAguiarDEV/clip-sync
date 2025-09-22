package types

const MaxInlineBytes = 64 * 1024

type Hello struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type Clip struct {
	MsgID     string `json:"msg_id"`
	Mime      string `json:"mime"`
	Size      int    `json:"size"`
	Data      []byte `json:"data,omitempty"`
	UploadURL string `json:"upload_url,omitempty"`
	From      string `json:"from"`
}

type Envelope struct {
	Type  string `json:"type"`
	Hello *Hello `json:"hello,omitempty"`
	Clip  *Clip  `json:"clip,omitempty"`
}
