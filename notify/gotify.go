package notify

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GotifyStruct struct {
	ServerRootUrl string // 不包含/message部分
	Token         string
	Priority      int
}

func (s GotifyStruct) PushNotify(content, title string) error {
	var reqUrl string
	if !strings.HasSuffix(s.ServerRootUrl, "/message") {
		reqUrl = s.ServerRootUrl + "/message?token=" + s.Token
	} else {
		reqUrl = s.ServerRootUrl + "?token=" + s.Token
	}

	resp, err := http.PostForm(reqUrl,
		url.Values{"message": {content}, "title": {title}})
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应体失败: %w", err)
		}
		_ = resp.Body.Close()
		return fmt.Errorf("意外的状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	}
	return nil
}
