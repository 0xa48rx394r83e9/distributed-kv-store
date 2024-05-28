// client/client.go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type KVClient struct {
    serverURL string
}

func NewKVClient(serverURL string) *KVClient {
    return &KVClient{serverURL: serverURL}
}

func (c *KVClient) Get(key string) (string, error) {
    resp, err := http.Get(fmt.Sprintf("%s/get?key=%s", c.serverURL, key))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("server returned status code %d", resp.StatusCode)
    }

    var reply GetReply
    err = json.NewDecoder(resp.Body).Decode(&reply)
    if err != nil {
        return "", err
    }

    return reply.Value, nil
}

func (c *KVClient) Set(key, value string) error {
    args := SetArgs{Key: key, Value: value}
    data, err := json.Marshal(args)
    if err != nil {
        return err
    }

    resp, err := http.Post(fmt.Sprintf("%s/set", c.serverURL), "application/json", bytes.NewBuffer(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server returned status code %d", resp.StatusCode)
    }

    return nil
}

func (c *KVClient) Delete(key string) error {
    req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/delete?key=%s", c.serverURL, key), nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server returned status code %d", resp.StatusCode)
    }

    return nil
}

type GetReply struct {
    Value string `json:"value"`
}

type SetArgs struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

