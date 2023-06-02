package service

import "net/http"

func TestThirdAPI() {
	req, _ := http.NewRequest("GET", "url", nil)
	req.Header.Set("", "")
	http.DefaultClient.Do(req)
}

func TestMulti() {
	for i := 1; i <= 1000; i++ {
		go TestThirdAPI()
	}
}
