package worker

import (
    "context"
    "time"

    "github.com/example/LottoSmash/internal/logger"
    "github.com/example/LottoSmash/internal/txid"
)

type JobType int

const (
    JobTypeExample JobType = iota
)

type Job struct {
    Type   JobType
    TxID   string
    Ctx    context.Context
    Input  interface{}
    Result chan Result
}

type Result struct {
    Data interface{}
    Err  error
}

type Pools struct {
    MainInput chan Job
    DBInput   chan Job
    ExtInput  chan Job
}

func StartMainWorkers(ctx context.Context, count int, pools *Pools, log *logger.Logger) {
    for i := 0; i < count; i++ {
        go func(id int) {
            log.Infof("main worker %d started", id)
            for {
                select {
                case <-ctx.Done():
                    log.Infof("main worker %d stopping", id)
                    return
                case job := <-pools.MainInput:
                    handleMainJob(ctx, log, pools, job)
                }
            }
        }(i)
    }
}

func handleMainJob(ctx context.Context, log *logger.Logger, pools *Pools, job Job) {
    tx := job.TxID
    if tx == "" {
        tx = txid.NewID()
    }
    log.Debugf("handling main job type=%d tx=%s", job.Type, tx)
    // simple example: echo input with small delay
    select {
    case <-ctx.Done():
        return
    case <-time.After(10 * time.Millisecond):
    }

    if job.Result != nil {
        job.Result <- Result{Data: job.Input, Err: nil}
    }
}

func StartDBWorkers(ctx context.Context, count int, pools *Pools, log *logger.Logger) {
    for i := 0; i < count; i++ {
        go func(id int) {
            log.Infof("db worker %d started", id)
            for {
                select {
                case <-ctx.Done():
                    log.Infof("db worker %d stopping", id)
                    return
                case job := <-pools.DBInput:
                    log.Debugf("handling db job type=%d tx=%s", job.Type, job.TxID)
                    if job.Result != nil {
                        job.Result <- Result{Data: job.Input, Err: nil}
                    }
                }
            }
        }(i)
    }
}

func StartExternalWorkers(ctx context.Context, count int, pools *Pools, log *logger.Logger) {
    for i := 0; i < count; i++ {
        go func(id int) {
            log.Infof("external worker %d started", id)
            for {
                select {
                case <-ctx.Done():
                    log.Infof("external worker %d stopping", id)
                    return
                case job := <-pools.ExtInput:
                    log.Debugf("handling external job type=%d tx=%s", job.Type, job.TxID)
                    if job.Result != nil {
                        job.Result <- Result{Data: job.Input, Err: nil}
                    }
                }
            }
        }(i)
    }
}
