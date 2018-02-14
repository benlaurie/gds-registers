package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Register struct {
	baseURL string
	info    map[string]interface{}
}

type Info struct {
	Records     uint64
	Entries     uint64
	LastUpdated time.Time
	Text        string
}

func NewRegister(name string) (*Register, error) {
	r := &Register{baseURL: "https://" + name + ".register.gov.uk/"}
	err := r.getInfo()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Register) getJSON(url string, j interface{}) error {
	req, err := http.NewRequest("GET", r.baseURL+url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Status for %s not 200: %s", url, resp.Status)
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	return d.Decode(j)
}

func (r *Register) getInfo() error {
	return r.getJSON("register", &r.info)
}

func jsonUint64(i interface{}) (uint64, error) {
	var vi uint64
	switch v := i.(type) {
	case float64:
		vi = uint64(v)
		if vi > 1<<53 {
			return 0, errors.New("JSON was a bad choice")
		}

	case string:
		t, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		vi = t
	}
	return vi, nil
}

func (r *Register) Info() (*Info, error) {
	rf, err := jsonUint64(r.info["total-records"])
	if err != nil {
		return nil, err
	}
	re, err := jsonUint64(r.info["total-entries"])
	if err != nil {
		return nil, err
	}
	l, err := time.Parse(time.RFC3339, r.info["last-updated"].(string))
	if err != nil {
		return nil, err
	}
	t, _ := r.info["register-record"].(map[string]interface{})["text"].(string)
	return &Info{Records: rf, Entries: re, LastUpdated: l, Text: t}, nil
}

type Processor interface {
	Process(map[string]interface{}) error
}

func (r *Register) GetSummaryEntries(p Processor) error {
	i, err := r.Info()
	if err != nil {
		return err
	}
	for n := uint64(1); n < i.Entries; {
		var j []map[string]interface{}
		err = r.getJSON(fmt.Sprintf("entries?start=%d", n), &j)
		if err != nil {
			return err
		}
		for _, e := range j {
			ien, err := jsonUint64(e["index-entry-number"])
			if err != nil {
				return err
			}
			in, err := jsonUint64(e["entry-number"])
			if err != nil {
				return err
			}
			if ien != n || in != n {
				return fmt.Errorf("Expecting entry %d: %#v", n, e)
			}
			err = p.Process(e)
			if err != nil {
				return err
			}
			n++
		}
	}
	return nil
}

type ItemProcessor interface {
	Process(entry map[string]interface{}, hash string, item map[string]interface{}) error
}

type entryProcessor struct {
	r *Register
	p ItemProcessor
}

func (p *entryProcessor) Process(e map[string]interface{}) error {
	for _, h := range e["item-hash"].([]interface{}) {
		var j map[string]interface{}
		err := p.r.getJSON(fmt.Sprintf("item/"+h.(string)), &j)
		if err != nil {
			return err
		}
		err = p.p.Process(e, h.(string), j)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Register) GetEntries(p ItemProcessor) error {
	return r.GetSummaryEntries(&entryProcessor{r: r, p: p})
}
