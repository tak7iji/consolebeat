package beater

import (
    "fmt"
    "bufio"
    "os"
    "time"

    "github.com/elastic/beats/libbeat/beat"
    "github.com/elastic/beats/libbeat/common"
    "github.com/elastic/beats/libbeat/logp"
    "github.com/elastic/beats/libbeat/publisher"

    "github.com/tak7iji/consolebeat/config"
)

type Consolebeat struct {
    done       chan struct{}
    config     config.Config
    client     publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
    config := config.DefaultConfig
    if err := cfg.Unpack(&config); err != nil {
        return nil, fmt.Errorf("Error reading config file: %v", err)
    }

    bt := &Consolebeat{
        done: make(chan struct{}),
        config: config,
    }
    return bt, nil
}

func (bt *Consolebeat) Run(b *beat.Beat) error {
    logp.Info("consolebeat is running! Hit CTRL-C to stop it.")

    bt.client = b.Publisher.Connect()
    ticker := time.NewTicker(bt.config.Period)
    ch := make(chan string)

    go func(ch chan string) {
        scanner := bufio.NewScanner(os.Stdin)
        for {
            for scanner.Scan() {
                if line := scanner.Text(); !bt.isSkipEmptyLine(line) {
                    if !bt.isExcludeLines(line) {
                        ch <- line
                    }
                }
            }
            if scanner.Err() != nil {
                close(ch)
                return
            }
        }
    }(ch)

    for {
        select {
        case <-bt.done:
            return nil
        case text, ok := <-ch:
            if !ok {
                return nil
            }
//            event := common.MapStr{
//                "@timestamp": common.Time(time.Now()),
//                "type":       b.Name,
//                "messaage":   text,
//            }
            event := common.MapStr
            event.Put("@timestamp", common.Time(time.Now()))
            event.Put("type", b.Name)
            event.Put("message", text)
            bt.client.PublishEvent(event)
            logp.Info("Event sent")
        case <-ticker.C:
        }
    }
}

func (bt *Consolebeat) Stop() {
    bt.client.Close()
    close(bt.done)
}

func (bt *Consolebeat) isSkipEmptyLine(line string) bool {
    return bt.config.SkipEmptyLine && line == ""
}

func (bt *Consolebeat) isExcludeLines(line string) bool {
    if len(bt.config.ExcludeLines) > 0 {
        for _, rexp := range bt.config.ExcludeLines {
            if rexp.MatchString(line) {
                return true
            }
        }
    }

    return false
}
