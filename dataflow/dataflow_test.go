package dataflow

import (
    "context"
    "fmt"
    "runtime"
    "testing"
    "time"
)

func TestDataFlow(t *testing.T) {
    dataFlow := NewDataFlow(20000, 100000)
    go func() {
        for {
            fmt.Println("NumGoroutine:", runtime.NumGoroutine())
            time.Sleep(time.Second)
            if runtime.NumGoroutine() <= 10 {
                return
            }
        }
    }()
    //go MonitorGoroutines(1*time.Second, &dataFlow.Gc)
    dataFlow.RegisterDataSource(func(ctx context.Context, out chan<- interface{}, ticker *time.Ticker) {
        for i := 0; i < 100; i++ {
            //time.Sleep(50 * time.Millisecond)
            select {
            case <-ctx.Done():
                //fmt.Println("stage0", "close")
                return
            case <-ticker.C:
                //fmt.Println("stage0", i)
                out <- i
            }
        }
    }, 100)
    //
    dataFlow.RegisterDataSource(func(ctx context.Context, out chan<- interface{}, ticker *time.Ticker) {
        for letter := 'a'; letter <= 'z'; letter++ {
            //time.Sleep(50 * time.Millisecond)
            select {
            case <-ctx.Done():
                //fmt.Println("stage0", "close")
                return
            case <-ticker.C:
                //fmt.Printf("letter %c\n", letter)
                //out <- letter
                out <- fmt.Sprintf("%c", letter)
            }
        }
    }, 200)

    dataFlow.RegisterDataProcessor(func(ctx context.Context, sourceChan chan interface{}, outputChan chan interface{}) {
        for dataPack := range sourceChan {
            //app.Log.Info("stage1", dataPack)
            outputChan <- dataPack
        }
        //app.Log.Info("stage1", "end")
    }, 50, 100000)

    dataFlow.RegisterDataProcessor(func(ctx context.Context, sourceChan chan interface{}, outputChan chan interface{}) {
        for dataPack := range sourceChan {
            //app.Log.Info("stage2", dataPack)
            time.Sleep(400 * time.Millisecond)
            outputChan <- dataPack
        }
        //app.Log.Info("stage2", "end")
    }, 100, 100000)

    dataFlow.RegisterDataProcessor(func(ctx context.Context, sourceChan chan interface{}, outputChan chan interface{}) {
        for dataPack := range sourceChan {
            //app.Log.Info("stage3", dataPack)
            time.Sleep(400 * time.Millisecond)
            outputChan <- dataPack
        }
        //app.Log.Info("stage3", "end")
    }, 100, 100000)
    dataFlow.Run()
    dataFlow.Listen()
}
