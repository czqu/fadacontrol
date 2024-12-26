package bootstrap

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type ProfilingBootstrap struct {
	ctx       context.Context
	startOnce sync.Once
	stopOnce  sync.Once
	enable    bool
}

func NewProfilingBootstrap(ctx context.Context) *ProfilingBootstrap {
	return &ProfilingBootstrap{ctx: ctx}
}
func (p *ProfilingBootstrap) Start() error {
	p.startOnce.Do(func() {
		_conf := utils.GetValueFromContext(p.ctx, constants.ConfKey, conf.NewDefaultConf())
		p.enable = _conf.EnableProfiling
		if _conf.EnableProfiling == false {
			return
		}
		logger.Info("start profiling")
		profPath := filepath.Join(_conf.GetWorkdir(), "prof")
		err := os.MkdirAll(profPath, os.ModePerm)
		if err != nil {
			logger.Error("create prof dir error: %v", err)
		}
		times := time.Now().Format("20060102150405")
		cpuPath := filepath.Join(profPath, "cpu"+"_"+times+".profile")
		fc, err := os.OpenFile(cpuPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.Error("open cpu profile error: %v", err)

		}
		defer fc.Close()
		err = pprof.StartCPUProfile(fc)
		if err != nil {
			logger.Error("start cpu profile error: %v", err)
		}
		runtime.SetBlockProfileRate(1)
		runtime.SetCPUProfileRate(200)
		defer pprof.StopCPUProfile()
		memPath := filepath.Join(profPath, "mem"+"_"+times+".profile")
		fm, err := os.OpenFile(memPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.Error("open mem profile error: %v", err)
		}
		pprof.WriteHeapProfile(fm)
		defer fm.Close()
		select {
		case <-p.ctx.Done():
			logger.Info("stop profiling")
			runtime.GC()
			return
		}

	})

	return nil
}
func (p *ProfilingBootstrap) Stop() error {

	return nil
}
