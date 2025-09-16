package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

func SendReport(url string, path string, body any, timeDuration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeDuration)
	defer cancel()
	data, err := json.Marshal(body)
	if err != nil {
		log.Println("error parsing body")
		return
	}
	log.Println(string(data))
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/"+path, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error creating request ...")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("error fetching api")
		return
	}

	defer resp.Body.Close()

	responseData, _ := io.ReadAll(resp.Body)

	log.Println("Log saved at level "+path+":", string(responseData))
}
