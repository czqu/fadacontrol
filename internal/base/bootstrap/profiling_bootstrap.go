package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type ProfilingBootstrap struct {
	conf      *conf.Conf
	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

func NewProfilingBootstrap(conf *conf.Conf) *ProfilingBootstrap {
	return &ProfilingBootstrap{conf: conf, done: make(chan struct{})}
}
func (p *ProfilingBootstrap) Start() error {
	p.startOnce.Do(func() {
		if p.conf.EnableProfiling == false {
			return
		}
		logger.Info("start profiling")
		profPath := filepath.Join(p.conf.GetWorkdir(), "prof")
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
		<-p.done
		runtime.GC()
	})

	return nil
}
func (p *ProfilingBootstrap) Stop() error {
	p.stopOnce.Do(func() {
		if p.conf.EnableProfiling == false {
			return
		}
		p.done <- struct{}{}
		logger.Info("stop profiling")
	})

	return nil
}
